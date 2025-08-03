# 智能化一键部署脚本使用指南

## 📋 概述

`scripts/interactive-deploy.sh` 是一个功能完整的智能化部署脚本，提供交互式界面和自动化部署功能，支持多种部署模式和环境配置。

## ✨ 主要特性

### 🎯 交互式功能
- **清晰的菜单选项**: 完整部署、仅Coze Studio、仅CozeRights、开发环境、自定义部署
- **智能默认值**: 所有配置项都有合理的默认值
- **超时输入**: 30秒超时自动使用默认值
- **实时进度条**: 显示部署进度和状态信息
- **用户确认**: 关键操作需要用户确认

### 🔧 系统兼容性
- **自动系统检测**: 支持 macOS 和 Linux (Ubuntu/CentOS/RHEL)
- **架构适配**: 支持 x86_64 和 ARM64 架构
- **包管理器适配**: brew (macOS)、apt (Ubuntu)、yum/dnf (CentOS/RHEL)
- **Shell兼容**: 支持 bash 和 zsh

### 🛡️ 资源检查与预安装
- **系统资源检查**: CPU (≥4核)、内存 (≥8GB)、磁盘 (≥50GB)
- **网络连接检查**: Docker Hub、GitHub、必要的API端点
- **自动依赖安装**: Docker、Docker Compose、Git、curl、jq
- **版本兼容性验证**: Docker ≥20.10，Docker Compose ≥2.0

### 🚀 部署内容
- **Coze Studio**: 后端服务、前端界面、MySQL、Redis、MinIO
- **CozeRights**: Go后端API、React管理界面、PostgreSQL、审计日志
- **集成组件**: 权限验证中间件、数据同步服务、API网关
- **监控套件**: Prometheus指标收集、Grafana可视化、告警配置

### 🔄 高级功能
- **配置文件自动生成**: 环境变量、数据库连接、API密钥
- **回滚机制**: 部署失败时自动清理和恢复
- **部署报告**: 服务状态、访问地址、账号信息
- **静默模式**: 无交互部署，使用预设配置
- **健康检查**: 完整的服务验证和状态检查

## 🚀 使用方法

### 基本使用

```bash
# 交互式部署（推荐）
./scripts/interactive-deploy.sh

# 静默模式部署
./scripts/interactive-deploy.sh --silent

# 指定部署模式
./scripts/interactive-deploy.sh --mode complete
```

### 命令行参数

```bash
# 显示帮助信息
./scripts/interactive-deploy.sh --help

# 静默模式，使用默认配置
./scripts/interactive-deploy.sh --silent

# 指定部署模式
./scripts/interactive-deploy.sh --mode [complete|coze|cozerights|dev|custom]

# 禁用回滚功能
./scripts/interactive-deploy.sh --no-rollback
```

### 部署模式说明

| 模式 | 说明 | 包含组件 |
|------|------|----------|
| `complete` | 完整部署（推荐） | Coze Studio + CozeRights + 监控套件 |
| `coze` | 仅部署 Coze Studio | Coze Studio + MySQL + Redis + MinIO |
| `cozerights` | 仅部署 CozeRights | CozeRights + PostgreSQL + Redis |
| `dev` | 开发环境部署 | 完整组件 + 调试配置 |
| `custom` | 自定义部署 | 用户选择组件 |

## 📋 部署流程

### 1. 系统检查阶段
- 检测操作系统和架构
- 验证系统资源（CPU、内存、磁盘）
- 检查网络连接
- 安装缺失的依赖
- 验证工具版本

### 2. 配置阶段
- 选择部署模式
- 配置端口和密码
- 生成安全密钥
- 创建目录结构
- 生成配置文件

### 3. 部署阶段
- 构建Docker镜像
- 启动基础设施服务
- 启动应用服务
- 启动监控服务
- 等待服务稳定

### 4. 验证阶段
- 执行健康检查
- 初始化系统数据
- 生成部署报告
- 显示访问信息

## 🔧 配置选项

### 默认端口配置
- **Coze Studio**: 8888
- **CozeRights**: 8080
- **Grafana**: 3000
- **Prometheus**: 9090

### 默认密码配置
- **MySQL**: coze123
- **PostgreSQL**: cozerights123
- **管理员账号**: admin/admin123

### 环境变量
脚本会自动生成 `.env` 文件，包含所有必要的环境变量：

```bash
# 安全配置
JWT_SECRET=<自动生成>
ENCRYPTION_KEY=<自动生成>

# 服务端口
COZE_PORT=8888
COZERIGHTS_PORT=8080

# 数据库密码
MYSQL_ROOT_PASSWORD=coze123
POSTGRES_PASSWORD=cozerights123

# 管理员配置
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
```

## 🛠️ 故障排查

### 常见问题

**1. 权限不足**
```bash
# 如果遇到权限问题，确保用户在docker组中
sudo usermod -aG docker $USER
# 重新登录或重启终端
```

**2. 端口冲突**
```bash
# 检查端口占用
netstat -tulpn | grep :8888
# 修改端口配置或停止冲突服务
```

**3. 内存不足**
```bash
# 检查内存使用
free -h
# 关闭不必要的服务或增加内存
```

**4. Docker服务未启动**
```bash
# 启动Docker服务
sudo systemctl start docker
# 设置开机自启
sudo systemctl enable docker
```

### 日志查看

```bash
# 查看部署日志
./scripts/interactive-deploy.sh 2>&1 | tee deployment.log

# 查看服务日志
docker-compose -f docker-compose.deploy.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.deploy.yml logs -f coze-server
```

### 回滚操作

```bash
# 手动回滚
docker-compose -f docker-compose.deploy.yml down -v
docker image prune -f
rm -f docker-compose.deploy.yml .env
```

## 📊 部署后操作

### 服务管理

```bash
# 查看服务状态
docker-compose -f docker-compose.deploy.yml ps

# 重启服务
docker-compose -f docker-compose.deploy.yml restart

# 停止服务
docker-compose -f docker-compose.deploy.yml down

# 查看资源使用
docker stats
```

### 访问服务

部署完成后，您可以通过以下地址访问服务：

- **Coze Studio**: http://localhost:8888
- **CozeRights 管理**: http://localhost:8080/admin
- **Grafana 监控**: http://localhost:3000
- **Prometheus**: http://localhost:9090

### 默认账号

- **用户名**: admin
- **密码**: admin123

## 🔒 安全建议

1. **修改默认密码**: 部署后立即修改默认管理员密码
2. **配置防火墙**: 只开放必要的端口
3. **启用HTTPS**: 在生产环境中配置SSL证书
4. **定期备份**: 设置自动备份策略
5. **监控告警**: 配置监控告警规则

## 📞 技术支持

如遇到问题，请：

1. 查看部署日志和错误信息
2. 参考故障排查章节
3. 查看相关文档
4. 提交Issue到GitHub仓库

---

**智能化部署脚本让企业级AI平台部署变得简单高效！** 🚀
