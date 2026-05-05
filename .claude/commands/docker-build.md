---
description: 运行 make docker-build 构建容器镜像，支持自定义 Tag。
allowed-tools: Bash(docker:*), Bash(make:docker-build)
---

1. 确定镜像 Tag：如果用户提供了参数 `$1`，则使用该值作为 Tag；否则默认使用 `latest`。
2. 执行 `make docker-build` 构建容器镜像。
3. 如果构建失败，分析错误输出，定位根本原因并给出修复建议。
