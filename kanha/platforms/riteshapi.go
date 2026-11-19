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

package platforms

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"KanhaMusic/config"
	state "KanhaMusic/kanha/core/models"
	"KanhaMusic/kanha/logger"
	"KanhaMusic/kanha/utils"

	td "github.com/Kanha/Meow"
)

const (
	audioDownloadTimeout  = 300 * time.Second
	videoDownloadTimeout  = 600 * time.Second
	minValidDownloadBytes = 10_000
)

const (
	PlatformRiteshApi state.PlatformName = "RiteshApi"
	PlatformSpotify   state.PlatformName = "Spotify"
)

// RiteshApiPlatform integrates the official high-speed Ritesh Music API (https://www.riteshyt.in/docs).
type RiteshApiPlatform struct{}

func init() {
	Register(&RiteshApiPlatform{})
	Register(&SpotifyPlatform{
		cache: utils.NewCache[string, []*state.Track](1 * time.Hour),
	})
}

func (r *RiteshApiPlatform) Name() state.PlatformName { return PlatformRiteshApi }
func (r *RiteshApiPlatform) Priority() int             { return 85 }

// supportedPlatformDomains contains 90+ platforms supported by Ritesh API /api/platform/extract
var supportedPlatformDomains = []string{
	// Social & Video
	"instagram.com", "instagr.am",
	"tiktok.com",
	"twitter.com", "x.com",
	"facebook.com", "fb.watch", "fb.com",
	"threads.net",
	"snapchat.com",
	"reddit.com", "v.redd.it",
	"pinterest.com", "pin.it",
	"linkedin.com",
	"bsky.app",
	"tumblr.com",
	"mastodon",
	"douyin.com",
	"kuaishou.com",
	"triller.co",
	"likee.video", "likee.com",
	"weibo.com", "weibo.cn",
	"xiaohongshu.com",
	"9gag.com",

	// Streaming & Music
	"jiosaavn.com", "saavn.com",
	"soundcloud.com",
	"music.apple.com", "apple.co",
	"gaana.com",
	"deezer.com", "deezer.page.link",
	"resso.com",
	"tidal.com",
	"bandcamp.com",
	"mixcloud.com",
	"audiomack.com",
	"hearthis.at",
	"jamendo.com",
	"reverbnation.com",
	"audius.co",
	"boomplay.com",
	"anghami.com",
	"wynk.in",
	"hungama.com",
	"napster.com",
	"kugou.com",
	"kuwo.cn",
	"qq.com",
	"music.163.com", "163sub.com",
	"iheart.com",
	"tunein.com",
	"pandora.com",
	"last.fm",
	"genius.com",

	// Video Portals
	"dailymotion.com", "dai.ly",
	"vimeo.com",
	"vk.com",
	"rumble.com",
	"streamable.com",
	"ok.ru",
	"coub.com",
	"nicovideo.jp", "niconico.jp",
	"odysee.com", "lbry.tv",
	"bitchute.com",
	"veoh.com",
	"rutube.ru",
	"ted.com",
	"archive.org",
	"youku.com",
	"iq.com", "iqiyi.com",
	"peertube",
	"afreecatv.com",
	"trovo.live",
	"twitch.tv",
	"kick.com",
	"bilibili.com", "b23.tv",
	"mxplayer.in",
	"redbull.com",
	"izlesene.com",
	"newgrounds.com",
	"imdb.com",

	// Media Hosting & Storage
	"imgur.com",
	"loom.com",
	"wistia.com", "wistia.net",
	"flickr.com",
	"smugmug.com",
	"artstation.com",
	"deviantart.com",
	"pixiv.net",
	"behance.net",
	"dribbble.com",
	"unsplash.com",
	"pexels.com",
	"pixabay.com",
	"giphy.com",
	"tenor.com",
	"terabox.com", "teraboxapp.com", "1024tera.com",
}

func isSupportedPlatformURL(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if !strings.HasPrefix(q, "http://") && !strings.HasPrefix(q, "https://") {
		return false
	}
	u, err := url.Parse(q)
	if err == nil && u.Host != "" {
		host := strings.ToLower(u.Host)
		for _, domain := range supportedPlatformDomains {
			if strings.Contains(host, domain) || strings.Contains(q, domain) {
				return true
			}
		}
	} else {
		for _, domain := range supportedPlatformDomains {
			if strings.Contains(q, domain) {
				return true
			}
		}
	}
	return false
}

// CanGet reports whether Ritesh API can resolve multi-platform URLs (Instagram, TikTok, SoundCloud, etc.)
func (r *RiteshApiPlatform) CanGet(query string) bool {
	if config.RiteshAPIKey == "" {
		return false
	}
	return isSupportedPlatformURL(query)
}

// Get resolves media using Ritesh API's multi-platform extractor.
func (r *RiteshApiPlatform) Get(query string, video bool) ([]*state.Track, error) {
	return RiteshExtractPlatform(query, video)
}

// CanDownload reports whether Ritesh API can download tracks from YouTube, Spotify, or RiteshApi.
func (r *RiteshApiPlatform) CanDownload(source state.PlatformName) bool {
	if config.RiteshAPIKey == "" {
		return false
	}
	return source == PlatformYouTube || source == PlatformSpotify || source == PlatformRiteshApi
}

// Download downloads audio/video using Ritesh Music API stream endpoints.
func (r *RiteshApiPlatform) Download(
	ctx context.Context,
	track *state.Track,
	_ *td.Message,
) (string, error) {
	if p := findFile(track); p != "" {
		logger.Debug("RiteshApi: cache hit " + p)
		return p, nil
	}

	if track.ID == "" {
		return "", errors.New("riteshapi: missing video id")
	}

	apiURL := getRiteshBaseURL()
	dtype, quality, ext, timeout := "audio", "320k", ".mp3", audioDownloadTimeout
	if track.Video {
		dtype, quality, ext, timeout = "video", "720p", ".mp4", videoDownloadTimeout
	}

	dctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	path := getPath(track, ext)

	// Multi-platform URL handling (Instagram, TikTok, SoundCloud, Facebook, etc.)
	if track.Source == PlatformRiteshApi && isSupportedPlatformURL(track.URL) {
		extractURL := fmt.Sprintf("%s/api/platform/extract?url=%s&api_key=%s",
			apiURL, url.QueryEscape(track.URL), url.QueryEscape(config.RiteshAPIKey))

		var extResp struct {
			Success     bool   `json:"success"`
			DownloadURL string `json:"download_url"`
			Error       string `json:"error"`
		}

		rResp, err := rc.R().
			SetContext(dctx).
			SetHeader("User-Agent", "KanhaMusic/1.0").
			SetResult(&extResp).
			Get(extractURL)
		if err != nil {
			return "", sanitizeAPIError(err, config.RiteshAPIKey)
		}
		if rResp.IsStatusFailure() || !extResp.Success || extResp.DownloadURL == "" {
			errMsg := extResp.Error
			if errMsg == "" {
				errMsg = fmt.Sprintf("HTTP %d", rResp.StatusCode())
			}
			return "", fmt.Errorf("riteshapi extract failed: %s", errMsg)
		}

		dlURL := extResp.DownloadURL
		if strings.HasPrefix(dlURL, "/") {
			dlURL = apiURL + dlURL
		}
		if config.RiteshAPIKey != "" && !strings.Contains(dlURL, "api_key=") {
			if strings.Contains(dlURL, "?") {
				dlURL += "&api_key=" + url.QueryEscape(config.RiteshAPIKey)
			} else {
				dlURL += "?api_key=" + url.QueryEscape(config.RiteshAPIKey)
			}
		}

		dlResp, dlErr := rc.R().
			SetContext(dctx).
			SetHeader("User-Agent", "KanhaMusic/1.0").
			SetResponseSaveFileName(path).
			Get(dlURL)
		if dlErr != nil {
			os.Remove(path)
			return "", dlErr
		}
		if dlResp.IsStatusFailure() {
			os.Remove(path)
			return "", fmt.Errorf("riteshapi media download failed with HTTP %d", dlResp.StatusCode())
		}
		info, statErr := os.Stat(path)
		if statErr != nil || info.Size() < minValidDownloadBytes {
			os.Remove(path)
			return "", errors.New("riteshapi returned an invalid or empty file")
		}
		return path, nil
	}

	// Primary Endpoint: /download with HTTP 206 range seeking support
	streamURL := fmt.Sprintf(
		"%s/download?query=%s&dl_type=%s&quality=%s&api_key=%s",
		apiURL, url.QueryEscape(track.ID), dtype, quality, url.QueryEscape(config.RiteshAPIKey),
	)

	logger.Infof("[RiteshApi] Downloading track %s (%s) from %s", track.ID, dtype, apiURL)

	rResp, err := rc.R().
		SetContext(dctx).
		SetHeader("User-Agent", "KanhaMusic/1.0").
		SetResponseSaveFileName(path).
		Get(streamURL)

	isJSON := false
	if rResp != nil && strings.Contains(strings.ToLower(rResp.Header().Get("Content-Type")), "application/json") {
		isJSON = true
	}

	// Fallback Endpoint for Audio: /api/musicbot/play/{vidid}?api_key={key}
	if (err != nil || rResp.IsStatusFailure() || isJSON) && !track.Video {
		os.Remove(path)
		fallbackURL := fmt.Sprintf("%s/api/musicbot/play/%s?api_key=%s",
			apiURL, url.PathEscape(track.ID), url.QueryEscape(config.RiteshAPIKey))

		logger.Warnf("[RiteshApi] Primary download returned error or JSON, trying musicbot redirect fallback...")
		rResp, err = rc.R().
			SetContext(dctx).
			SetHeader("User-Agent", "KanhaMusic/1.0").
			SetResponseSaveFileName(path).
			Get(fallbackURL)

		if rResp != nil && strings.Contains(strings.ToLower(rResp.Header().Get("Content-Type")), "application/json") {
			isJSON = true
		} else {
			isJSON = false
		}
	}

	if err != nil {
		os.Remove(path)
		if isDownloadCancelled(err) {
			return "", err
		}
		return "", sanitizeAPIError(
			fmt.Errorf("riteshapi request failed: %w", err),
			config.RiteshAPIKey,
		)
	}

	if rResp.IsStatusFailure() || isJSON {
		os.Remove(path)
		return "", sanitizeAPIError(fmt.Errorf(
			"riteshapi returned HTTP %d (or invalid content-type)", rResp.StatusCode(),
		), config.RiteshAPIKey)
	}

	info, statErr := os.Stat(path)
	if statErr != nil || info.Size() < minValidDownloadBytes {
		os.Remove(path)
		return "", errors.New("riteshapi returned an invalid or empty file")
	}

	return path, nil
}

func getRiteshBaseURL() string {
	u := strings.TrimRight(config.RiteshAPIURL, "/")
	if u == "" {
		return "https://api.riteshyt.in"
	}
	return u
}

func parseDurationString(durStr string) int {
	parts := strings.Split(strings.TrimSpace(durStr), ":")
	switch len(parts) {
	case 2:
		m, _ := strconv.Atoi(parts[0])
		s, _ := strconv.Atoi(parts[1])
		return m*60 + s
	case 3:
		h, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])
		s, _ := strconv.Atoi(parts[2])
		return h*3600 + m*60 + s
	default:
		return 0
	}
}

// ============================================================================
// Official Ritesh API Endpoints (https://www.riteshyt.in/docs)
// ============================================================================

// RiteshHealth checks system health status via GET /health endpoint.
func RiteshHealth() (bool, error) {
	apiURL := getRiteshBaseURL()
	reqURL := apiURL + "/health"

	var resp struct {
		Status string `json:"status"`
	}

	rResp, err := rc.R().
		SetTimeout(5 * time.Second).
		SetHeader("User-Agent", "KanhaMusic/1.0").
		SetResult(&resp).
		Get(reqURL)
	if err != nil {
		return false, err
	}
	return !rResp.IsStatusFailure() && (resp.Status == "ok" || resp.Status == "healthy"), nil
}

// RiteshSearch searches YouTube for tracks via GET /search?q={query}&limit={limit}&api_key={api_key}.
func RiteshSearch(query string, limit int) ([]*state.Track, error) {
	if config.RiteshAPIKey == "" {
		return nil, errors.New("riteshapi: api key is not configured")
	}
	if limit <= 0 {
		limit = 1
	}

	apiURL := getRiteshBaseURL()
	reqURL := fmt.Sprintf("%s/search?q=%s&limit=%d&api_key=%s",
		apiURL, url.QueryEscape(strings.TrimSpace(query)), limit, url.QueryEscape(config.RiteshAPIKey))

	var resp struct {
		Result []struct {
			Type       string `json:"type"`
			ID         string `json:"id"`
			Title      string `json:"title"`
			Duration   string `json:"duration"`
			Thumbnails []struct {
				URL string `json:"url"`
			} `json:"thumbnails"`
			Link string `json:"link"`
		} `json:"result"`
	}

	rResp, err := rc.R().
		SetHeader("User-Agent", "KanhaMusic/1.0").
		SetResult(&resp).
		Get(reqURL)
	if err != nil {
		return nil, sanitizeAPIError(err, config.RiteshAPIKey)
	}
	if rResp.IsStatusFailure() || len(resp.Result) == 0 {
		return nil, fmt.Errorf("ritesh search failed with HTTP %d", rResp.StatusCode())
	}

	tracks := make([]*state.Track, 0, len(resp.Result))
	for _, item := range resp.Result {
		if item.ID == "" || item.Title == "" {
			continue
		}
		artwork := ""
		if len(item.Thumbnails) > 0 {
			artwork = item.Thumbnails[len(item.Thumbnails)-1].URL
		}
		link := item.Link
		if link == "" {
			link = "https://www.youtube.com/watch?v=" + item.ID
		}
		tracks = append(tracks, &state.Track{
			ID:       item.ID,
			Title:    item.Title,
			Duration: parseDurationString(item.Duration),
			Artwork:  artwork,
			URL:      link,
			Source:   PlatformYouTube,
		})
	}

	if len(tracks) == 0 {
		return nil, errors.New("ritesh search returned no valid tracks")
	}
	return tracks, nil
}

// RiteshDetails retrieves single video or Spotify track metadata via GET /details?link={link_or_id}&api_key={api_key}.
func RiteshDetails(linkOrID string) (*state.Track, error) {
	if config.RiteshAPIKey == "" {
		return nil, errors.New("riteshapi: api key is not configured")
	}

	apiURL := getRiteshBaseURL()
	reqURL := fmt.Sprintf("%s/details?link=%s&api_key=%s",
		apiURL, url.QueryEscape(strings.TrimSpace(linkOrID)), url.QueryEscape(config.RiteshAPIKey))

	var resp struct {
		Title       string `json:"title"`
		DurationMin string `json:"duration_min"`
		DurationSec int    `json:"duration_sec"`
		Thumbnail   string `json:"thumbnail"`
		VidID       string `json:"vidid"`
		Link        string `json:"link"`
	}

	rResp, err := rc.R().
		SetHeader("User-Agent", "KanhaMusic/1.0").
		SetResult(&resp).
		Get(reqURL)
	if err != nil {
		return nil, sanitizeAPIError(err, config.RiteshAPIKey)
	}
	if rResp.IsStatusFailure() || resp.Title == "" {
		return nil, fmt.Errorf("ritesh details failed with HTTP %d", rResp.StatusCode())
	}

	trackID := resp.VidID
	if trackID == "" {
		trackID = linkOrID
	}
	trackURL := resp.Link
	if trackURL == "" {
		trackURL = "https://www.youtube.com/watch?v=" + trackID
	}

	dur := resp.DurationSec
	if dur <= 0 && resp.DurationMin != "" {
		dur = parseDurationString(resp.DurationMin)
	}

	return &state.Track{
		ID:       trackID,
		Title:    resp.Title,
		Duration: dur,
		Artwork:  resp.Thumbnail,
		URL:      trackURL,
		Source:   PlatformYouTube,
	}, nil
}

// RiteshPlaylist extracts tracks from YouTube/Spotify playlists via GET /playlist?link={playlist_url_or_id}&limit={limit}&api_key={api_key}.
func RiteshPlaylist(linkOrID string, limit int) ([]*state.Track, error) {
	if config.RiteshAPIKey == "" {
		return nil, errors.New("riteshapi: api key is not configured")
	}
	if limit <= 0 {
		limit = 25
	}

	apiURL := getRiteshBaseURL()
	reqURL := fmt.Sprintf("%s/playlist?link=%s&limit=%d&api_key=%s",
		apiURL, url.QueryEscape(strings.TrimSpace(linkOrID)), limit, url.QueryEscape(config.RiteshAPIKey))

	var resp struct {
		Videos []struct {
			VidID       string `json:"vidid"`
			Title       string `json:"title"`
			DurationMin string `json:"duration_min"`
			DurationSec int    `json:"duration_sec"`
			Thumbnail   string `json:"thumbnail"`
		} `json:"videos"`
	}

	rResp, err := rc.R().
		SetHeader("User-Agent", "KanhaMusic/1.0").
		SetResult(&resp).
		Get(reqURL)
	if err != nil {
		return nil, sanitizeAPIError(err, config.RiteshAPIKey)
	}
	if rResp.IsStatusFailure() || len(resp.Videos) == 0 {
		return nil, fmt.Errorf("ritesh playlist failed with HTTP %d", rResp.StatusCode())
	}

	tracks := make([]*state.Track, 0, len(resp.Videos))
	for _, v := range resp.Videos {
		if v.Title == "" {
			continue
		}
		dur := v.DurationSec
		if dur <= 0 && v.DurationMin != "" {
			dur = parseDurationString(v.DurationMin)
		}
		vid := v.VidID
		if vid == "" {
			vid = v.Title
		}
		tracks = append(tracks, &state.Track{
			ID:       vid,
			Title:    v.Title,
			Duration: dur,
			Artwork:  v.Thumbnail,
			URL:      "https://www.youtube.com/watch?v=" + vid,
			Source:   PlatformYouTube,
		})
	}
	return tracks, nil
}

// RiteshVideoInfo retrieves technical format lists and media metadata via GET /video_info?vidid={vidid}&api_key={api_key}.
func RiteshVideoInfo(vidid string) (map[string]any, error) {
	if config.RiteshAPIKey == "" {
		return nil, errors.New("riteshapi: api key is not configured")
	}

	apiURL := getRiteshBaseURL()
	reqURL := fmt.Sprintf("%s/video_info?vidid=%s&api_key=%s",
		apiURL, url.QueryEscape(strings.TrimSpace(vidid)), url.QueryEscape(config.RiteshAPIKey))

	var resp map[string]any
	rResp, err := rc.R().
		SetHeader("User-Agent", "KanhaMusic/1.0").
		SetResult(&resp).
		Get(reqURL)
	if err != nil {
		return nil, sanitizeAPIError(err, config.RiteshAPIKey)
	}
	if rResp.IsStatusFailure() {
		return nil, fmt.Errorf("ritesh video_info failed with HTTP %d", rResp.StatusCode())
	}
	return resp, nil
}

// RiteshMusicBotPlayURL returns the direct high-speed redirect stream URL for voice chats via GET /api/musicbot/play/{vidid}?api_key={api_key}.
func RiteshMusicBotPlayURL(vidid string) string {
	apiURL := getRiteshBaseURL()
	return fmt.Sprintf("%s/api/musicbot/play/%s?api_key=%s",
		apiURL, url.PathEscape(strings.TrimSpace(vidid)), url.QueryEscape(config.RiteshAPIKey))
}

// RiteshDownloadURL returns the direct download/streaming endpoint URL via GET /download?query={query}&dl_type={audio|video}&quality={quality}&api_key={api_key}.
func RiteshDownloadURL(query, dlType, quality string) string {
	apiURL := getRiteshBaseURL()
	if dlType == "" {
		dlType = "audio"
	}
	if quality == "" {
		if dlType == "audio" {
			quality = "320k"
		} else {
			quality = "720p"
		}
	}
	return fmt.Sprintf("%s/download?query=%s&dl_type=%s&quality=%s&api_key=%s",
		apiURL, url.QueryEscape(strings.TrimSpace(query)), url.QueryEscape(dlType), url.QueryEscape(quality), url.QueryEscape(config.RiteshAPIKey))
}

// RiteshExtractPlatform extracts media from 90+ social/streaming platforms via GET /api/platform/extract?url={url}&api_key={api_key}.
func RiteshExtractPlatform(query string, video bool) ([]*state.Track, error) {
	if config.RiteshAPIKey == "" {
		return nil, errors.New("riteshapi: api key is required for multi-platform extraction")
	}

	apiURL := getRiteshBaseURL()
	reqURL := fmt.Sprintf("%s/api/platform/extract?url=%s&api_key=%s",
		apiURL, url.QueryEscape(strings.TrimSpace(query)), url.QueryEscape(config.RiteshAPIKey))

	var resp struct {
		Success     bool   `json:"success"`
		Platform    string `json:"platform"`
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Thumbnail   string `json:"thumbnail"`
		Duration    int    `json:"duration"`
		MediaType   string `json:"media_type"`
		DownloadURL string `json:"download_url"`
		Error       string `json:"error"`
	}

	rResp, err := rc.R().
		SetHeader("User-Agent", "KanhaMusic/1.0").
		SetResult(&resp).
		Get(reqURL)
	if err != nil {
		return nil, sanitizeAPIError(fmt.Errorf("riteshapi extract request failed: %w", err), config.RiteshAPIKey)
	}
	if rResp.IsStatusFailure() || !resp.Success {
		errMsg := resp.Error
		if errMsg == "" {
			errMsg = fmt.Sprintf("status code %d", rResp.StatusCode())
		}
		return nil, fmt.Errorf("riteshapi extract failed: %s", errMsg)
	}

	trackID := resp.ID
	if trackID == "" {
		h := md5.Sum([]byte(query))
		trackID = hex.EncodeToString(h[:10])
	}

	title := strings.TrimSpace(resp.Title)
	if title == "" {
		if resp.Description != "" {
			lines := strings.Split(strings.TrimSpace(resp.Description), "\n")
			if len(lines) > 0 && lines[0] != "" {
				title = strings.TrimSpace(lines[0])
			}
		}
		if title == "" {
			if resp.Platform != "" {
				title = strings.ToUpper(resp.Platform[:1]) + resp.Platform[1:] + " Media"
			} else {
				title = "Multi-Platform Track"
			}
		}
	}

	isVideo := video || resp.MediaType == "video"
	track := &state.Track{
		ID:       trackID,
		Title:    title,
		Duration: resp.Duration,
		Artwork:  resp.Thumbnail,
		URL:      query,
		Video:    isVideo,
		Source:   PlatformRiteshApi,
	}

	return []*state.Track{track}, nil
}

// ============================================================================
// Spotify Platform (Resolves Spotify Tracks & Playlists via Ritesh API / Embed)
// ============================================================================

type SpotifyPlatform struct {
	cache *utils.Cache[string, []*state.Track]
}

var (
	spotifyEntityRe = regexp.MustCompile(
		`(?i)open\.spotify\.com/(?:intl-[a-z-]+/)?(?:embed/)?(track|album|playlist|artist)/([A-Za-z0-9]{22})`,
	)
	spotifyNextDataRe = regexp.MustCompile(
		`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`,
	)
	spotifyHTTP = &http.Client{Timeout: 15 * time.Second}
)

func (s *SpotifyPlatform) Name() state.PlatformName { return PlatformSpotify }
func (s *SpotifyPlatform) Priority() int            { return 80 }

func (s *SpotifyPlatform) CanGet(query string) bool {
	q := strings.TrimSpace(query)
	return spotifyEntityRe.MatchString(q)
}

func (s *SpotifyPlatform) Get(query string, _ bool) ([]*state.Track, error) {
	link := strings.TrimSpace(query)

	m := spotifyEntityRe.FindStringSubmatch(link)
	if m == nil {
		return nil, errors.New("spotify: unsupported link")
	}
	kind, id := strings.ToLower(m[1]), m[2]

	cacheKey := kind + ":" + id
	if cached, ok := s.cache.Get(cacheKey); ok && len(cached) > 0 {
		return cloneTracks(cached), nil
	}

	var (
		tracks []*state.Track
		err    error
	)

	// 1. Try Ritesh API if API key is configured
	if config.RiteshAPIKey != "" {
		if kind == "track" {
			tracks, err = s.fetchTrackViaRitesh(link)
		} else {
			tracks, err = s.fetchPlaylistViaRitesh(link)
		}
		if err == nil && len(tracks) > 0 {
			s.cache.Set(cacheKey, tracks)
			return cloneTracks(tracks), nil
		}
		logger.Warnf("[Spotify] Ritesh API lookup failed for %s (%v), falling back to embed resolver", link, err)
	}

	// 2. Standalone Embed Scraper Fallback
	if kind == "track" {
		tracks, err = s.fetchTrackEmbed(id)
	} else {
		tracks, err = s.fetchCollectionEmbed(kind, id)
	}
	if err != nil {
		return nil, fmt.Errorf("spotify: %w", err)
	}
	if len(tracks) == 0 {
		return nil, errors.New("spotify: no tracks found")
	}

	s.cache.Set(cacheKey, tracks)
	return cloneTracks(tracks), nil
}

func (s *SpotifyPlatform) CanDownload(_ state.PlatformName) bool { return false }

func (s *SpotifyPlatform) Download(_ context.Context, _ *state.Track, _ *td.Message) (string, error) {
	return "", errors.New("spotify platform delegates downloads to audio resolver")
}

// fetchTrackViaRitesh resolves Spotify track using Ritesh API /details endpoint
func (s *SpotifyPlatform) fetchTrackViaRitesh(link string) ([]*state.Track, error) {
	t, err := RiteshDetails(link)
	if err != nil {
		return nil, err
	}
	return []*state.Track{t}, nil
}

// fetchPlaylistViaRitesh resolves Spotify collection using Ritesh API /playlist endpoint
func (s *SpotifyPlatform) fetchPlaylistViaRitesh(link string) ([]*state.Track, error) {
	return RiteshPlaylist(link, config.QueueLimit)
}

// Embed scrapers for fallback when API key is not present or offline
func (s *SpotifyPlatform) fetchTrackEmbed(id string) ([]*state.Track, error) {
	entity, err := s.fetchEntity("track", id)
	if err == nil {
		if t := spotifyTrackFromEntity(entity, id); t != nil {
			return []*state.Track{t}, nil
		}
	}
	return s.fetchOEmbed(id)
}

func (s *SpotifyPlatform) fetchCollectionEmbed(kind, id string) ([]*state.Track, error) {
	entity, err := s.fetchEntity(kind, id)
	if err != nil {
		return nil, err
	}

	list, _ := entity["trackList"].([]any)
	cover := spotifyCover(entity)

	limit := config.QueueLimit
	if limit <= 0 {
		limit = 15
	}

	var tracks []*state.Track
	for _, raw := range list {
		if len(tracks) >= limit {
			break
		}
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		title, _ := item["title"].(string)
		sub, _ := item["subtitle"].(string)
		if title == "" {
			continue
		}
		q := title
		if sub != "" {
			q = sub + " - " + title
		}
		dur := 0
		if d, ok := item["duration"].(float64); ok {
			dur = int(d / 1000)
		}
		tracks = append(tracks, &state.Track{
			ID:       q,
			Title:    q,
			Duration: dur,
			Artwork:  cover,
			URL:      "https://open.spotify.com/" + kind + "/" + id,
			Source:   PlatformYouTube,
		})
	}
	return tracks, nil
}

func (s *SpotifyPlatform) fetchEntity(kind, id string) (map[string]any, error) {
	embedURL := "https://open.spotify.com/embed/" + kind + "/" + id
	req, err := http.NewRequest("GET", embedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := spotifyHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := spotifyNextDataRe.FindSubmatch(body)
	if m == nil {
		return nil, errors.New("missing __NEXT_DATA__")
	}

	var root struct {
		Props struct {
			PageProps struct {
				State struct {
					Data struct {
						Entity map[string]any `json:"entity"`
					} `json:"data"`
				} `json:"state"`
			} `json:"pageProps"`
		} `json:"props"`
	}

	if err := json.Unmarshal(m[1], &root); err != nil {
		return nil, err
	}
	if root.Props.PageProps.State.Data.Entity == nil {
		return nil, errors.New("empty entity payload")
	}
	return root.Props.PageProps.State.Data.Entity, nil
}

func (s *SpotifyPlatform) fetchOEmbed(id string) ([]*state.Track, error) {
	oembedURL := "https://open.spotify.com/oembed?url=https://open.spotify.com/track/" + id
	req, err := http.NewRequest("GET", oembedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := spotifyHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oembed HTTP %d", resp.StatusCode)
	}

	var payload struct {
		Title        string `json:"title"`
		ThumbnailURL string `json:"thumbnail_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Title == "" {
		return nil, errors.New("oembed returned empty title")
	}

	return []*state.Track{{
		ID:      payload.Title,
		Title:   payload.Title,
		Artwork: payload.ThumbnailURL,
		URL:     "https://open.spotify.com/track/" + id,
		Source:  PlatformYouTube,
	}}, nil
}

func spotifyTrackFromEntity(entity map[string]any, id string) *state.Track {
	name, _ := entity["name"].(string)
	if name == "" {
		name, _ = entity["title"].(string)
	}
	if name == "" {
		return nil
	}

	artist := ""
	if artists, ok := entity["artists"].([]any); ok && len(artists) > 0 {
		if a, ok := artists[0].(map[string]any); ok {
			artist, _ = a["name"].(string)
		}
	}

	title := name
	if artist != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(artist)) {
		title = artist + " - " + name
	}

	dur := 0
	if d, ok := entity["duration"].(float64); ok {
		dur = int(d / 1000)
	}

	return &state.Track{
		ID:       title,
		Title:    title,
		Duration: dur,
		Artwork:  spotifyCover(entity),
		URL:      "https://open.spotify.com/track/" + id,
		Source:   PlatformYouTube,
	}
}

func spotifyCover(entity map[string]any) string {
	if cover, ok := entity["coverArt"].(map[string]any); ok {
		if sources, ok := cover["sources"].([]any); ok && len(sources) > 0 {
			if s, ok := sources[len(sources)-1].(map[string]any); ok {
				if u, ok := s["url"].(string); ok {
					return u
				}
			}
		}
	}
	return ""
}

func cloneTracks(in []*state.Track) []*state.Track {
	out := make([]*state.Track, 0, len(in))
	for _, t := range in {
		if t == nil {
			continue
		}
		c := *t
		out = append(out, &c)
	}
	return out
}

