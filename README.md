# D4S 🍊 (Docker for Services)

> **The K9s experience for Docker.**  
> Manage your Docker Swarm, Compose stacks, and Containers with a fancy, fast, and keyboard-centric Terminal User Interface.

D4S (pronounced *D-Force*) brings the power and ergonomics of K9s to the local Docker ecosystem. Stop wrestling with verbose CLI commands and start managing your containers like a pro.

## ✨ Features

- 🍊 **Fancy UI**: Modern TUI with Dracula theme, smooth navigation, and live updates.
- ⌨️ **Keyboard Centric**: Vim-like navigation (`j`/`k`), shortcuts for everything. No mouse needed.
- 🐳 **Full Scope**: Supports **Containers**, **Images**, **Volumes**, **Networks**.
- 🔍 **Powerful Search**: Instant fuzzy filtering (`/`) and command palette (`:`).
- 📊 **Live Stats**: Real-time CPU/Mem usage for containers and host context.
- 📜 **Advanced Logs**: Streaming logs with auto-scroll, timestamps toggle, and wrap mode.
- 🐚 **Quick Shell**: Drop into a container shell (`s`) in a split second.
- 🛠 **Contextual Actions**: Inspect, Restart, Stop, Prune, Delete with safety confirmations.
- 📦 **Compose Aware**: Easily identify containers belonging to Compose stacks.

## 🚀 Installation

### From Source
Requirement: Go 1.21+

```bash
git clone https://github.com/jessym/d4s.git
cd d4s
go build -o d4s cmd/d4s/main.go
sudo mv d4s /usr/local/bin/
```

### Quick Run
```bash
go run cmd/d4s/main.go
```

## 🎮 Shortcuts Cheat Sheet

### Navigation
- `Arrow Keys` or `j`/`k`: Navigate rows
- `Enter`: Inspect resource
- `Tab`: Switch views (implied via commands)
- `>` / `<`: Change sort column
- `+`:  Toggle Sort Order (ASC/DESC)
- `/`: Filter view
- `:`: Command Palette
- `?`: Help / Shortcuts

### Views
- `:c` : **C**ontainers
- `:i` : **I**mages
- `:v` : **V**olumes
- `:n` : **N**etworks

### Container Actions
- `l`: **L**ogs (Stream)
- `s`: **S**hell (Exec /bin/sh)
- `d`: **D**escribe / Inspect
- `r`: **R**estart
- `x`: Stop (eXit)
- `Ctrl+d`: Delete

### Log Viewer
- `s`: Toggle Auto-**S**croll
- `w`: Toggle **W**rap
- `t`: Toggle **T**imestamps
- `f`: Toggle **F**ullscreen (Border)
- `Shift+c`: Clear logs
- `Esc`: Back

## 🍊 Mascotte

Meet **Citrus**, our vitamin-packed helper ensuring your containers stay fresh and healthy! 🍊

---
*Built with Go & Tview. Inspired by the legendary K9s.*
