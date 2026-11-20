<p align="center">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,11,19,26&height=220&section=header&text=KanhaMusic&fontSize=62&fontColor=ffffff&animation=fadeIn&fontAlignY=38&desc=Next-Generation%20Telegram%20Voice%20Chat%20Music%20Engine&descAlignY=60&descAlign=50" width="100%"/>
</p>

<p align="center">
  <img src="https://readme-typing-svg.demolab.com?font=Fira+Code&weight=600&size=20&pause=1000&color=00E5FF&center=true&vCenter=true&width=620&lines=High-Performance+Go+Voice+Chat+Engine;Ultra-Low+Latency+Music+Streaming;Multi-Assistant+Seamless+Load+Balancing;Custom+Lyrics%2C+Thumbnails+%26+Animations;Developed+%26+Maintained+by+Kanha+Team" alt="Typing SVG" />
</p>

<p align="center">
  <a href="https://github.com/Kanha-Team/KanhaMusic/stargazers">
    <img src="https://img.shields.io/github/stars/Kanha-Team/KanhaMusic?style=for-the-badge&logo=github&color=FFD700&labelColor=1a1a2e" alt="GitHub Stars">
  </a>
  <a href="https://github.com/Kanha-Team/KanhaMusic/network/members">
    <img src="https://img.shields.io/github/forks/Kanha-Team/KanhaMusic?style=for-the-badge&logo=github&color=00C853&labelColor=1a1a2e" alt="GitHub Forks">
  </a>
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Language-Go%201.26-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=1a1a2e" alt="Go Version">
  </a>
  <a href="https://github.com/Kanha-Team/KanhaMusic/blob/dev/LICENSE">
    <img src="https://img.shields.io/badge/License-GPL--3.0-blue?style=for-the-badge&logo=opensourceinitiative&logoColor=white&labelColor=1a1a2e" alt="License">
  </a>
  <a href="https://t.me/MeowClone">
    <img src="https://img.shields.io/badge/Telegram-Support-26A69A?style=for-the-badge&logo=telegram&logoColor=white&labelColor=1a1a2e" alt="Support Chat">
  </a>
  <a href="https://t.me/MeowCloner">
    <img src="https://img.shields.io/badge/Telegram-Channel-0288D1?style=for-the-badge&logo=telegram&logoColor=white&labelColor=1a1a2e" alt="Updates Channel">
  </a>
</p>

---

## ⚡ Overview & Features

**KanhaMusic** is a high-performance, modular Telegram voice chat music bot written in **Go**. Built with native `ntgcalls` CGO audio bindings and direct TDLib MTProto client architecture, it delivers ultra-low latency, crystal-clear sound, and handles high-volume voice call playback with minimal memory overhead.

- 🚀 **Ultra-Low Latency Streaming:** Real-time PCM audio streaming with optimized buffer management and zero frame stutter.
- 🤖 **Multi-Assistant Pooling:** Automatically distributes group voice chats across multiple assistant userbots with automatic load-balancing.
- 🎵 **Multi-Platform Music Resolver:** Seamless resolution for YouTube, Spotify, Soundcloud, 90+ platforms via Ritesh API, and direct Telegram audio files.
- 🎛️ **Full Playback Controls:** Comprehensive controls including `/seek`, `/speed` (0.5x - 2.0x), `/loop`, `/shuffle`, and `/replay`.
- 🎭 **Dynamic Spoiler Mode:** Toggleable spoiler artwork for `/start`, `/ping`, and track thumbnails to keep chats clean and modern.
- 🌐 **Dual Language Support:** Native embedded translations in **English (`en`)** and **Hindi (`hi`)**.
- 🛡️ **Enterprise Security & Admin:** Dynamic sudo authorization, anti-flood protection, custom blacklist, and global maintenance switch.
- 🎨 **Rich Visual Feedback:** Dynamic cover thumbnails, inline control keyboards, and animated Telegram reaction effect triggers.

---

## ⚡ Performance Benchmark: Go vs Python Bots

| Metric / Benchmark | Standard Python Bots (PyTgCalls) | 🚀 KanhaMusic (Go + ntgcalls) | Advantage |
| :--- | :---: | :---: | :---: |
| **RAM / Memory Consumption** | 350 MB – 650 MB+ | **~35 MB – 70 MB** | **90% Less RAM** |
| **CPU Overhead** | High (Python GIL lock) | **Ultra Low (Native Goroutines)** | **Zero CPU bottleneck** |
| **Audio Start Latency** | 2.5s – 4.2s | **< 0.5s** | **Near-Instant Playback** |
| **Concurrent Voice Chats** | Struggles at 5–8 calls | **50+ simultaneous chats** | **Massive Scalability** |
| **Crash & Leak Resistance** | Prone to memory leaks | **Compiled static binary, rock-solid** | **24/7 Zero Downtime** |

---

## 🚀 One-Click Cloud Deploy

Deploy your own **KanhaMusic** instance in under 60 seconds on any cloud provider:

<p align="center">
  <a href="https://heroku.com/deploy?template=https://github.com/Kanha-Team/KanhaMusic">
    <img src="https://www.herokucdn.com/deploy/button.svg" alt="Deploy to Heroku" height="40">
  </a>
  &nbsp;&nbsp;&nbsp;
  <a href="https://railway.com/new/template?template=https://github.com/Kanha-Team/KanhaMusic">
    <img src="https://railway.com/button.svg" alt="Deploy on Railway" height="40">
  </a>
  &nbsp;&nbsp;&nbsp;
  <a href="https://render.com/deploy?repo=https://github.com/Kanha-Team/KanhaMusic">
    <img src="https://render.com/images/deploy-to-render-button.svg" alt="Deploy to Render" height="40">
  </a>
  &nbsp;&nbsp;&nbsp;
  <a href="https://app.koyeb.com/deploy?type=git&repository=Kanha-Team/KanhaMusic&branch=dev&name=kanhamusic&service_type=worker&env[API_ID]=&env[API_HASH]=&env[TOKEN]=&env[MONGO_DB_URI]=&env[STRING_SESSION]=&env[OWNER_ID]=">
    <img src="https://www.koyeb.com/static/images/deploy/button.svg" alt="Deploy to Koyeb" height="40">
  </a>
</p>

---

## 🛠️ Deployment Methods

<details>
  <summary><b>🖥️ [Click to Expand] VPS / Linux Server Installation Guide</b></summary>
  <br>

  ### 1. Install System Dependencies
  ```bash
  sudo apt-get update && sudo apt-get upgrade -y
  sudo apt-get install -y git curl ffmpeg build-essential
  ```

  ### 2. Clone the Repository
  ```bash
  git clone https://github.com/Kanha-Team/KanhaMusic.git
  cd KanhaMusic
  ```

  ### 3. Run Automated Dependency Installer
  ```bash
  # Automatically sets up ntgcalls, TDLib, yt-dlp, and Go tools
  bash install.sh
  ```

  ### 4. Setup Configuration
  ```bash
  cp sample.env .env
  nano .env  # Enter your API_ID, TOKEN, MONGO_DB_URI, etc.
  ```

  ### 5. Build & Run
  ```bash
  # Option A: Run directly
  bash start.sh

  # Option B: Run as a background Systemd Service (Auto-restarts 24/7)
  sudo bash setup_service.sh
  sudo systemctl start kanhamusic
  sudo systemctl status kanhamusic
  ```
</details>

<details>
  <summary><b>🐳 [Click to Expand] Docker & Docker Compose Deployment Guide</b></summary>
  <br>

  ```bash
  # 1. Clone repository
  git clone https://github.com/Kanha-Team/KanhaMusic.git
  cd KanhaMusic

  # 2. Configure environment
  cp sample.env .env
  nano .env

  # 3. Build & start container
  docker compose up -d --build

  # View live real-time logs
  docker compose logs -f kanhamusic
  ```
</details>

---

## ⚙️ Configuration Variables

| Variable | Required | Default | Description |
| :--- | :---: | :---: | :--- |
| `API_ID` | **Yes** | — | Telegram API ID from [my.telegram.org](https://my.telegram.org) |
| `API_HASH` | **Yes** | — | Telegram API Hash from [my.telegram.org](https://my.telegram.org) |
| `TOKEN` | **Yes** | — | Telegram Bot Token from [@BotFather](https://t.me/BotFather) |
| `MONGO_DB_URI` | **Yes** | — | MongoDB Connection String (`mongodb+srv://...`) |
| `STRING_SESSIONS` | **Yes** | — | Assistant account session strings (separated by commas or spaces) |
| `OWNER_ID` | **Yes** | — | Telegram User ID of the bot owner |
| `SESSION_TYPE` | No | `pyrogram` | String session format: `pyrogram`, `telethon`, or `gogram` |
| `DB_NAME` | No | `KanhaMusic` | MongoDB database name |
| `ENABLE_SPOILER` | No | `false` | Enable Telegram spoiler effect on bot images (`true` / `false`) |
| `LOGGER_ID` | No | `0` | Telegram Chat ID for logs and errors |
| `DURATION_LIMIT` | No | `4200` | Max track duration in seconds (70 minutes) |
| `QUEUE_LIMIT` | No | `24` | Maximum tracks allowed in queue per chat |
| `DEFAULT_LANG` | No | `en` | Default language code (`en` or `hi`) |
| `SUPPORT_CHAT` | No | `https://t.me/MeowClone` | Link to Telegram support group |
| `SUPPORT_CHANNEL` | No | `https://t.me/MeowCloner` | Link to Telegram announcements channel |
| `COOKIES_LINK` | No | — | batbin URL to YouTube cookies file for `yt-dlp` |

---

## 🎮 Bot Commands Reference

<details>
  <summary><b>🎵 [Click to Expand] Playback Commands</b></summary>
  <br>

  | Command | Description |
  | :--- | :--- |
  | `/play <query/url>` | Streams requested song into group voice chat |
  | `/fplay <query/url>` | Force plays track, skipping current song immediately |
  | `/pause` | Pauses active playback stream |
  | `/resume` | Resumes paused audio stream |
  | `/skip` | Skips current track to next in queue |
  | `/stop` | Stops playback and clears the playlist queue |
  | `/queue` | Displays the current playlist and upcoming tracks |
  | `/seek <seconds>` | Seeks playback to a specific timestamp |
  | `/speed <0.5-2.0>` | Changes audio playback speed dynamically |
  | `/loop <1-10>` | Loops the current track specified number of times |
  | `/shuffle` | Shuffles queued songs randomly |
  | `/replay` | Replays the current playing song from start |
</details>

<details>
  <summary><b>👥 [Click to Expand] Group Admin Commands</b></summary>
  <br>

  | Command | Description |
  | :--- | :--- |
  | `/auth <user>` | Authorizes user to use playback commands |
  | `/unauth <user>` | Revokes bot access authorization from a user |
  | `/authusers` | Lists all authorized users in current group |
  | `/language` | Changes bot language for current chat (`English` / `Hindi`) |
  | `/settings` | Opens the interactive visual settings control panel |
  | `/reload` | Refreshes admin cache and reconnects the assistant |
  | `/cplay <channel>` | Streams audio directly linked to a Telegram channel |
</details>

<details>
  <summary><b>🛡️ [Click to Expand] Owner & Sudo Commands</b></summary>
  <br>

  | Command | Description |
  | :--- | :--- |
  | `/addsudo <user>` | Adds a user to sudoers list |
  | `/delsudo <user>` | Removes a user from sudoers list |
  | `/sudolist` | Shows list of all current bot sudoers |
  | `/stats` | Shows server performance, memory, and database metrics |
  | `/broadcast <msg>` | Broadcasts announcements to all served chats and users |
  | `/maintenance` | Toggles global bot maintenance mode |
  | `/eval <code>` | Executes internal runtime evaluation (Owner only) |
</details>

---

## 📁 Project Architecture

```
KanhaMusic/
├── cmd/app/              # Application entry point & dependencies verification
├── config/               # Flexible configuration loader & cookies system
├── kanha/
│   ├── core/             # State machine, room controllers, & assistant pooling
│   ├── database/         # MongoDB schemas, caching layers, & automated migrations
│   ├── locales/          # YAML-based translation packs (English & Hindi)
│   ├── meow/             # High-performance Telegram MTProto client engine
│   ├── modules/          # Command modules (play, skip, pause, settings, etc.)
│   ├── platforms/        # Media stream resolvers (YouTube, Spotify, Telegram)
│   └── utils/            # Admin checks, parsing, formatters, and helpers
├── Dockerfile            # Multi-stage production container build
├── docker-compose.yml    # Docker Compose orchestration recipe
├── install.sh            # Automated server setup and dependency script
├── start.sh              # Production runner with auto-restart & library bindings
└── sample.env            # Template environment variables file
```

---

## 📜 License

Distributed under the **GNU General Public License v3.0**. See the [`LICENSE`](./LICENSE) file for full details.

<p align="center">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,11,19,26&height=100&section=footer" width="100%"/>
</p>

<p align="center">
  <b>Developed & Maintained with ❤️ by <a href="https://github.com/Kanha-Team">Kanha Team</a></b><br>
  <sub>🚀 Explore our Portfolio: <a href="https://kanha-portfolio-ayush.vercel.app">kanha-portfolio-ayush.vercel.app</a></sub>
</p>
<!-- kanha-sync-20261120 -->
