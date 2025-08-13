# Docker 自动构建

本项目使用 GitHub Actions 自动构建 Docker 镜像并推送到 GitHub Container Registry。

## 快速开始

### 1. 启用 Actions
确保仓库设置中启用了 Actions 功能。

### 2. 配置权限
GitHub Actions 会自动使用 `GITHUB_TOKEN` 进行认证，无需额外配置。

### 3. 触发构建
推送代码到 `main` 分支或创建标签即可自动触发构建：

```bash
# 推送到主分支（构建 latest 标签）
git push origin main

# 创建版本标签（构建 v1.0.0 等标签）
git tag v1.0.0
git push origin v1.0.0
```

## 构建触发器

| 事件 | 触发条件 | 生成的标签 |
|------|----------|------------|
| 推送到 main/master | 分支名称 | `branch-main` |
| 推送到其他分支 | 分支名称 | `branch-<name>` |
| 创建标签 v* | 版本号 | `v1.0.0`, `v1.0`, `v1`, `latest` |
| Pull Request | PR 编号 | `pr-<number>` |
| 手动触发 | - | 根据上下文 |

## 镜像信息

- **注册表**: `ghcr.io`
- **镜像名**: `ghcr.io/<your-username>/<repository-name>`
- **架构**: `linux/amd64`, `linux/arm64`
- **基础镜像**: Alpine Linux

### 使用示例

```bash
# 拉取最新版本
docker pull ghcr.io/your-username/done-hub:latest

# 拉取特定版本
docker pull ghcr.io/your-username/done-hub:v1.0.0

# 运行容器
docker run -p 3000:3000 ghcr.io/your-username/done-hub:latest
```

## 开发工作流

### 1. 功能开发
```bash
# 创建功能分支
git checkout -b feature/enhancement

# 开发并提交
git add .
git commit -m "Add new feature"

# 推送并创建 PR
git push origin feature/enhancement
```

### 2. 版本发布
```bash
# 更新版本号
echo "v1.1.0" > VERSION

# 提交版本变更
git add VERSION
git commit -m "Bump version to v1.1.0"

# 创建标签
git tag v1.1.0

# 推送代码和标签
git push origin main
git push origin v1.1.0
```

## 监控和调试

### 构建状态
- 进入仓库的 Actions 标签页查看构建状态
- 每次构建都有详细的日志输出
- 构建失败会自动发送通知

### 本地测试
```bash
# 克隆仓库
git clone https://github.com/your-username/done-hub.git
cd done-hub

# 本地构建测试
docker build -t done-hub:test .

# 运行本地构建的镜像
docker run -p 3000:3000 done-hub:test
```

### 常见问题
1. **构建失败**: 检查代码语法错误和依赖项
2. **推送失败**: 确认 GitHub Token 权限足够
3. **镜像过大**: 检查 `.dockerignore` 文件配置

## 最佳实践

### 标签管理
- 使用语义化版本控制 (SemVer)
- 主版本号: `v1.0.0` (破坏性变更)
- 次版本号: `v1.1.0` (新功能)
- 修订版本号: `v1.0.1` (Bug 修复)

### 分支策略
- `main` / `master`: 生产就绪代码
- `develop`: 开发分支 (可选)
- `feature/*`: 功能开发
- `hotfix/*`: 紧急修复

### 安全考虑
- 定期更新基础镜像
- 使用 GitHub 的漏洞扫描功能
- 限制容器权限（非 root 用户运行）

## 支持和反馈

如果遇到问题：
1. 检查 [GitHub Issues](https://github.com/your-username/done-hub/issues)
2. 查看构建日志获取详细错误信息
3. 创建新的 Issue 报告问题

---

**注意**: 首次构建可能需要较长时间下载依赖项，后续构建会使用缓存加速。