#!/bin/bash

# CozeRights + Coze Studio 一键部署脚本
# 版本: 1.0.0
# 作者: CozeRights Team

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 未安装，请先安装 $1"
        exit 1
    fi
}

# 检查系统要求
check_requirements() {
    log_info "检查系统要求..."
    
    # 检查必要命令
    check_command "docker"
    check_command "docker-compose"
    check_command "curl"
    check_command "jq"
    
    # 检查Docker版本
    DOCKER_VERSION=$(docker --version | grep -oE '[0-9]+\.[0-9]+' | head -1)
    if [[ $(echo "$DOCKER_VERSION < 20.10" | bc -l) -eq 1 ]]; then
        log_warning "Docker版本较低，建议升级到20.10+，当前版本: $DOCKER_VERSION"
    fi
    
    # 检查系统资源
    MEMORY_GB=$(free -g | awk '/^Mem:/{print $2}')
    if [ $MEMORY_GB -lt 8 ]; then
        log_warning "系统内存不足8GB，当前: ${MEMORY_GB}GB，可能影响性能"
    fi
    
    DISK_GB=$(df -BG / | awk 'NR==2{print $4}' | sed 's/G//')
    if [ $DISK_GB -lt 50 ]; then
        log_warning "磁盘空间不足50GB，当前可用: ${DISK_GB}GB"
    fi
    
    log_success "系统要求检查完成"
}

# 创建目录结构
create_directories() {
    log_info "创建目录结构..."
    
    mkdir -p {logs/{coze,cozerights,sync,nginx},data/{mysql,postgres,redis,minio},backups,monitoring/{prometheus,grafana/{dashboards,datasources}}}
    mkdir -p docker/nginx/ssl
    mkdir -p backend/{configs,migrations}
    
    log_success "目录结构创建完成"
}

# 生成配置文件
generate_configs() {
    log_info "生成配置文件..."
    
    # 生成JWT密钥
    JWT_SECRET=$(openssl rand -base64 32)
    ENCRYPTION_KEY=$(openssl rand -base64 32)
    
    # 生成环境配置
    cat > .env << EOF
# 生成时间: $(date)
# CozeRights + Coze Studio 集成配置

# 安全配置
JWT_SECRET=$JWT_SECRET
ENCRYPTION_KEY=$ENCRYPTION_KEY

# 数据库配置
MYSQL_ROOT_PASSWORD=coze123
POSTGRES_PASSWORD=cozerights123

# 服务配置
COZE_PORT=8888
COZERIGHTS_PORT=8080

# 外部服务配置
OPENAI_API_KEY=${OPENAI_API_KEY:-}
OPENAI_BASE_URL=${OPENAI_BASE_URL:-https://api.openai.com/v1}

# 监控配置
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
GRAFANA_ADMIN_PASSWORD=admin123
EOF

    # 生成Nginx配置
    cat > docker/nginx/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream coze-studio {
        server coze-server:8888;
    }
    
    upstream cozerights {
        server cozerights-server:8080;
    }
    
    server {
        listen 80;
        server_name _;
        
        # Coze Studio
        location / {
            proxy_pass http://coze-studio;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        
        # CozeRights API
        location /api/v1/cozerights/ {
            rewrite ^/api/v1/cozerights/(.*) /api/v1/$1 break;
            proxy_pass http://cozerights;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        
        # CozeRights 管理界面
        location /admin/ {
            proxy_pass http://cozerights/admin/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
EOF

    # 生成Prometheus配置
    cat > monitoring/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'coze-studio'
    static_configs:
      - targets: ['coze-server:8888']
    metrics_path: '/metrics'
    
  - job_name: 'cozerights'
    static_configs:
      - targets: ['cozerights-server:8080']
    metrics_path: '/metrics'
    
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOF

    log_success "配置文件生成完成"
}

# 构建镜像
build_images() {
    log_info "构建Docker镜像..."
    
    # 构建Coze Studio镜像
    log_info "构建Coze Studio镜像..."
    docker-compose -f docker-compose.cozerights.yml build coze-server
    
    # 构建CozeRights镜像
    log_info "构建CozeRights镜像..."
    docker-compose -f docker-compose.cozerights.yml build cozerights-server
    
    # 构建数据同步服务镜像
    log_info "构建数据同步服务镜像..."
    docker-compose -f docker-compose.cozerights.yml build data-sync-service
    
    log_success "镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 启动基础设施服务
    log_info "启动基础设施服务..."
    docker-compose -f docker-compose.cozerights.yml up -d mysql postgres redis minio
    
    # 等待数据库启动
    log_info "等待数据库启动..."
    sleep 30
    
    # 启动应用服务
    log_info "启动应用服务..."
    docker-compose -f docker-compose.cozerights.yml up -d cozerights-server data-sync-service
    
    # 等待CozeRights启动
    log_info "等待CozeRights启动..."
    sleep 20
    
    # 启动Coze Studio
    log_info "启动Coze Studio..."
    docker-compose -f docker-compose.cozerights.yml up -d coze-server
    
    # 启动监控和代理服务
    log_info "启动监控和代理服务..."
    docker-compose -f docker-compose.cozerights.yml up -d nginx prometheus grafana
    
    log_success "所有服务启动完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 等待服务完全启动
    sleep 30
    
    # 检查Coze Studio
    if curl -f http://localhost:8888/health &>/dev/null; then
        log_success "Coze Studio 健康检查通过"
    else
        log_error "Coze Studio 健康检查失败"
        return 1
    fi
    
    # 检查CozeRights
    if curl -f http://localhost:8080/health &>/dev/null; then
        log_success "CozeRights 健康检查通过"
    else
        log_error "CozeRights 健康检查失败"
        return 1
    fi
    
    # 检查数据库连接
    if docker-compose -f docker-compose.cozerights.yml exec -T mysql mysql -u root -pcoze123 -e "SELECT 1;" &>/dev/null; then
        log_success "MySQL 连接正常"
    else
        log_error "MySQL 连接失败"
        return 1
    fi
    
    if docker-compose -f docker-compose.cozerights.yml exec -T postgres psql -U postgres -d cozerights -c "SELECT 1;" &>/dev/null; then
        log_success "PostgreSQL 连接正常"
    else
        log_error "PostgreSQL 连接失败"
        return 1
    fi
    
    log_success "所有健康检查通过"
}

# 初始化数据
init_data() {
    log_info "初始化系统数据..."
    
    # 等待服务稳定
    sleep 10
    
    # 创建默认管理员用户
    log_info "创建默认管理员用户..."
    curl -X POST http://localhost:8080/api/v1/admin/init \
        -H "Content-Type: application/json" \
        -d '{
            "username": "admin",
            "email": "admin@example.com",
            "password": "admin123"
        }' &>/dev/null || log_warning "管理员用户可能已存在"
    
    log_success "数据初始化完成"
}

# 显示部署结果
show_result() {
    log_success "🎉 CozeRights + Coze Studio 部署完成！"
    echo
    echo "📋 服务访问地址:"
    echo "  🌐 Coze Studio:      http://localhost:8888"
    echo "  🔐 CozeRights 管理:  http://localhost:8080/admin"
    echo "  📊 Grafana 监控:     http://localhost:3000 (admin/admin123)"
    echo "  📈 Prometheus:       http://localhost:9090"
    echo
    echo "🔑 默认管理员账号:"
    echo "  用户名: admin"
    echo "  密码:   admin123"
    echo
    echo "📚 更多信息:"
    echo "  📖 用户手册:   docs/USER_MANUAL.md"
    echo "  🔧 API文档:    docs/API_REFERENCE.md"
    echo "  🚀 部署指南:   docs/DEPLOYMENT_GUIDE.md"
    echo
    echo "🆘 如需帮助，请查看文档或联系技术支持"
}

# 主函数
main() {
    echo "🚀 开始部署 CozeRights + Coze Studio 集成系统"
    echo "=================================================="
    
    check_requirements
    create_directories
    generate_configs
    build_images
    start_services
    health_check
    init_data
    show_result
    
    log_success "部署完成！"
}

# 执行主函数
main "$@"
