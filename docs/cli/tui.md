# TUI

KubeMemo ships with an interactive terminal UI for browsing memos across namespaces, kinds, and runtime state.

## Launch

```bash
kubememo tui
```

PowerShell wrapper:

```powershell
Open-KubeMemoTui
```

## Defaults

- Includes runtime memos by default
- Auto-refreshes by default
- Supports namespace, kind, and text filtering
- Uses a split memo-board layout with detail pane rendering
- Paginates larger memo sets
- Stacks the detail pane below the memo list on narrower terminals

## Navigation

```text
[Arrows]/[j][k] move
[g]/[G] jump top/end
[Enter] focus/collapse detail
[PgUp]/[PgDn] or [u][d] scroll
[/] text filter
[n] namespace filter
[c] kind filter
[a]/[m]/[t] all/durable/runtime
[[] previous page
[]] next page
[x] clear filters
[r] refresh
[?] help
[q] quit
```

`j`/`k` and the arrow keys will also roll across page boundaries automatically when you reach the top or bottom of the current page.

## Why the TUI exists

The TUI is for discovery and incident-time reading. It complements the CLI rather than replacing it.
