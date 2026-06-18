package buildinfo

import (
  "runtime/debug"
  "strings"
)

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

// init fills Version / Commit / Date from runtime/debug.ReadBuildInfo when the
// ldflags injection (used by goreleaser) hasn't run. This is the case for
// users installing via `go install github.com/jr-k/d4s@vX.Y.Z`, where Go
// embeds the module version into the binary metadata.
func init() {
  info, ok := debug.ReadBuildInfo()
  if !ok {
    return
  }

  if Version == "dev" {
    if v := info.Main.Version; v != "" && v != "(devel)" {
      // goreleaser sets Version without the leading "v" (e.g. "0.49.99")
      // and the UI prepends a "v" at display time. Keep the same shape
      // here so we don't end up with "vv0.49.99". The "+dirty" suffix that
      // Go >= 1.24 appends when building from a dirty VCS tree is kept on
      // purpose so the displayed version reflects the real binary state.
      Version = strings.TrimPrefix(v, "v")
    }
  }

  for _, s := range info.Settings {
    switch s.Key {
    case "vcs.revision":
      if Commit == "none" && s.Value != "" {
        Commit = s.Value
      }
    case "vcs.time":
      if Date == "unknown" && s.Value != "" {
        Date = s.Value
      }
    }
  }
}

