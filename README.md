# vlt

Fast, standalone CLI for Obsidian vault operations. No Electron, no app dependency, no network calls. Just your vault and the filesystem.

```
vlt vault="MyVault" search query="architecture"
vlt vault="MyVault" backlinks file="Session Operating Mode"
vlt vault="MyVault" tags counts sort="count"
```

## Why vlt exists

Obsidian is a remarkable knowledge management tool. Its local-first philosophy, its plugin ecosystem, and the community around it have made it the go-to choice for millions of people who think in interlinked notes.

But Obsidian's official CLI requires the desktop app to be running. Every operation round-trips through Electron -- fine for interactive use, but a bottleneck when you need to script vault operations, run them in CI, integrate them into automated workflows, or use them from environments where a GUI simply isn't available.

**vlt** is a complementary tool that operates directly on your vault's markdown files. It reads the same Obsidian configuration, resolves notes the same way (including aliases), understands wikilinks, embeds, frontmatter, and tags -- but does it all through direct filesystem access.

Use cases where vlt shines:

- **AI agent workflows** -- LLM agents that read/write knowledge bases need fast, scriptable vault access without GUI dependencies
- **CI/CD pipelines** -- Validate link integrity, check for orphan notes, enforce tag conventions as part of your build
- **Shell scripting** -- Pipe vault content through standard Unix tools, batch-update properties, automate note creation
- **Remote/headless servers** -- Access your vault on machines where Obsidian can't run
- **Vault maintenance** -- Find orphan notes, broken links, and unresolved references across thousands of notes

vlt is not a replacement for Obsidian. It's a power tool for the command line that speaks the same language your vault already uses.

## Installation

### From source (requires Go 1.22+)

```bash
git clone https://github.com/paivot/vlt.git
cd vlt
make build     # produces ./vlt binary
make install   # installs to $GOPATH/bin
```

### Pre-built binaries

Check [Releases](https://github.com/paivot/vlt/releases) for pre-built binaries for macOS, Linux, and Windows.

## Quick start

```bash
# List your Obsidian vaults (discovered from Obsidian's config)
vlt vaults

# Read a note
vlt vault="MyVault" read file="Daily Note"

# Search by title and content
vlt vault="MyVault" search query="architecture"

# Create a note
vlt vault="MyVault" create name="New Idea" path="_inbox/New Idea.md" content="# New Idea"

# Pipe content from another command
echo "## Meeting Notes\n- Discussed roadmap" | vlt vault="MyVault" append file="New Idea"

# Find what links to a note
vlt vault="MyVault" backlinks file="Project Plan"

# Find broken links across the vault
vlt vault="MyVault" unresolved
```

### Setting a default vault

Instead of passing `vault=` every time, set an environment variable:

```bash
export VLT_VAULT="MyVault"
vlt search query="architecture"
```

If Obsidian's config file is unavailable (e.g., on a headless server), point directly to the vault path:

```bash
export VLT_VAULT_PATH="/path/to/my/vault"
export VLT_VAULT="MyVault"
vlt search query="architecture"
```

## Command reference

### File operations

| Command | Description |
|---------|-------------|
| `read file="<title>"` | Print note content (resolves by title or alias) |
| `create name="<title>" path="<path>" [content=...] [silent]` | Create a new note |
| `append file="<title>" [content="<text>"]` | Append content to end of note |
| `prepend file="<title>" [content="<text>"]` | Insert content after frontmatter |
| `move path="<from>" to="<to>"` | Move/rename note (auto-updates wikilinks and markdown links) |
| `delete file="<title>" [permanent]` | Move to .trash (or hard-delete) |
| `files [folder="<dir>"] [ext="<ext>"] [total]` | List vault files |
| `daily [date="YYYY-MM-DD"]` | Create or read daily note |

### Property (frontmatter) operations

| Command | Description |
|---------|-------------|
| `properties file="<title>"` | Show raw frontmatter block |
| `property:set file="<title>" name="<key>" value="<val>"` | Set or add a YAML property |
| `property:remove file="<title>" name="<key>"` | Remove a YAML property |

### Link operations

| Command | Description |
|---------|-------------|
| `backlinks file="<title>"` | Find notes linking to this note (includes embeds) |
| `links file="<title>"` | Show outgoing links (marks broken ones) |
| `orphans` | Find notes with no incoming links (alias-aware) |
| `unresolved` | Find all broken wikilinks across the vault |

### Tag operations

| Command | Description |
|---------|-------------|
| `tags [sort="count"] [counts]` | List all tags in vault |
| `tag tag="<tagname>"` | Find notes with tag or subtags |

### Task operations

| Command | Description |
|---------|-------------|
| `tasks [file="<title>"] [path="<dir>"] [done] [pending]` | List tasks (checkboxes) from one note or vault-wide |

### Search

| Command | Description |
|---------|-------------|
| `search query="<term> [key:value]"` | Search by title, content, and frontmatter properties |

### Other

| Command | Description |
|---------|-------------|
| `vaults` | List all discovered Obsidian vaults |
| `help` | Show usage information |
| `version` | Print version |

## Features in depth

### Vault discovery

vlt reads Obsidian's configuration to discover your vaults automatically:

| Platform | Config location |
|----------|----------------|
| macOS | `~/Library/Application Support/obsidian/obsidian.json` |
| Linux | `~/.config/obsidian/obsidian.json` |
| Windows | `%APPDATA%\obsidian\obsidian.json` |

You can reference a vault three ways:

```bash
vlt vault="MyVault" ...          # by name (directory basename from config)
vlt vault="/absolute/path" ...   # by absolute path
vlt vault="~/Documents/vault" ...# by home-relative path
```

### Note resolution

Notes are resolved by a two-pass algorithm:

1. **Fast pass** -- exact filename match (`<title>.md`), no file I/O needed
2. **Alias pass** -- if no filename match, scan frontmatter `aliases` for a case-insensitive match

This means you can reference notes by their aliases just like in Obsidian:

```yaml
---
aliases: [PKM, Personal Knowledge Management]
---
```

```bash
vlt vault="MyVault" read file="PKM"  # resolves via alias
```

### Wikilink support

vlt understands all standard Obsidian wikilink formats:

| Format | Example |
|--------|---------|
| Simple link | `[[Note Title]]` |
| Link to heading | `[[Note Title#Section]]` |
| Block reference | `[[Note Title#^block-id]]` |
| Display text | `[[Note Title\|Custom Text]]` |
| Heading + display | `[[Note Title#Section\|Custom Text]]` |
| Block ref + display | `[[Note Title#^block-id\|Custom Text]]` |
| Embed | `![[Note Title]]` |
| Embed with heading + display | `![[Note Title#Section\|Custom Text]]` |

When you rename a note with `move`, vlt automatically updates both wikilinks and markdown-style links across the vault:

```bash
vlt vault="MyVault" move path="drafts/Old Name.md" to="published/New Name.md"
# Output:
# moved: drafts/Old Name.md -> published/New Name.md
# updated [[Old Name]] -> [[New Name]] in 12 file(s)
# updated [...](drafts/Old Name.md) -> [...](published/New Name.md) in 3 file(s)
```

Link updates preserve headings, block references, display text, and embed prefixes. Markdown links have their relative paths recomputed correctly. If only the folder changes (same filename), wikilink updates are skipped since Obsidian resolves by title regardless of path, but markdown links are always updated since they use paths.

### Tag support

vlt collects tags from two sources, just like Obsidian:

**Frontmatter tags:**
```yaml
---
tags: [project, backend]
---
```

**Inline tags:**
```markdown
This is about #architecture and #design/patterns.
```

Tags are case-insensitive and deduplicated. Hierarchical tags support subtag matching:

```bash
vlt vault="MyVault" tag tag="design"
# Finds notes with #design, #design/patterns, #design/ux, etc.
```

### Stdin support

`create`, `append`, and `prepend` accept content from stdin when `content=` is omitted. This makes vlt composable with other Unix tools:

```bash
# Pipe output from another command
date | vlt vault="MyVault" append file="Daily Log"

# Use heredoc for multi-line content
vlt vault="MyVault" create name="Meeting" path="_inbox/Meeting.md" <<'EOF'
---
type: meeting
date: 2025-01-15
---
# Team Sync
- Discussed roadmap priorities
EOF
```

### Output formats

Most listing commands support `--json`, `--yaml`, and `--csv` output for programmatic consumption:

```bash
# JSON output for scripts
vlt vault="MyVault" orphans --json
# ["_inbox/Stale Note.md","drafts/Abandoned.md"]

# CSV for spreadsheets
vlt vault="MyVault" tags counts --csv
# tag,count
# project,15
# architecture,8

# YAML for config files
vlt vault="MyVault" search query="paivot" --yaml
# - title: Paivot Architecture
#   path: decisions/Paivot Architecture.md
```

### Property-based search

Search queries can include `[key:value]` filters to match frontmatter properties:

```bash
# Find all active decisions
vlt vault="MyVault" search query="[status:active] [type:decision]"

# Text + property filter
vlt vault="MyVault" search query="architecture [status:active]"

# Property filter only (no text search)
vlt vault="MyVault" search query="[type:pattern]"
```

### Task parsing

vlt parses `- [ ]` and `- [x]` checkboxes from notes:

```bash
# All tasks across the vault
vlt vault="MyVault" tasks

# Tasks from a specific note
vlt vault="MyVault" tasks file="Project Plan"

# Only pending tasks in a folder
vlt vault="MyVault" tasks path="projects" pending

# JSON output for programmatic use
vlt vault="MyVault" tasks --json
```

### Daily notes

Create or read daily notes following Obsidian's daily note conventions:

```bash
# Today's note (creates if missing, prints if exists)
vlt vault="MyVault" daily

# Specific date
vlt vault="MyVault" daily date="2025-01-15"
```

vlt reads configuration from `.obsidian/daily-notes.json` or `.obsidian/plugins/periodic-notes/data.json`, supporting custom folders, date formats (Moment.js tokens translated to Go), and templates with `{{date}}` and `{{title}}` variables.

### Output conventions

vlt follows Unix conventions for composability:

- One result per line (easy to pipe to `wc -l`, `grep`, `sort`, etc.)
- Relative paths from vault root
- Silent on empty results (exit code 0, no output -- like `grep`)
- Errors go to stderr with `vlt:` prefix
- Tab-separated fields where applicable (e.g., `tags counts`)

```bash
# Count orphan notes
vlt vault="MyVault" orphans | wc -l

# Find broken links in a specific folder
vlt vault="MyVault" unresolved | grep "^methodology/"

# List top 10 tags by frequency
vlt vault="MyVault" tags counts sort="count" | head -10
```

## Comparison with Obsidian CLI

vlt is a drop-in complement for the official [Obsidian CLI](https://github.com/Obsidian-CLI/obsidian-cli). The parameter syntax is intentionally compatible (`key="value"` style) to make migration straightforward.

| Capability | vlt | Obsidian CLI |
|------------|-----|--------------|
| read | Yes | Yes |
| search (with property filters) | Yes | Yes (no filters) |
| create | Yes | Yes |
| append | Yes | Yes |
| prepend | Yes | Yes |
| move (wiki + markdown link repair) | Yes | Yes (wiki only) |
| delete (trash + permanent) | Yes | Yes |
| files | Yes | Yes |
| daily notes | Yes | No |
| tasks | Yes | No |
| properties | Yes | Yes |
| property:set | Yes | Yes |
| property:remove | Yes | Yes |
| backlinks | Yes | Yes |
| links | Yes | Yes |
| orphans | Yes | Yes |
| unresolved | Yes | Yes |
| tags (list + counts) | Yes | Yes |
| tag (search + hierarchical) | Yes | Yes |
| Alias resolution | Yes | Yes |
| Block references `#^block-id` | Yes | Yes |
| Embed `![[...]]` support | Yes | Yes |
| Output formats (JSON/CSV/YAML) | Yes | No |
| Requires Obsidian running | **No** | Yes |
| External dependencies | **None** | Node.js |

## Architecture

vlt is a single-package Go binary with zero external dependencies. The entire tool runs on Go's standard library.

```
main.go          CLI entry point, argument parsing, command dispatch
vault.go         Vault discovery from Obsidian config, note resolution
commands.go      Command implementations (read, search, create, move, etc.)
wikilinks.go     Wikilink/embed parsing, replacement, markdown link repair
frontmatter.go   YAML frontmatter extraction and manipulation
tags.go          Inline tag parsing and tag-based queries
format.go        Output formatting (JSON, CSV, YAML, plain text)
tasks.go         Task/checkbox parsing and queries
daily.go         Daily note creation and config loading
```

**Design choices:**

- **Zero dependencies** -- The `go.mod` has no `require` lines. This eliminates supply chain risk and keeps the binary small and fast to compile.
- **Direct filesystem access** -- All operations read and write files directly. No database, no index, no daemon.
- **Two-pass note resolution** -- Filename match first (no I/O), then alias scan (reads frontmatter). Fast for the common case, correct for the edge case.
- **Case-insensitive link matching** -- Mirrors Obsidian's behavior. `[[my note]]` resolves to `My Note.md`.
- **Simple frontmatter parsing** -- String-based YAML parsing handles Obsidian's common patterns (key-value, inline lists, block lists) without pulling in a full YAML library.

### Stats

| Metric | Value |
|--------|-------|
| Lines of code | ~2,500 (source) |
| Lines of tests | ~2,450 |
| Test cases | 155 |
| Test coverage | 71% |
| External dependencies | 0 |
| Go version | 1.22+ |

## Development

```bash
make build    # compile
make test     # run tests (verbose)
make install  # install to $GOPATH/bin
make clean    # remove build artifacts
```

### Running tests

```bash
go test -v ./...             # verbose output
go test -cover ./...         # with coverage
go test -run TestCmdMove ./... # run specific test
```

All tests use `t.TempDir()` for isolated vault environments. No mocks -- every test creates real files and exercises real filesystem operations.

### Adding a new command

1. Add the command name to `knownCommands` in `main.go`
2. Implement `cmdYourCommand(vaultDir string, params map[string]string) error` in `commands.go`
3. Add the dispatch case in `main()` switch
4. Add usage line and examples in `usage()`
5. Write tests in `main_test.go`

## Contributing

Contributions are welcome. Please:

1. Open an issue describing the feature or bug before submitting a PR
2. Include tests for any new functionality
3. Keep the zero-dependency constraint -- no external modules
4. Follow the existing code style (simple, direct, no abstractions for one-off operations)
5. Run `make test` before submitting

## Roadmap

### Indexed full-text search (tantivy)

The current `search` command is a linear scan -- it reads every `.md` file in the vault on each query. For human-scale vaults (a few thousand notes) this is fast enough thanks to OS page cache. But vlt was built with AI agents in mind, and agents doing proper zettelkasten produce vaults that grow far beyond what a human would maintain by hand.

When demand warrants it, we plan to integrate [tantivy](https://github.com/quickwit-oss/tantivy) (the Rust full-text search engine that powers Quickwit and Meilisearch) to provide:

- Persistent inverted index with incremental updates
- Sub-millisecond search across arbitrarily large vaults
- Relevance-ranked results
- Fuzzy matching and phrase queries

This will be an opt-in feature -- the zero-dependency linear scan remains the default for simplicity. If this matters to you, open an issue or upvote an existing one.

### Recently shipped (v0.4.0)

- Block references (`[[Note#^block-id]]`) -- full support in parsing, rename, and backlinks
- Markdown link `[text](path.md)` repair on move -- relative paths recomputed correctly
- Property-based search filters (`search query="[status:active] [type:decision]"`)
- Output format flags (`--json`, `--yaml`, `--csv`) for all listing commands
- Daily note commands with Obsidian config support and templates
- Task/checkbox parsing with done/pending filters and vault-wide search

## License

Apache License 2.0. See [LICENSE](LICENSE) for full text.

Copyright 2025 Paivot Contributors.
