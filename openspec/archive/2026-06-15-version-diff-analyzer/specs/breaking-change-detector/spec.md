## ADDED Requirements

### Requirement: Heuristic breaking change detection
The system SHALL apply a set of configurable heuristic rules to identify candidate breaking changes in the diff. All detections SHALL be marked as "candidate" (not confirmed) to acknowledge potential false positives.

#### Scenario: Exported function signature changed
- **WHEN** a public/exported function's parameter list or return type changes between versions
- **THEN** system flags it as a breaking change candidate with the old and new signatures

#### Scenario: Configuration key removed
- **WHEN** a key present in a configuration file (YAML, JSON, TOML) in version A is absent in version B
- **THEN** system flags it as a breaking change candidate

#### Scenario: No breaking changes detected
- **WHEN** heuristic rules find no matches
- **THEN** the breaking changes section shows "No breaking change candidates detected"

### Requirement: Built-in heuristic rules
The system SHALL include built-in rules for common breaking change patterns:
- Function/method signature changes in Go, Python, JavaScript/TypeScript
- Removed or renamed configuration keys (YAML, JSON, TOML)
- Deleted public API endpoints (by pattern matching route definitions)
- Environment variable removals
- CLI flag removals

#### Scenario: Go exported function removed
- **WHEN** a Go file in version A exports `func FetchData(...)` and version B removes it entirely
- **THEN** system reports "Removed exported function: FetchData" as a breaking change candidate

### Requirement: Custom breaking change rules
The system SHALL support user-defined breaking change rules in `.verdiff.yaml` using file glob + regex pattern pairs.

#### Scenario: Custom rule for Nomad job spec
- **WHEN** user defines a rule matching `*.nomad` files for removed `task` blocks
- **THEN** system applies this rule during analysis and reports matches

### Requirement: Ignore list support
The system SHALL support a `.verdiff-ignore` file (or section in `.verdiff.yaml`) to suppress known false positives by file path pattern or rule ID.

#### Scenario: Suppressed false positive
- **WHEN** a detected candidate matches an ignore rule
- **THEN** it is excluded from the report (or shown in a collapsed "suppressed" section)
