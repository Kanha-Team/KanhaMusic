/*
 * ● KanhaMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 Kanha
 *
 * This program is free software: you can redistribute it and/or modify it under the
 * terms of the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
 * PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * Repository: https://github.com/Kanha-Team/KanhaMusic
 */

package modules

import (
	"fmt"
	"strings"
	"time"

	"KanhaMusic/kanha/logger"

	td "github.com/Kanha/Meow"

	"KanhaMusic/config"
	"KanhaMusic/kanha/core"
	"KanhaMusic/kanha/database"
	"KanhaMusic/kanha/locales"
	"KanhaMusic/kanha/utils"
)

func actionFilter(m *td.Message) bool {
	switch m.Content.(type) {
	case *td.MessageChatAddMembers,
		*td.MessageVideoChatStarted,
		*td.MessageVideoChatEnded:
		return true
	}
	return false
}

func handleActions(c *td.Client, m *td.Message) error {
	chat, err := m.GetChat(c)
	if err != nil {
		return nil
	}
	switch t := chat.Type.(type) {
	case *td.ChatTypeSupergroup:
		if t.IsChannel {
			return nil
		}
	case *td.ChatTypePrivate, *td.ChatTypeSecret:
		return nil
	default:
		text := F(m.ChatID(), "supergroup_needed", locales.Arg{
			"chat_id":       m.ChatID(),
			"support_group": config.SupportChat,
		})
		if _, err := c.SendTextMessage(m.ChatID(), text, &td.SendTextMessageOpts{DisableWebPagePreview: true}); err != nil {
			logger.Errorf("failed to send supergroup conversion message to chat %d: %v", m.ChatID(), err)
			return nil
		}
		go func() {
			leaveChat(c, m.ChatID())
		}()
		return nil
	}

	switch m.Content.(type) {
	case *td.MessageChatAddMembers:
		return handleChatMemberAdd(c, m)
	case *td.MessageVideoChatStarted, *td.MessageVideoChatEnded:
		return handleVoiceChatAction(c, m)
	}

	return nil
}

func handleChatMemberAdd(c *td.Client, m *td.Message) error {
	chatID := m.ChatID()

	if c.Me == nil {
		return nil
	}
	content := m.Content.(*td.MessageChatAddMembers)

	for _, uid := range content.MemberUserIds {
		if uid != c.Me.Id {
			continue
		}

		if blockedChat, _ := database.IsBlacklistedChat(chatID); blockedChat && !isOwnerOrSudo(m.SenderID()) {
			if _, rerr := m.ReplyText(c, F(chatID, "blacklist_chat_blocked"), nil); rerr != nil {
				logger.Error(rerr)
			}
			leaveChat(c, chatID)
			return nil
		}

		ownerID, err := utils.GetChatOwner(c, chatID)
		if err == nil {
			if blockedOwner, _ := database.IsBlacklistedUser(ownerID); blockedOwner && !isOwnerOrSudo(m.SenderID()) {
				if _, rerr := m.ReplyText(c, F(chatID, "blacklist_owner_blocked_leave"), nil); rerr != nil {
					logger.Error(rerr)
				}
				leaveChat(c, chatID)
				return nil
			}
		}

		logger.Debug("Bot added to " + utils.IntToStr(chatID))
		if _, rerr := m.ReplyText(c, F(chatID, "bot_added_normal"), nil); rerr != nil {
			return rerr
		}
		database.AddServedChat(chatID)

		totalChats, _ := database.TotalServedChats()
		logArgs := buildLogArgs(c, m, chatID, "added")
		logArgs["total_chats"] = totalChats

		if config.LoggerID != 0 && isLoggerEnabled() {
			if _, rerr := c.SendTextMessage(config.LoggerID, F(config.LoggerID, "logger_bot_added", logArgs), &td.SendTextMessageOpts{
				ParseMode: td.ParseModeHTML,
			}); rerr != nil {
				return rerr
			}
		}

		return nil
	}

	return nil
}

func handleUserBlocked(c *td.Client, u *td.UpdateChatMember, userID int64) {
	logger.Infof("User %d blocked or stopped the bot", userID)

	_ = database.RemoveServedUser(userID)

	if config.LoggerID == 0 || config.LoggerID == userID || !isLoggerEnabled() {
		return
	}

	user, err := c.GetUser(userID)
	userName := "N/A"
	userUsername := "N/A"
	userMention := fmt.Sprintf("<code>%d</code>", userID)
	if err == nil && user != nil {
		userName = strings.TrimSpace(user.FirstName + " " + user.LastName)
		if userName == "" {
			userName = "Deleted Account"
		}
		if user.Usernames != nil && len(user.Usernames.ActiveUsernames) > 0 {
			userUsername = "@" + user.Usernames.ActiveUsernames[0]
		}
		userMention = mentionOf(user, userID)
	}

	totalUsers, _ := database.TotalServedUsers()

	msg := F(config.LoggerID, "logger_user_blocked", locales.Arg{
		"user_name":     userName,
		"user_id":       userID,
		"user_username": userUsername,
		"mention":       userMention,
		"total_users":   totalUsers,
		"date_time":     time.Now().Format("02 Jan 2006 • 15:04"),
	})

	if _, err := c.SendTextMessage(config.LoggerID, msg, &td.SendTextMessageOpts{
		ParseMode: td.ParseModeHTML,
	}); err != nil {
		logger.Errorf("failed to send user-blocked log: %v", err)
	}
}

func handleBotRemoved(c *td.Client, u *td.UpdateChatMember, chatID int64) {
	if chatID > 0 {
		handleUserBlocked(c, u, chatID)
		return
	}

	logger.Debugf("Bot removed from chat %d", chatID)

	cleanScheduler.cancel(chatID)
	core.DropRoom(chatID)
	core.DropChatState(chatID)
	_ = database.RemoveServedChat(chatID)

	if config.LoggerID == 0 || config.LoggerID == chatID || !isLoggerEnabled() {
		return
	}

	groupName := "N/A"
	groupUsername := "N/A"
	if chat, err := c.GetChat(chatID); err == nil && chat != nil {
		if chat.Title != "" {
			groupName = chat.Title
		}
		if t, ok := chat.Type.(*td.ChatTypeSupergroup); ok {
			if sg, err := c.GetSupergroup(t.SupergroupId); err == nil && sg != nil && sg.Usernames != nil && len(sg.Usernames.ActiveUsernames) > 0 {
				groupUsername = "@" + sg.Usernames.ActiveUsernames[0]
			}
		}
	}

	actorID := u.ActorUserId
	actorName := "N/A"
	actorMention := fmt.Sprintf("<code>%d</code>", actorID)
	actorUsername := "N/A"
	if actor, err := c.GetUser(actorID); err == nil && actor != nil {
		actorName = strings.TrimSpace(actor.FirstName + " " + actor.LastName)
		if actorName == "" {
			actorName = "Unknown"
		}
		actorMention = mentionOf(actor, actorID)
		if actor.Usernames != nil && len(actor.Usernames.ActiveUsernames) > 0 {
			actorUsername = "@" + actor.Usernames.ActiveUsernames[0]
		}
	}

	totalChats, _ := database.TotalServedChats()

	msg := F(config.LoggerID, "logger_bot_removed", locales.Arg{
		"group_name":          groupName,
		"group_id":            chatID,
		"group_username":      groupUsername,
		"removed_by_name":     actorName,
		"removed_by_id":       actorID,
		"removed_by_username": actorMention,
		"actor_username":      actorUsername,
		"total_chats":         totalChats,
		"date_time":           time.Now().Format("02 Jan 2006 • 15:04"),
	})

	if _, err := c.SendTextMessage(config.LoggerID, msg, &td.SendTextMessageOpts{
		ParseMode: td.ParseModeHTML,
	}); err != nil {
		logger.Errorf("failed to send bot-removed log: %v", err)
	}
}

func handleVoiceChatAction(c *td.Client, m *td.Message) error {
	if isMaint, _ := database.IsMaintenanceEnabled(); isMaint {
		return nil
	}

	chatID := m.ChatID()
	isActive := true
	var duration int32

	if ended, ok := m.Content.(*td.MessageVideoChatEnded); ok {
		isActive = false
		duration = ended.Duration
	}

	s, err := core.ChatStateFor(chatID)
	if err != nil {
		logger.Errorf("Failed to get chat state for %d: %v", chatID, err)
		return nil
	}

	s.SetVoiceChat(isActive)

	msgKey := utils.IfElse(isActive, "voicechat_started", "voicechat_ended")
	if _, rerr := c.SendTextMessage(chatID, F(chatID, msgKey, locales.Arg{"duration": utils.FormatDuration(int(duration))}), nil); rerr != nil {
		logger.Error(rerr)
	}
	logger.Debugf("Voice chat %s in %d", msgKey, chatID)

	if !isActive {
		room, ok := core.RoomFor(chatID)
		go func() {
			time.Sleep(500 * time.Millisecond)
			if ok {
				scheduleOldPlayingMessage(room)
			}
			core.DropRoom(chatID)
			logger.Debugf("Room destroyed for ended voice chat in %d", chatID)
		}()
	}

	return nil
}

func buildLogArgs(c *td.Client, m *td.Message, chatID int64, action string) locales.Arg {
	groupName := "N/A"
	groupUsername := "N/A"
	membersCount := 0

	if chat, err := m.GetChat(c); err == nil && chat != nil {
		if chat.Title != "" {
			groupName = chat.Title
		}
		if t, ok := chat.Type.(*td.ChatTypeSupergroup); ok {
			if sg, err := c.GetSupergroup(t.SupergroupId); err == nil && sg != nil && sg.Usernames != nil && len(sg.Usernames.ActiveUsernames) > 0 {
				groupUsername = "@" + sg.Usernames.ActiveUsernames[0]
			}
			if sgFull, err := c.GetSupergroupFullInfo(t.SupergroupId); err == nil && sgFull != nil {
				membersCount = int(sgFull.MemberCount)
			}
		} else if t, ok := chat.Type.(*td.ChatTypeBasicGroup); ok {
			if bgFull, err := c.GetBasicGroupFullInfo(t.BasicGroupId); err == nil && bgFull != nil {
				membersCount = len(bgFull.Members)
			}
		}
	}

	sender, _ := m.GetUser(c)
	actorID := m.SenderID()
	actorName := "N/A"
	actorUsername := "N/A"
	if sender != nil {
		if full := strings.TrimSpace(sender.FirstName + " " + sender.LastName); full != "" {
			actorName = full
		}
		if sender.Usernames != nil && len(sender.Usernames.ActiveUsernames) > 0 {
			actorUsername = "@" + sender.Usernames.ActiveUsernames[0]
		}
	}

	membersStr := "N/A"
	if membersCount > 0 {
		membersStr = fmt.Sprintf("%d", membersCount)
	}

	return locales.Arg{
		"group_name":            groupName,
		"group_id":              chatID,
		"group_username":        groupUsername,
		"members_count":         membersStr,
		action + "_by_name":     actorName,
		action + "_by_id":       actorID,
		action + "_by_username": mentionOf(sender, actorID),
		"actor_username":        actorUsername,
		"date_time":             time.Now().Format("02 Jan 2006 • 15:04"),
	}
}
