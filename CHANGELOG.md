# Changelog

All notable changes to Glow will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- 一键安装脚本支持（Linux/macOS）
- 自我更新功能（`glow-server update` 和 `glow update`）
- 版本管理命令（`glow-server version` 和 `glow version`）
- 日志自动清理功能（按年龄 + 总量）
- 服务环境文件支持
- 固定目录约定（`/var/lib/glow-server`, `/etc/glow-server`）
- 一键卸载脚本（保留配置与数据库）
- 本地开发安装脚本（不常驻服务）

### Changed
- 移除 `glow-server install` 命令（统一使用安装脚本）
- 更新 `serve` 命令支持 `--data-dir` 参数
- 更新服务模板（systemd/launchd）支持环境文件

### Fixed
- 修复日志目录管理
- 修复服务注册的幂等性

---

## 版本号说明

- **[Unreleased]** - 即将发布的变更
- **[1.0.0]** - 第一个稳定版本（待发布）

### 变更类型

- **Added** - 新增功能
- **Changed** - 功能变更
- **Deprecated** - 即将废弃的功能
- **Removed** - 已移除的功能
- **Fixed** - Bug 修复
- **Security** - 安全性修复
