# verdiff

Analyze differences between two git versions and generate a structured report.

Given any git repository and two version identifiers (tags, branches, or commits), verdiff produces:

- **Change overview** — files changed, lines added/deleted, hotspot files
- **Directory tree** — collapsible tree showing change volume per directory
- **Dependency tracking** — version changes in go.mod, package.json, Gemfile.lock, Policyfile.rb, requirements.txt
- **Breaking change detection** — heuristic identification of removed exported functions, config keys, env vars, CLI flags
- **HTML report** — self-contained single-file report with dark/light theme, search, and interactive navigation
- **Text output** — plain-text summary for terminal and CI usage

## Install

```bash
go install github.com/txie/verdiff/cmd/verdiff@latest
```

Or build from source:

```bash
git clone https://github.com/txie/verdiff.git
cd verdiff
go build -o verdiff ./cmd/verdiff/
```

## Usage

```bash
# Basic usage — generates an HTML report
verdiff v1.0.0 v2.0.0

# Analyze a remote repo
verdiff --repo /path/to/repo v1.0.0 v2.0.0

# Focus on a specific directory
verdiff --path src/ v1.0.0 v2.0.0

# Text output for terminal/CI
verdiff --format text v1.0.0 v2.0.0

# Use git CLI backend (faster for large repos)
verdiff --use-git-cli v1.0.0 v2.0.0

# Custom output path
verdiff --output report.html v1.0.0 v2.0.0
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--repo` | `.` | Path to the git repository |
| `--output` | auto | Output file path |
| `--format` | `html` | Output format: `html` or `text` |
| `--path` | (all) | Limit diff to a specific directory |
| `--use-git-cli` | `false` | Use git CLI instead of go-git (faster for large repos) |
| `--top` | `10` | Number of hotspot files to highlight |
| `--config` | `.verdiff.yaml` | Path to configuration file |

## Configuration

Create a `.verdiff.yaml` in your repo root for project-specific analysis:

```yaml
# Custom version extraction patterns
version_patterns:
  - name: cookbook-version
    file_glob: "*/recipes/default.rb"
    pattern: 'VERSION\s*=\s*[''"](?P<version>[^''"]+)[''"]'

# Custom breaking change rules
breaking_change_rules:
  - id: nomad-task-removed
    name: Nomad task block removed
    file_glob: "*.nomad"
    pattern: 'task\s+"[^"]+"'

# Suppress known false positives
ignore:
  - path: "vendor/*"
  - path: "testdata/*"

# Report options
report:
  default_theme: auto   # auto, light, or dark
  diff_line_limit: 500
```

## Built-in Dependency Parsers

| File | What it extracts |
|---|---|
| `go.mod` | Go module dependencies |
| `package.json` | npm dependencies |
| `Gemfile.lock` | Ruby gem versions |
| `Policyfile.rb` | Chef cookbook pins |
| `Policyfile.lock.json` | Chef cookbook lock versions |
| `requirements.txt` | Python packages |
| `pyproject.toml` | Python packages |

## Built-in Breaking Change Rules

| Rule | Detects |
|---|---|
| Go exported function | Removed `func FuncName(` patterns |
| Go exported type | Removed `type TypeName struct/interface` |
| Config key removal | Top-level YAML/JSON keys deleted |
| Env var removal | Removed `os.Getenv("...")` references |
| CLI flag removal | Removed `flag.String/Int/Bool(` patterns |

## Architecture

```
cmd/verdiff/          CLI entry point
internal/verdiff/
  types.go            Core data structures
  analyzer.go         Analyzer interface + registry
  config.go           .verdiff.yaml config loading
  gitdiff/            Git diff engine (go-git + CLI fallback)
  deps/               Dependency version tracking
  breaking/           Breaking change detection
  report/             HTML + text report generation
```

The analyzer pipeline runs in order: git diff → dependency tracking → breaking change detection → custom analyzers. Each analyzer receives results from prior ones.

## License

MIT
