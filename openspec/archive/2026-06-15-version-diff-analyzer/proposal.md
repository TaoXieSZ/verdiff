## Why

大型项目（如 prophase）在版本迭代时，理解两个 release 之间到底改了什么是高频且耗时的工作。当前做法是手动 `git diff` + 肉眼扫描，效率低且容易遗漏关键变化（如 cookbook 版本跳升、Policy runlist 变更、接口 breaking change）。之前我们用 Cursor Canvas 手工构建了 v0.30.1→v0.37.0 的差异分析，证明结构化版本差异分析的价值很大，但那是一次性的人工产出，无法复用。

需要一个通用工具，能对任意 git 项目的两个版本自动生成结构化差异分析报告，降低读懂 release 变更的门槛。

## What Changes

- 新增独立 CLI 工具 `verdiff`（version diff），接受 git repo + 两个版本标识（tag/branch/commit），输出 HTML 可视化报告
- 核心分析引擎基于 git diff，纯 git 通用，不假设项目结构
- 通过可插拔的 analyzer 机制支持项目特定的深度分析（如 Chef cookbook 版本提取、Policyfile 变化检测）
- HTML 报告包含：改动概览统计、文件热力图、目录级变化树、依赖版本变化矩阵、Breaking change 候选项、详细 diff 浏览
- CLI 同时支持纯文本摘要输出（适合 CI / terminal 场景）

## Capabilities

### New Capabilities
- `git-diff-engine`: 核心 git diff 解析引擎 — 文件级增删改统计、行级变化提取、目录结构变化树
- `dependency-tracker`: 依赖/组件版本变化追踪 — 从代码中提取版本号变化（go.mod, package.json, Gemfile, Policyfile.rb 等）
- `breaking-change-detector`: Breaking change 检测 — 基于 AST 或模式匹配识别公开接口变更、配置格式变化、行为变更
- `html-report-generator`: HTML 可视化报告生成 — 将分析结果渲染为交互式单文件 HTML 报告
- `analyzer-plugin`: 可插拔分析器框架 — 允许注册项目特定的分析器（如 Chef cookbook 版本、Terraform module 变化）

### Modified Capabilities

（无现有 capability 被修改）

## Impact

- 新增独立的 CLI 二进制，不影响 prophase 现有代码
- 依赖 git CLI 或 go-git 库
- HTML 报告为自包含单文件，无外部依赖
- analyzer plugin 接口需要稳定设计，后续可能被多个项目依赖
