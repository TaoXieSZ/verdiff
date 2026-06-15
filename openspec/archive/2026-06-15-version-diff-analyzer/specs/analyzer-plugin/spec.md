## ADDED Requirements

### Requirement: Analyzer interface
The system SHALL define a Go interface that all analyzers (built-in and custom) MUST implement. The interface SHALL accept the diff result and return structured findings.

#### Scenario: Built-in analyzer registration
- **WHEN** the application starts
- **THEN** all built-in analyzers (file stats, dependency tracker, breaking change detector) are registered and active

### Requirement: Declarative custom analyzers via configuration
The system SHALL support user-defined analyzers in `.verdiff.yaml` that specify: file glob patterns to match, regex patterns to extract data, and output category name.

#### Scenario: Chef cookbook version analyzer
- **WHEN** user configures a custom analyzer matching `cookbooks/*/recipes/default.rb` with pattern `VERSION\s*=\s*["'](.+?)["']`
- **THEN** system extracts version values from matched files in both versions and reports changes

#### Scenario: Terraform module version analyzer
- **WHEN** user configures a custom analyzer matching `*.tf` with pattern `source\s*=.*\?ref=(.+?)"`
- **THEN** system tracks module version ref changes

### Requirement: Analyzer execution order
The system SHALL execute analyzers in dependency order: core diff analysis first, then dependency tracking, then breaking change detection, then custom analyzers. Each analyzer MAY access results from previously executed analyzers.

#### Scenario: Breaking change analyzer uses dependency data
- **WHEN** breaking change analyzer runs
- **THEN** it has access to the dependency tracker's output and can flag major version bumps as additional breaking change candidates

### Requirement: Analyzer result schema
Each analyzer SHALL return results conforming to a common schema: category (string), severity (info/warning/danger), items (list of findings), each item with a title, description, file path(s), and optional before/after values.

#### Scenario: Consistent result format
- **WHEN** multiple analyzers produce findings
- **THEN** all findings follow the same schema and can be rendered uniformly in the HTML report

### Requirement: Configuration file format
The `.verdiff.yaml` configuration file SHALL support:
- `analyzers`: list of custom analyzer definitions
- `version_patterns`: list of custom version extraction patterns
- `breaking_change_rules`: list of custom breaking change heuristics
- `ignore`: list of suppression rules
- `report`: report customization (theme default, diff line threshold)

#### Scenario: Full configuration example
- **WHEN** user creates `.verdiff.yaml` with custom analyzers and version patterns
- **THEN** system loads and applies all configurations, merging with built-in defaults
