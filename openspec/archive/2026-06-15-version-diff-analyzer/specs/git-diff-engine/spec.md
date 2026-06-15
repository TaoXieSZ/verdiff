## ADDED Requirements

### Requirement: Parse git diff between two versions
The system SHALL accept two git version identifiers (tag, branch, or commit hash) and produce a structured representation of all file-level changes between them.

#### Scenario: Compare two tags
- **WHEN** user runs `verdiff v0.30.1 v0.37.0`
- **THEN** system outputs a structured diff covering all files changed between the two tags

#### Scenario: Compare branch to tag
- **WHEN** user runs `verdiff main v1.2.0`
- **THEN** system treats each identifier as a valid git ref and produces the diff

#### Scenario: Invalid ref
- **WHEN** user provides a ref that does not exist in the repository
- **THEN** system exits with a clear error message naming the invalid ref

### Requirement: File-level change statistics
The system SHALL compute per-file statistics: lines added, lines deleted, change type (added/modified/deleted/renamed).

#### Scenario: New file added between versions
- **WHEN** a file exists in version B but not in version A
- **THEN** system reports the file as "added" with total lines as lines added

#### Scenario: File deleted between versions
- **WHEN** a file exists in version A but not in version B
- **THEN** system reports the file as "deleted" with total lines as lines deleted

#### Scenario: File renamed
- **WHEN** a file is renamed between versions (detected by content similarity)
- **THEN** system reports it as "renamed" with old and new paths

### Requirement: Directory-level aggregation
The system SHALL aggregate file-level statistics into a directory tree structure, computing totals at each directory level.

#### Scenario: Nested directory aggregation
- **WHEN** files in `pkg/secrets/` and `pkg/terraform/` have changes
- **THEN** `pkg/` level shows the sum of all child changes, and each sub-directory shows its own totals

### Requirement: Change hotspot identification
The system SHALL identify the top N files and directories by change volume (lines added + deleted), marking them as hotspots.

#### Scenario: Top 10 hotspots
- **WHEN** analysis completes with default settings
- **THEN** the top 10 most-changed files are identified and ranked by total change volume

### Requirement: Git CLI fallback
The system SHALL fall back to invoking the `git` CLI when go-git cannot handle the repository (e.g., shallow clones, partial clones, or performance issues with very large repos).

#### Scenario: Large repository performance fallback
- **WHEN** user provides `--use-git-cli` flag
- **THEN** system uses `git diff --stat` and `git diff` subprocess instead of go-git
