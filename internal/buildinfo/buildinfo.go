package buildinfo

var (
  Version = "dev"
  Author  = "jr-k"
  Email   = "jrk@jierka.com"
  Website = "https://d4scli.io"
  License = "Apache-2.0"
  Copyright = "2026 jr-k"
  Description = "A modern, fast, and user-friendly Docker CLI for the terminal"
  Features = []string{
    "🍊 Fancy UI: Modern TUI with Dracula theme, smooth navigation, and live updates.",
    "⌨️ Keyboard Centric: Vim-like navigation (`j`/`k`), shortcuts for everything. No mouse needed.",
    "🐳 Full Scope: Supports Containers, Images, Volumes, Networks.",
    "📦 Compose Aware: Easily identify containers belonging to Compose stacks.",
    "🐝 Swarm Aware: Supports Nodes, Services.",
    "🔍 Powerful Search: Instant fuzzy filtering (`/`) and command palette (`:`).",
    "📊 Live Stats: Real-time CPU/Mem usage for containers and host context.",
    "📜 Advanced Logs: Streaming logs with auto-scroll, timestamps toggle, and wrap mode.",
    "🐚 Quick Shell: Drop into a container shell (`s`) in a split second.",
    "🛠 Contextual Actions: Inspect, Restart, Stop, Prune, Delete with safety confirmations.",
  }
  Commit  = "none"
  Date    = "unknown"
)

func Long() string {
  return "Version: " + Version + "\nCommit:  " + Commit + "\nBuilt:   " + Date
}

