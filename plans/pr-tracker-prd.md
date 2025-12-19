This is coming together perfectly. The logic you've described for "Stacked PRs" and the prioritization of "Reviewer > Team > Others" is exactly the kind of logic that transforms a tool from "useful" to "indispensable."

As your Product Manager, I have compiled these requirements into a concrete **Product Requirements Document (PRD)**.

Following the PRD, I have sketched the **Configuration Schema** and a **Technical Architecture** plan so we can start building immediately.

---

## 🏗️ Product Requirements Document: PRT (PR Tracker)

### 1. Product Overview

**PRT** is a CLI tool for macOS developers to visualize the status of GitHub Pull Requests across multiple local repositories. It solves "PR Fatigue" by aggregating data, highlighting actionable items (Reviewer Requests), and visualizing dependencies (Stacked PRs).

### 2. User Experience (CLI)

* **Command:** `prt`
* **Default Behavior:** Scans configured paths, detects git repos, fetches PRs via `gh`, and outputs a colorized report.
* **Arguments:**
* `--path`, `-p`: Override config paths (support globbing).
* `--filter`, `-f`: Filter by repo name (e.g., `aiml-*`).
* `--group`, `-g`: Override grouping (options: `project`, `author`).
* `--help`: Standard help.



### 3. Core Logic & Features

#### A. Data Gathering (Concurrency is Key)

* The tool must identify Git repositories within the provided paths (looking for `.git` folders).
* It must execute `gh pr list` (GitHub CLI) against these repos **in parallel** (using Go Goroutines) to ensure performance isn't bogged down by network latency.
* **Pre-requisite:** Check for `gh` installation. If missing, exit with: *"Error: GitHub CLI (gh) not found. Please run `brew install gh`."*

#### B. The "Stacked PR" Algorithm

* **Stack Detection:**
* If `PR_A` merges into `Branch_X`...
* And `PR_B` is coming *from* `Branch_X`...
* Then `PR_B` is a child of `PR_A`.


* **Visual:** Display as a tree.
* **Parent:** Full opacity/color.
* **Children (Dependent):** "Grayed out" (dimmed) to indicate they are blocked/waiting.



#### C. Prioritization & Filtering

The display logic follows this hierarchy of importance:

1. **High Priority:** PRs where **User is a Reviewer**. (Color: Green/Highlighted).
2. **Medium Priority:** PRs created by **Team Members** (defined in config). (Color: Standard/White).
3. **Low Priority:** PRs from bots/others. (Color: Dim/Gray or separate "Other" section).

#### D. Grouping & Sorting

* **Group By:** Configurable (Project vs. Author).
* **Default Sort:** Oldest -> Newest (to highlight stale items).
* **CI Status:** If `statusCheckRollup` is failing, append a red `✗` or similar icon.

### 4. Configuration System

* **Location:** `~/.prt/config.yaml`
* **First Run:** If file is missing, create it with commented-out defaults and print: *"Created new config at ~/.prt/config.yaml"*
* **Precedence:** CLI Flags > Config File > Hardcoded Defaults.

---

## ⚙️ Configuration Schema (Draft)

Here is how the `config.yaml` will look. This handles all the complexity you mentioned regarding "Team Members" vs "Others" and repository lists.

```yaml
# ~/.prt/config.yaml

# Your GitHub username (to identify "Reviewer" status)
# If left blank, we can try to auto-detect via `gh api user`
github_username: "jdoe"

# List of usernames for your immediate team
# PRs from these users get higher visibility than bots/outsiders
team_members:
  - "sarah-dev"
  - "mike-frontend"
  - "tom-backend"

# Where should PRT look for repositories?
# Can be exact paths or folders to recurse into
search_paths:
  - "/Users/jdoe/code/work"
  - "/Users/jdoe/code/personal/contrast-sdk"

# Glob patterns to MATCH (allow list)
# e.g., only show repos starting with 'aiml-' or 'contrast-'
# Leave empty to match all found git repos
include_repos:
  - "aiml-*"
  - "contrast-*"
  - "*-sdk"

# View Settings
default_group_by: "project" # options: project, author
default_sort: "oldest"      # options: oldest, newest

# Visuals
show_icons: true

```

---

## 🖥️ Visual Output Mockup (Refined)

Based on your feedback, here is the target output for `prt` in the terminal:

```text
PRT TRACKER ----------------------------------------------------

GROUP: PROJECT (contrast-sdk)
├── 📦 #402 Feature: Transformer Logic           (Waiting Review)
│   │  Created: 4d ago | Author: @sarah-dev | CI: ✅
│   │
│   └── 🔒 #405 Fix: Unit tests for transformer  (Blocked)
│          Created: 2h ago | Author: @sarah-dev | CI: ❌
│
└── 🟢 #399 Docs: Update Readme                  (REVIEW REQUESTED)
       Created: 5d ago | Author: @tom-backend | CI: ✅

GROUP: PROJECT (aiml-core)
└── ⚪ #102 Refactor: Python bindings            (Team Member)
       Created: 1d ago | Author: @mike-frontend | CI: ⏳

```

*(Note: The "Blocked" PR #405 is dimmed/gray in the actual terminal).*

---

## 🛠️ Technical Architecture (Go)

To execute this, we will use the following Go libraries to make it robust and beautiful:

1. **CLI Framework:** `github.com/spf13/cobra` (Industry standard for CLIs).
2. **Config:** `github.com/spf13/viper` (Handles the YAML and env overrides seamlessly).
3. **Styling:** `github.com/charmbracelet/lipgloss` (The best library for creating beautiful, colorful layouts in the terminal).
4. **Tree Rendering:** We will write a custom recursive render function to handle the "Stacked PR" logic using standard Unicode box-drawing characters (`├─`, `└─`).
