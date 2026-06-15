## Context

我们在 prophase v0.30.1→v0.37.0 升级时，手工用 Cursor Canvas 构建了一份结构化差异分析，包括 cookbook 版本变化矩阵、Policy runlist 变更、breaking change 识别。这份产出极有价值，但完全是手工的——每次升级都要重来一遍。

当前没有通用工具能做到：给定两个 git 版本，自动输出「改了什么 + 影响多大 + 哪些东西可能 break」的结构化报告。现有工具要么太底层（git diff / diffstat），要么太重（GitHub Release Notes 只有 PR 标题）。

目标用户：需要理解大型项目版本变更的工程师（升级执行者、code reviewer、on-call 工程师）。

## Goals / Non-Goals

**Goals:**
- 任意 git 项目，给两个版本标识，生成结构化差异分析 HTML 报告
- 分析维度覆盖：文件变更统计、依赖版本变化、配置/policy 变化、breaking change 候选
- 通过 analyzer plugin 机制支持项目特定的深度分析
- CLI 同时支持纯文本摘要（terminal / CI 友好）
- 单二进制，零配置即可对任何 git repo 使用

**Non-Goals:**
- 不做实时监控或持续集成集成（first version）
- 不做语义级代码分析（AST parsing 仅在 plugin 层可选）
- 不替代 changelog 或 release notes 的人工撰写
- 不做 git GUI 或 diff viewer

## Decisions

### D1: 语言选择 — Go

**选择**: Go  
**理由**: 与 prophase 技术栈一致；单二进制分发；go-git 库成熟；并发分析天然适合 goroutine。  
**备选**: Python（原型快但分发不便）、Rust（学习成本高且团队不熟）。

### D2: Git 交互 — go-git 库 + git CLI fallback

**选择**: 优先使用 `go-git`（纯 Go 实现），大仓库或特殊操作 fallback 到 `git` CLI。  
**理由**: go-git 避免外部依赖，但对超大仓库的 diff 性能可能不如原生 git。  
**备选**: 纯 git CLI subprocess（简单但多了外部依赖）、libgit2/git2go（CGO 编译复杂）。

### D3: 分析器架构 — Go interface + 内置 analyzers + 用户可扩展

**选择**: 定义 `Analyzer` interface，内置通用 analyzers（文件统计、依赖版本、breaking change），用户可通过配置文件注册自定义 pattern-based analyzers。  
**理由**: 保持核心通用，项目特定逻辑不污染主干。不用插件系统（Go plugin 跨平台差），改用声明式配置 + 内置模式匹配。  
**备选**: Go plugin（跨平台问题多）、外部进程 plugin（复杂度高、first version 不需要）。

### D4: 依赖版本检测 — 基于文件模式匹配 + 内置 parser

**选择**: 预定义常见依赖文件的 parser（go.mod, package.json, Gemfile.lock, Policyfile.lock.json, requirements.txt），用户可通过 `.verdiff.yaml` 添加自定义 version pattern。  
**理由**: 覆盖主流生态；自定义 pattern 支持 prophase 特有的 cookbook VERSION 常量等场景。  
**备选**: 仅正则匹配（不够精确）、每种生态单独写解析器（工作量大且封闭）。

### D5: HTML 报告 — 自包含单文件，Go html/template + 内嵌 CSS/JS

**选择**: Go 内置模板引擎渲染，CSS/JS 全部内联，零外部依赖的单 HTML 文件。  
**理由**: 一个文件就能发给同事、放进 artifact、浏览器直接打开。  
**备选**: React SPA（需要 node 构建链）、Markdown 报告（无法做交互式折叠和热力图）。

### D6: Breaking Change 检测 — 基于 heuristic pattern matching

**选择**: 通过一组可配置的 heuristic 规则识别 breaking change 候选（函数签名变化、配置 key 删除/重命名、API 路径变化等），标记为「候选」而非「确认」。  
**理由**: 精确的 breaking change 检测需要语义分析，成本高且语言相关。Heuristic 方式覆盖 80% 场景，误报由人 review。  
**备选**: 基于 AST 的精确分析（语言特定、first version 不做）。

## Risks / Trade-offs

- **[大仓库性能]** go-git 对超大仓库（10GB+）diff 可能慢 → 提供 `--use-git-cli` fallback flag，大仓库场景用原生 git
- **[依赖 parser 覆盖不全]** 内置 parser 无法覆盖所有生态 → 提供自定义 version pattern 配置；社区可贡献新 parser
- **[Breaking change 误报]** heuristic 检测会有 false positive → 在报告中标记为「候选」，不做确定性断言；支持 `.verdiff-ignore` 配置排除已知项
- **[HTML 报告体积]** 大量 diff 可能导致 HTML 文件过大 → 设置默认阈值，超限 diff 折叠或截断，提供 `--full` flag
- **[go-git 版本兼容]** go-git 可能不支持最新 git 特性 → fallback 机制兜底

## Open Questions

- 是否需要支持对比多个版本（如 v1→v2→v3 的趋势分析）？暂定 first version 只做两版对比。
- HTML 报告是否需要支持暗色/亮色主题切换？倾向支持，实现成本低。
