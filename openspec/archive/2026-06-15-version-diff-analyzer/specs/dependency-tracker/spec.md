## ADDED Requirements

### Requirement: Detect dependency file changes
The system SHALL identify files that declare dependencies or component versions by matching known file patterns (go.mod, go.sum, package.json, Gemfile, Gemfile.lock, Policyfile.rb, Policyfile.lock.json, requirements.txt, pyproject.toml, Cargo.toml).

#### Scenario: go.mod changed between versions
- **WHEN** `go.mod` has changes between version A and version B
- **THEN** system parses both versions and reports added, removed, and upgraded/downgraded dependencies with old and new versions

#### Scenario: No dependency files changed
- **WHEN** no recognized dependency files have changes
- **THEN** the dependency section of the report shows "No dependency changes detected"

### Requirement: Version diff matrix
The system SHALL produce a matrix showing each dependency/component with its version in A and version in B, plus the change direction (upgrade/downgrade/added/removed).

#### Scenario: Multiple dependency files
- **WHEN** both `go.mod` and a `Policyfile.rb` have changes
- **THEN** the matrix groups changes by dependency file type and shows all version transitions

### Requirement: Custom version pattern support
The system SHALL support user-defined version extraction patterns via `.verdiff.yaml` configuration, allowing extraction of version numbers from arbitrary files using regex with named capture groups.

#### Scenario: Custom cookbook VERSION constant
- **WHEN** user configures a pattern to extract `VERSION = "x.y.z"` from Ruby files
- **THEN** system detects version changes in matching files and includes them in the dependency matrix

#### Scenario: No configuration file
- **WHEN** no `.verdiff.yaml` exists in the repository
- **THEN** system uses only built-in parsers and does not error

### Requirement: Semantic version comparison
The system SHALL parse versions as semantic versions when possible and indicate whether a change is a major, minor, or patch bump.

#### Scenario: Major version bump
- **WHEN** a dependency changes from `1.4.3` to `2.0.0`
- **THEN** system flags this as a major version bump with a warning indicator
