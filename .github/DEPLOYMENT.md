# Docker镜像构建说明

## GitHub Actions 自动构建

本项目已配置GitHub Actions来自动构建Docker镜像并推送到GitHub Container Registry (ghcr.io)。

### 触发条件
- **推送分支**: `main` 或 `master` 分支
- **标签推送**: 以 `v` 开头的标签 (如 `v1.0.0`)
- **Pull Request**: 用于验证构建
- **手动触发**: 可通过GitHub Actions界面手动触发

### 镜像标签策略
- **分支推送**: `branch-<branch-name>`
- **PR验证**: `pr-<number>`
- **语义版本**: `v1.0.0`, `v1.0`, `v1`
- **最新版本**: `latest` (仅默认分支)

### 多架构支持
- `linux/amd64` (x86_64)
- `linux/arm64` (ARM 64位)

## 手动构建

### 本地构建
```bash
# 构建镜像
docker build -t done-hub:latest .

# 构建指定版本
docker build --build-arg VERSION=v1.0.0 -t done-hub:v1.0.0 .

# 多架构构建
docker buildx build --platform linux/amd64,linux/arm64 -t done-hub:latest .
```

### 运行容器
```bash
# 基本运行
docker run -p 3000:3000 ghcr.io/your-username/done-hub:latest

# 挂载卷持久化数据
docker run -p 3000:3000 -v /path/to/data:/data ghcr.io/your-username/done-hub:latest

# 环境变量配置
docker run -p 3000:3000 \
  -e GITHUB_TOKEN=your_token \
  -e DATABASE_URL=your_db_url \
  ghcr.io/your-username/done-hub:latest
```

## 环境变量配置

### GitHub Actions Secrets
在仓库设置中配置以下secrets：

| 变量名 | 描述 | 必需 |
|--------|------|------|
| `GITHUB_TOKEN` | GitHub认证令牌 | ✅ |
| `DOCKER_REGISTRY` | 容器注册表 (默认ghcr.io) | ❌ |
| `REGISTRY_USERNAME` | 注册表用户名 | ❌ |
| `REGISTRY_PASSWORD` | 注册表密码 | ❌ |

### 容器运行时环境变量
根据done-hub配置，可能需要设置：

```bash
# 数据库配置
DATABASE_URL=postgresql://user:pass@host:port/dbname
REDIS_URL=redis://host:port

# API密钥
OPENAI_API_KEY=sk-xxx
ANTHROPIC_API_KEY=sk-ant-xxx

# 其他配置
PORT=3000
DEBUG=false
LOG_LEVEL=info
```

## 安全最佳实践

### GitHub Container Registry (ghcr.io)
1. **自动认证**: 使用GITHUB_TOKEN自动认证
2. **权限管理**: 通过仓库设置控制访问权限
3. **漏洞扫描**: GitHub自动进行安全扫描

### 镜像安全
- 使用多阶段构建减少镜像体积
- Alpine Linux基础镜像减少攻击面
- 非root用户运行 (如果需要)
- 定期更新基础镜像

### 构建优化
- 使用BuildKit缓存层
- 并行构建多架构镜像
- .dockerignore排除不必要文件
- 构建参数传递版本信息

## 监控和日志

### 构建监控
- GitHub Actions提供详细构建日志
- 构建失败会自动通知
- PR构建用于验证更改

### 运行监控
```bash
# 查看容器日志
docker logs <container-id>

# 实时查看日志
docker logs -f <container-id>

# 查看资源使用
docker stats <container-id>
```

## 故障排除

### 常见问题

#### 1. 构建失败
- 检查GitHub Actions日志
- 确认Dockerfile语法正确
- 验证依赖项是否完整

#### 2. 推送失败
- 确认GITHUB_TOKEN权限足够
- 检查容器注册表配额
- 验证镜像命名规范

#### 3. 运行问题
- 检查端口映射
- 验证环境变量配置
- 查看容器日志排查错误

### 调试命令
```bash
# 进入容器调试
docker exec -it <container-id> /bin/sh

# 检查镜像信息
docker inspect ghcr.io/your-username/done-hub:latest

# 清理构建缓存
docker builder prune
```

## 更新和维护

### 自动更新
- 推送代码到main分支自动触发构建
- 创建新标签自动创建新版本
- 定期检查基础镜像安全更新

### 手动更新
```bash
# 重新构建并推送
docker buildx build --platform linux/amd64,linux/arm64 \
  --push \
  -t ghcr.io/your-username/done-hub:latest \
  .
```