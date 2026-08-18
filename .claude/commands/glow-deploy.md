---
description: Deploy the application to Glow using glow CLI

请使用 glow CLI 部署应用到服务器。

部署命令：
./scripts/deploy.sh              # 单应用自动部署
./scripts/deploy.sh <app_name>   # 指定应用部署

多应用项目：
./scripts/deploy.sh              # 交互式选择应用（或选择全部部署）

部署脚本会自动：
1. 扫描 cmd/ 目录检测应用
2. 自动构建应用（如需要）
3. 部署到 Glow 服务器

如果有错误，请检查：
- 应用代码是否正确
- glow 服务器连接是否正常
