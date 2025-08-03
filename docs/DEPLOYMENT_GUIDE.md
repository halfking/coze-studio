# CozeRights + Coze Studio 部署指南

## 📋 部署概述

本指南将帮助您在生产环境中部署 CozeRights 权限管理系统与 Coze Studio 的完整集成方案。

## 🔧 系统要求

### 硬件要求
- **CPU**: 4核心以上
- **内存**: 8GB以上（推荐16GB）
- **存储**: 50GB以上可用空间
- **网络**: 稳定的互联网连接

### 软件要求
- **操作系统**: Linux (Ubuntu 20.04+, CentOS 8+) 或 macOS
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **Git**: 2.0+

## 🚀 快速部署

### 1. 获取源码
```bash
# 克隆仓库
git clone https://github.com/coze-dev/coze-studio.git
cd coze-studio

# 检查CozeRights集成组件
ls -la backend/internal/
```

### 2. 环境配置
```bash
# 复制环境配置模板
cp docker/.env.example docker/.env

# 编辑配置文件
vim docker/.env
```

### 3. 模型配置
```bash
# 复制模型配置模板
cp backend/conf/model/template/model_template_ark_doubao-seed-1.6.yaml \
   backend/conf/model/ark_doubao-seed-1.6.yaml

# 配置模型参数
vim backend/conf/model/ark_doubao-seed-1.6.yaml
```

### 4. 启动服务
```bash
# 构建并启动所有服务
cd docker
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 5. 验证部署
```bash
# 检查Coze Studio
curl http://localhost:8888/health

# 检查CozeRights
curl http://localhost:8080/health

# 访问Web界面
open http://localhost:8888
```

## 🔐 安全配置

### SSL/TLS配置
```bash
# 生成SSL证书
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/private.key -out ssl/certificate.crt

# 更新docker-compose配置
vim docker-compose.yml
```

### 防火墙配置
```bash
# 开放必要端口
sudo ufw allow 8888/tcp  # Coze Studio
sudo ufw allow 8080/tcp  # CozeRights
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

## 📊 监控配置

### Prometheus监控
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'coze-studio'
    static_configs:
      - targets: ['localhost:8888']
  
  - job_name: 'cozerights'
    static_configs:
      - targets: ['localhost:8080']
```

### Grafana仪表板
```bash
# 导入预配置的仪表板
curl -X POST http://admin:admin@localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @monitoring/grafana-dashboard.json
```

## 🔄 数据备份

### 自动备份脚本
```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR

# 备份MySQL数据库
docker-compose exec mysql mysqldump -u root -p$MYSQL_ROOT_PASSWORD opencoze > $BACKUP_DIR/coze.sql

# 备份PostgreSQL数据库
docker-compose exec postgres pg_dump -U postgres cozerights > $BACKUP_DIR/cozerights.sql

# 备份配置文件
tar -czf $BACKUP_DIR/configs.tar.gz backend/conf/ docker/.env

echo "备份完成: $BACKUP_DIR"
```

## 🚨 故障排查

### 常见问题

**1. 服务启动失败**
```bash
# 查看日志
docker-compose logs coze-server
docker-compose logs cozerights-server

# 检查端口占用
netstat -tulpn | grep :8888
```

**2. 数据库连接失败**
```bash
# 检查数据库状态
docker-compose exec mysql mysql -u root -p -e "SHOW DATABASES;"
docker-compose exec postgres psql -U postgres -l
```

**3. 权限验证失败**
```bash
# 检查CozeRights服务
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

## 📈 性能优化

### 数据库优化
```sql
-- MySQL优化
SET GLOBAL innodb_buffer_pool_size = 2G;
SET GLOBAL max_connections = 1000;

-- PostgreSQL优化
ALTER SYSTEM SET shared_buffers = '2GB';
ALTER SYSTEM SET max_connections = 1000;
```

### Redis缓存优化
```bash
# 配置Redis内存限制
docker-compose exec redis redis-cli CONFIG SET maxmemory 1gb
docker-compose exec redis redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

## 🔧 维护操作

### 日常维护
```bash
# 清理Docker资源
docker system prune -f

# 更新镜像
docker-compose pull
docker-compose up -d

# 查看资源使用
docker stats
```

### 日志管理
```bash
# 配置日志轮转
cat > /etc/logrotate.d/coze-studio << EOF
/var/log/coze-studio/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 root root
}
EOF
```

## 📞 技术支持

如遇到部署问题，请：

1. 查看详细日志：`docker-compose logs -f`
2. 检查系统资源：`htop`, `df -h`
3. 参考故障排查章节
4. 提交Issue到GitHub仓库

---

**部署成功后，您将拥有一个完整的企业级AI工作平台！** 🎉
