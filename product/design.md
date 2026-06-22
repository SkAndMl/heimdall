# Heimdall Design System

> **Purpose:** This document is the visual and interaction reference for every Heimdall surface: CLI output, Bubble Tea TUI screens, weekly reports, and any future desktop/web interface.
>
> **Product feeling:** Heimdall is a calm, premium, local-first developer utility. It should feel like a guardian of the user's storage—not a flashy cleanup app, an antivirus product, or a game-themed Viking tool.

---

## 1. Design North Star

Heimdall helps users understand and reclaim disk space with confidence.

The interface must communicate three things at all times:

1. **What Heimdall found** — clear categories, sizes, locations, and evidence.
2. **How safe an action is** — never blur “old” with “safe to remove.”
3. **What will happen next** — actions should be predictable, reversible where possible, and explained before execution.

### The intended personality

- Quiet, deliberate, authoritative
- Technical but not intimidating
- Precise rather than decorative
- Protective rather than aggressive
- Minimal visual noise; strong hierarchy

### Explicit non-goals

Do **not** make Heimdall look like:

- A generic “PC cleaner” with oversized gauges, gradients, or alarming red warnings
- A fantasy/Viking game UI with helmets, runes, axes, or ornamental chrome
- A dense monitoring dashboard where every metric has equal visual weight
- A bright developer terminal theme with many competing ANSI colours

---

## 2. Brand Concept

The Heimdall mark is a **guardian lens / eye**, represented in terminal contexts by a restrained circular glyph:

```text
◉ HEIMDALL
```

The mark should stand for observation, understanding, and protection. It may be used as:

- `◉` for scan complete, active monitoring, and primary status
- `○` for inactive, pending, or not-yet-scanned states
- `◌` for loading / scanning states when animation is available

Avoid logos that look like a broom, trash can, shield, Viking helmet, or security product.

---

## 3. Core Colour System

Heimdall defaults to **Midnight Guardian**: a dark navy environment with warm gold primary emphasis and cool cyan technical detail.

### 3.1 Base palette

| Token | Hex | Usage |
|---|---:|---|
| `bg` | `#0B1020` | App / terminal background |
| `surface` | `#121A2B` | Cards, panels, grouped sections |
| `surface-raised` | `#18243A` | Selected rows, hover/focus fill |
| `surface-strong` | `#202E47` | Strong active state, modal edge |
| `border` | `#26334B` | Quiet dividers and card borders |
| `border-focus` | `#69D9E7` | Keyboard focus indicator |
| `text-primary` | `#E8EEF8` | Primary headings and body text |
| `text-secondary` | `#B6C3D7` | Supporting text |
| `text-muted` | `#91A0B6` | Hints, timestamps, lower-priority metadata |

### 3.2 Semantic palette

| Token | Hex | Meaning | Typical use |
|---|---:|---|---|
| `primary-gold` | `#F2B950` | The user’s attention / primary action | selected checkbox, primary CTA, total reclaimable size |
| `info-cyan` | `#69D9E7` | System information / navigation | paths, IDs, links, scan progress, keyboard cues |
| `success-mint` | `#67D7A7` | Safe or completed | moved to Trash, cleanup succeeded, verified-safe |
| `warning-amber` | `#F5A524` | Requires review or has uncertainty | inactive project, unverified cache, confirmation needed |
| `danger-coral` | `#F27474` | Destructive or irreversible | permanent delete, failed cleanup, critical warning |

### 3.3 Colour rules

1. **Gold is scarce.** Use it for the primary action, selected state, and one key number per view.
2. **Cyan is informational, not promotional.** Use it for paths, navigational cues, links, and progress.
3. **Mint means a positive conclusion that Heimdall can defend.** Do not label an artifact “safe” merely because it is old.
4. **Amber means uncertainty, not danger.** Pair it with the evidence Heimdall lacks or the decision the user must make.
5. **Coral is reserved for irreversible actions and actual failures.** Do not use it for every large file or ordinary warning.
6. Never rely on colour alone. Pair semantic colours with words or glyphs: `Safe`, `Review`, `Permanent`, `Failed`.

---

## 4. Terminal Colour Fallbacks

Heimdall must remain readable in terminals that support only 16 or 256 colours.

| Semantic token | Preferred fallback |
|---|---|
| `primary-gold` | bright yellow |
| `info-cyan` | bright cyan |
| `success-mint` | bright green |
| `warning-amber` | yellow |
| `danger-coral` | bright red |
| `text-primary` | bright white |
| `text-muted` | bright black / gray |

### ANSI capability policy

- Detect colour support when practical.
- Honour `NO_COLOR`.
- Provide `--color=auto|always|never`.
- In monochrome mode, rely on labels, brackets, spacing, bold, and symbols—not colour.
- Never make an action inaccessible because the terminal lacks true colour support.

---

## 5. Typography and Text Hierarchy

Heimdall lives primarily in monospaced environments. Typography should be achieved through **weight, spacing, alignment, and restraint**, not elaborate font choices.

### Text levels

| Level | Terminal treatment | Example |
|---|---|---|
| Product label | bold + primary text | `◉ HEIMDALL` |
| Screen title | bold | `Cleanup candidates` |
| Major metric | bold + gold | `24.6 GB` |
| Category title | primary text | `Python environments` |
| Body / explanation | secondary text | `9 environments · 4 inactive for 90+ days` |
| Metadata / hint | muted text | `Last scanned 2 min ago` |
| Destructive warning | coral label + primary explanation | `Permanent deletion` |

### Copy rules

- Prefer concrete language: “Move to Trash” over “Optimize storage.”
- Use “potentially reclaimable” for estimated capacity, never “free space gained” before cleanup succeeds.
- Say what Heimdall knows and what it does not know.
- Write filesystem paths exactly; truncate visually only from the left where necessary.
- Avoid hype: no “turbo”, “deep clean”, “supercharge”, or “junk.”

Good:

```text
This environment is 2.1 GB and was last modified 167 days ago.
It can be recreated from requirements.txt.
```

Bad:

```text
Huge junk folder detected! Clean it now to boost your Mac.
```

---

## 6. Spacing, Layout, and Density

Heimdall should feel spacious enough to scan, but compact enough for a terminal.

### Spacing scale

| Token | Terminal cells | Use |
|---|---:|---|
| `space-1` | 1 | inline gaps, label-value separation |
| `space-2` | 2 | row indentation, small groups |
| `space-3` | 3 | section breathing room |
| `space-4` | 4 | major section separation |

### Layout principles

- Reserve the top line for product identity and disk status.
- Show a summary before details.
- Align sizes right so users can compare them quickly.
- Keep category rows to 1–2 lines whenever possible.
- Use borders to group meaningful information—not to box every line.
- Avoid nested boxes deeper than one level.
- Keep keyboard hints consistently at the bottom.

### Minimum terminal width

- Full TUI: target `>= 80 columns`.
- Comfortable TUI: target `>= 100 columns`.
- Below 80 columns: switch to compact single-column layout.
- Below 60 columns: show a concise warning and provide non-interactive CLI output rather than broken layout.

---

## 7. Iconography and Symbols

Use simple Unicode symbols that degrade gracefully. Do not depend on Nerd Fonts.

| Meaning | Primary symbol | ASCII fallback |
|---|---|---|
| Heimdall / observed | `◉` | `*` |
| Selected | `✓` | `x` |
| Unselected | `○` | ` ` |
| Safe | `✓` | `OK` |
| Review | `!` | `!` |
| Permanent / destructive | `×` | `X` |
| Info | `i` | `i` |
| Expand | `›` | `>` |
| Collapse | `⌄` | `v` |
| Folder | `▸` | `>` |
| Loading | `◌` / spinner | `...` |

### Symbol rules

- Symbols supplement text; they do not replace it.
- Use one symbol family consistently within a screen.
- Avoid emoji; their width and rendering vary across terminals.
- Avoid decorative icon clusters.

---

## 8. Interaction States

Every interactive row and action must have predictable states.

### 8.1 List row states

| State | Visual treatment |
|---|---|
| Default | primary title, muted metadata, no fill |
| Cursor / focused | `surface-raised` fill, cyan left rail or border |
| Selected | gold checkbox / indicator; retain readable text |
| Disabled | muted text plus reason visible on request |
| Warning | amber badge and short reason |
| Destructive | coral badge only when action is irreversible |

### 8.2 Button / action states

| Action tier | Treatment | Examples |
|---|---|---|
| Primary | gold emphasis | `Review cleanup plan`, `Move selected to Trash` |
| Secondary | cyan/outlined emphasis | `Inspect`, `Show evidence`, `Rescan` |
| Quiet | muted text | `Back`, `Skip`, `Quit` |
| Destructive | coral + explicit wording | `Delete permanently` |

### 8.3 Confirmations

Use a confirmation step when:

- An action permanently deletes anything.
- A native tool cleanup may remove multiple resource types.
- The selected total is substantial and contains medium/high-risk artifacts.

A confirmation screen must show:

1. Total estimated space affected
2. Number of artifacts
3. Exact cleanup method: Trash, native command, permanent deletion
4. At least one sentence about reversibility
5. The non-default safe choice

Example:

```text
Review cleanup plan

Selected: 14 artifacts · 9.7 GB
Method: Move to macOS Trash
Reversible: Yes, until Trash is emptied.

[Enter] Move to Trash     [Esc] Go back
```

Avoid irreversible actions as the default focused button.

---

## 9. Risk and Confidence Language

Heimdall must visually distinguish **classification confidence** from **cleanup risk**.

### Confidence

| Label | Meaning | Colour |
|---|---|---|
| `Verified` | Strong sentinel evidence / native tool confirms it | mint |
| `Likely` | Multiple strong signals, but not complete proof | cyan |
| `Possible` | Heuristic match only | amber |

### Cleanup risk

| Label | Meaning | Colour |
|---|---|---|
| `Low risk` | Re-creatable cache, installer, disposable build artifact | mint |
| `Review` | Likely removable but potentially meaningful to a project | amber |
| `High risk` | User data, unknown directory, or irreversible external state | coral |

Never collapse these labels into one ambiguous word like “safe.”

---

## 10. Screen Specifications

### 10.1 Home / scan summary

The first screen should answer: **How much space is involved and what should I do next?**

```text
◉ HEIMDALL                                      127.4 GB free

Disk intelligence for your machine
──────────────────────────────────────────────────────────────────

Potentially reclaimable                              24.6 GB
Safe to move to Trash                                 9.7 GB
Needs review                                         14.9 GB

┌────────────────────────────────────────────────────────────────┐
│ [✓] Python environments                               8.3 GB   │
│     9 environments · 4 inactive for 90+ days                    │
│                                                                │
│ [✓] node_modules                                     11.7 GB   │
│     14 directories · 6 detached from active projects            │
│                                                                │
│ [ ] Hugging Face cache                                10.4 GB  │
│     6 repositories · inspect before cleanup                     │
│                                                                │
│ [✓] Installers                                         3.1 GB  │
│     11 files · low-risk removal                                 │
└────────────────────────────────────────────────────────────────┘

↑↓ Navigate   Space Select   Enter Review plan   i Inspect   q Quit
```

Rules:

- `Potentially reclaimable` is the main gold metric.
- The “safe” and “review” metrics use mint and amber respectively.
- Category checkboxes indicate user selection, not automatic safety.
- Do not show every category if there are many; show top categories and a `View all` row.

### 10.2 Artifact detail / inspect view

The inspect view should build trust through evidence.

```text
‹ Back   Python environment                                  Review

old-api/.venv                                      2.1 GB
~/projects/old-api/.venv

Status
  Classification  Verified Python virtual environment
  Cleanup risk    Review
  Last modified   167 days ago
  Reclaim method  Move directory to Trash

Why Heimdall found this
  ✓ Contains pyvenv.cfg
  ✓ Contains bin/python
  ✓ Parent project contains requirements.txt
  ! Heimdall cannot confirm whether this project is still needed

[Space] Select for cleanup     [o] Open folder     [Esc] Back
```

Rules:

- Make `Why Heimdall found this` a first-class section.
- Do not bury the cleanup method below file listings.
- Show paths in cyan, but preserve good contrast.
- Do not show raw implementation jargon unless there is an “advanced details” view.

### 10.3 Cleanup review

```text
Review cleanup plan

14 artifacts selected                                  9.7 GB

Move to Trash
  Python environments             4 items              5.2 GB
  node_modules                    6 items              3.1 GB
  Installers                      4 items              1.4 GB

No permanent deletion is included.
You can restore these items from Trash until it is emptied.

[Enter] Move 14 items to Trash       [Esc] Adjust selection
```

Rules:

- The final action names the effect and quantity.
- Put reversibility near the action.
- Do not use alarming red for a reversible Trash operation.
- Use gold for the selected total and primary action.

### 10.4 Cleanup progress

```text
Moving selected items to Trash

██████████████████████░░░░░░░░░░  68%   6.6 / 9.7 GB

Current: ~/projects/demo/node_modules
Completed: 11 of 14 artifacts

[Esc] Stop after current item
```

Rules:

- Progress should be determinate when total bytes are known.
- Show the current path, but truncate from the left if necessary.
- A cancellation should stop safely after the current atomic operation; explain this in the UI.

### 10.5 Cleanup complete

```text
✓ Cleanup complete

Moved 14 artifacts to Trash                           9.7 GB
Your available disk space may update after the OS refreshes.

[Enter] Return to summary     [r] Rescan     q Quit
```

Rules:

- Use mint for success and gold only for the reclaimed amount.
- Do not promise exact free-space changes until the filesystem reports them.
- Offer rescan as a secondary action.

### 10.6 Empty state

```text
◉ HEIMDALL

No cleanup candidates found yet.

Run a scan to inspect developer artifacts, installers,
and caches on this machine.

[Enter] Scan now     q Quit
```

Avoid celebratory or judgemental language such as “Your disk is perfectly clean!”

### 10.7 Error state

```text
! Some locations could not be scanned

Heimdall could not read 3 directories because permission was denied.
Your scan results are still available, but totals may be incomplete.

[Enter] View results     [d] View details     q Quit
```

Errors should be clear, recoverable, and non-alarmist.

---

## 11. Reports and Non-Interactive CLI Output

Weekly reports should use the same hierarchy but not mimic the TUI excessively.

```text
HEIMDALL WEEKLY REPORT — Jun 15–22

Disk free: 127.4 GB   Change: -8.6 GB
Potentially reclaimable: 24.6 GB

Largest changes
  +5.1 GB   Hugging Face cache
  +2.0 GB   node_modules
  +1.3 GB   Docker build cache

Recommended review
  8.3 GB    Python environments (4 inactive for 90+ days)
  3.1 GB    Installers (11 files)

Run `heimdall clean` to review a reversible cleanup plan.
```

Rules:

- Lead with the trend, not every raw category.
- Use signs (`+`, `-`) for period deltas.
- Do not silently schedule or execute cleanup.
- Keep recommendations to three max in default output.

---

## 12. Component Contracts for Bubble Tea

All visual components should consume semantic design tokens rather than hardcoded hex values.

```go
// Example: conceptual token interface. Adjust to project architecture.
type Palette struct {
    Background  lipgloss.Color
    Surface     lipgloss.Color
    Raised      lipgloss.Color
    Border      lipgloss.Color
    FocusBorder lipgloss.Color

    TextPrimary lipgloss.Color
    TextMuted   lipgloss.Color

    Gold        lipgloss.Color
    Cyan        lipgloss.Color
    Mint        lipgloss.Color
    Amber       lipgloss.Color
    Coral       lipgloss.Color
}
```

Recommended component boundaries:

```text
app/
  screens/
    summary.go
    artifact_detail.go
    cleanup_review.go
    cleanup_progress.go
    result.go
  components/
    header.go
    metric.go
    category_row.go
    artifact_row.go
    status_badge.go
    evidence_list.go
    footer_help.go
    confirmation.go
  style/
    palette.go
    typography.go
    layout.go
    symbols.go
```

### Component rules

- Components receive state and render it; they should not decide filesystem policy.
- Do not embed selection logic inside styling code.
- Keep strings and key bindings centralized where practical.
- Every component must render coherently in monochrome mode.
- Put width-dependent layout behavior in one place instead of scattering it across views.

---

## 13. Accessibility and Terminal Resilience

Heimdall is a terminal product; accessibility means predictable keyboard behavior and information that survives stripped-down terminals.

Required:

- Full keyboard navigation for every action
- Visible cursor/focus state
- Colour-independent labels for status and risk
- No emoji-only controls
- Clear focus return after opening/closing a detail screen or modal
- No critical information in an animation alone
- Respect `NO_COLOR`
- Do not require mouse support
- Keep contrast high; muted text must remain readable against `bg`

Keyboard conventions:

| Key | Meaning |
|---|---|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Space` | Toggle selection |
| `Enter` | Primary contextual action |
| `i` | Inspect / explain |
| `r` | Rescan / refresh |
| `Esc` | Back / cancel current modal |
| `q` | Quit from a safe top-level state |
| `?` | Show keyboard help |

Never overload `Enter` with a destructive action without a prior review/confirmation screen.

---

## 14. Motion and Loading

Motion is optional and should indicate work, never decorate.

- Use a small spinner beside “Scanning” or “Calculating size.”
- Prefer determinate progress bars for cleanup.
- Do not animate large panels, gauges, or decorative particles.
- Avoid fast blinking or repeated alerts.
- Honor a future `--no-animation` setting.

Example:

```text
◌ Scanning ~/Library/Caches ...
```

---

## 15. Future Graphical UI Guidance

If Heimdall later gets a desktop/web interface, preserve the same hierarchy:

- Dark navy background
- Gold as a scarce primary emphasis
- Cyan for technical navigation and metadata
- A summary-first home screen
- Evidence-first inspection panels
- Reversible cleanup plans

Desktop-specific additions that are acceptable:

- A simple horizontal storage breakdown bar
- A timeline of weekly storage changes
- Search and filtering
- A small artifact “confidence / risk” legend

Avoid:

- Circular “disk health” gauges
- Giant percentage counters
- Multiple pie charts
- Decorative Norse illustrations
- Red/yellow/green traffic-light dashboards without textual labels

---

## 16. Quality Checklist for Every UI Change

Before merging a new interface or screen, verify:

### Hierarchy
- Is the most important metric obvious within two seconds?
- Is the next user action obvious?
- Are sizes consistently right-aligned and easy to compare?

### Safety
- Does the interface distinguish age, confidence, and cleanup risk?
- Does it state whether the action goes to Trash, uses a tool-native cleanup, or deletes permanently?
- Is permanent deletion opt-in and visually distinct?

### Consistency
- Are colours semantic and used according to this document?
- Is gold limited to selection/primary emphasis?
- Are paths cyan and metadata muted?
- Are focus, selected, warning, and disabled states all represented?

### Terminal resilience
- Does it work at 80 columns and degrade below that?
- Does it work with no colour?
- Are symbols accompanied by words where meaning matters?
- Are keyboard hints visible and correct?

### Copy
- Does it avoid hype, fear, and vague claims?
- Does it explain what Heimdall knows and does not know?
- Does it use “potentially reclaimable” before cleanup and not overpromise space recovery?

---

## 17. One-Sentence Design Rule

> Heimdall should make every cleanup decision feel **understood before it is executed**.
