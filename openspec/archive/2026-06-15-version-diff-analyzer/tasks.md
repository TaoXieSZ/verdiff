## 1. 项目骨架

- [x] 1.1 创建 `cmd/verdiff/` 目录，初始化 main.go 和 CLI flag 解析（cobra 或 flag 包）
- [x] 1.2 定义核心数据结构：DiffResult, FileDiff, DirNode, AnalyzerFinding, VersionChange
- [x] 1.3 定义 Analyzer interface：`Analyze(ctx, DiffResult) ([]Finding, error)`
- [x] 1.4 创建 `.verdiff.yaml` 配置文件结构和加载逻辑

## 2. Git Diff 引擎

- [x] 2.1 实现 go-git 版本解析：接受 tag/branch/commit，resolve 到 commit object
- [x] 2.2 实现 go-git diff：两个 commit 之间的文件级 diff（增删改重命名）
- [x] 2.3 实现行级统计：per-file lines added/deleted
- [x] 2.4 实现目录树聚合：从文件级统计构建 DirNode 树，每层累加
- [x] 2.5 实现 hotspot 排序：按 total change volume 排序，取 top N
- [x] 2.6 实现 `--use-git-cli` fallback：调 git CLI subprocess 做 diff
- [x] 2.7 写 git-diff-engine 单元测试（用 test fixture repo）

## 3. 依赖版本追踪

- [x] 3.1 实现 go.mod parser：提取 require 块的 module→version 映射
- [x] 3.2 实现 package.json parser：提取 dependencies/devDependencies
- [x] 3.3 实现 Gemfile.lock parser：提取 gem→version
- [x] 3.4 实现 Policyfile.rb / Policyfile.lock.json parser
- [x] 3.5 实现 requirements.txt / pyproject.toml parser
- [x] 3.6 实现版本 diff 矩阵：对比两版的解析结果，输出 added/removed/upgraded/downgraded
- [x] 3.7 实现 semver 解析和 major/minor/patch 判断
- [x] 3.8 实现自定义 version pattern（从 .verdiff.yaml 加载 regex + file glob）
- [x] 3.9 写 dependency-tracker 单元测试

## 4. Breaking Change 检测

- [x] 4.1 实现 heuristic 规则引擎框架：rule 定义 + matcher + reporter
- [x] 4.2 实现内置规则：Go 导出函数签名变化（正则匹配 `func [A-Z]`）
- [x] 4.3 实现内置规则：配置 key 删除（YAML/JSON/TOML key diff）
- [x] 4.4 实现内置规则：环境变量引用删除
- [x] 4.5 实现内置规则：CLI flag 删除（flag.String/cobra 命令定义变化）
- [x] 4.6 实现自定义规则加载（从 .verdiff.yaml 的 breaking_change_rules）
- [x] 4.7 实现 .verdiff-ignore 抑制逻辑
- [x] 4.8 写 breaking-change-detector 单元测试

## 5. Analyzer Plugin 框架

- [x] 5.1 实现 analyzer 注册表：Register() + 按依赖顺序执行
- [x] 5.2 将 git-diff-engine、dependency-tracker、breaking-change-detector 注册为内置 analyzer
- [x] 5.3 实现声明式自定义 analyzer：从 .verdiff.yaml 加载 file glob + regex pattern → Finding
- [x] 5.4 实现 analyzer 间数据传递：后续 analyzer 可读取前序结果
- [x] 5.5 写 analyzer-plugin 集成测试

## 6. HTML 报告生成

- [x] 6.1 设计 HTML 报告布局和组件（header、dashboard、directory tree、version matrix、breaking changes、file detail）
- [x] 6.2 实现 Go html/template 模板，CSS/JS 全部内联
- [x] 6.3 实现 change overview dashboard（统计卡片：文件数、行数、hotspot 数）
- [x] 6.4 实现目录级变化树（可折叠展开，带 change volume 色条）
- [x] 6.5 实现依赖版本变化矩阵表格
- [x] 6.6 实现 breaking change 候选列表（severity 标记 + 文件路径链接）
- [x] 6.7 实现文件级 inline diff（默认折叠，超阈值截断 + show more）
- [x] 6.8 实现文件搜索/过滤（client-side JS）
- [x] 6.9 实现亮色/暗色主题切换（prefers-color-scheme + toggle）
- [x] 6.10 实现 `--format text` 纯文本摘要输出
- [x] 6.11 写 HTML 报告生成测试（验证生成的 HTML 包含所有必要 section）

## 7. CLI 集成与收尾

- [x] 7.1 实现完整 CLI 流程：parse args → open repo → diff → run analyzers → generate report
- [x] 7.2 实现 `--output` flag 指定输出文件路径（默认 `verdiff-{repoName}-{vA}-{vB}.html`）
- [x] 7.3 实现 `--config` flag 指定 .verdiff.yaml 路径
- [x] 7.4 实现进度输出（stderr，大仓库时让用户知道在跑）
- [x] 7.5 用 prophase v0.30.1→v0.37.0 做端到端验证，对比之前手工 canvas 的结论
- [x] 7.6 写 README.md：安装、用法、配置示例、截图
