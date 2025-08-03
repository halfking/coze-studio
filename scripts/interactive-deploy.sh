#!/bin/bash

# CozeRights + Coze Studio 智能化一键部署脚本
# 版本: 2.0.0
# 作者: CozeRights Team
# 功能: 交互式部署、系统兼容性检查、自动依赖安装、健康检查

set -e

# 脚本配置
SCRIPT_VERSION="2.0.0"
SCRIPT_NAME="CozeRights Interactive Deployer"
TIMEOUT_SECONDS=30
SILENT_MODE=false
ROLLBACK_ENABLED=true

# 颜色和样式定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# 特殊字符
CHECK_MARK="✅"
CROSS_MARK="❌"
WARNING_MARK="⚠️"
INFO_MARK="ℹ️"
ROCKET="🚀"
GEAR="⚙️"

# 全局变量
OS_TYPE=""
ARCH_TYPE=""
PACKAGE_MANAGER=""
DOCKER_COMPOSE_CMD=""
DEPLOYMENT_MODE=""
DEPLOY_COZE_STUDIO=true
DEPLOY_COZERIGHTS=true
DEPLOY_MONITORING=true
ENVIRONMENT_TYPE="production"
ROLLBACK_STACK=()

# 默认配置
DEFAULT_COZE_PORT=8888
DEFAULT_COZERIGHTS_PORT=8080
DEFAULT_GRAFANA_PORT=3000
DEFAULT_PROMETHEUS_PORT=9090
DEFAULT_MYSQL_PASSWORD="coze123"
DEFAULT_POSTGRES_PASSWORD="cozerights123"
DEFAULT_ADMIN_USERNAME="admin"
DEFAULT_ADMIN_PASSWORD="admin123"

# 日志函数
log_info() {
    echo -e "${BLUE}${INFO_MARK} [INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}${CHECK_MARK} [SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}${WARNING_MARK} [WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}${CROSS_MARK} [ERROR]${NC} $1"
}

log_step() {
    echo -e "${PURPLE}${GEAR} [STEP]${NC} $1"
}

log_header() {
    echo -e "${WHITE}${BOLD}$1${NC}"
}

# 进度条函数
show_progress() {
    local current=$1
    local total=$2
    local message=$3
    local width=50
    local percentage=$((current * 100 / total))
    local completed=$((current * width / total))
    
    printf "\r${CYAN}[${NC}"
    for ((i=0; i<completed; i++)); do printf "█"; done
    for ((i=completed; i<width; i++)); do printf "░"; done
    printf "${CYAN}] %3d%% %s${NC}" "$percentage" "$message"
    
    if [ $current -eq $total ]; then
        echo
    fi
}

# 超时输入函数
read_with_timeout() {
    local prompt="$1"
    local default="$2"
    local timeout="$3"
    local result=""
    
    if [ "$SILENT_MODE" = true ]; then
        echo "$default"
        return 0
    fi
    
    echo -n -e "${CYAN}$prompt${NC}"
    if [ -n "$default" ]; then
        echo -n -e " ${YELLOW}[默认: $default]${NC}"
    fi
    echo -n ": "
    
    if read -t "$timeout" result; then
        if [ -z "$result" ] && [ -n "$default" ]; then
            echo "$default"
        else
            echo "$result"
        fi
    else
        echo
        log_warning "输入超时，使用默认值: $default"
        echo "$default"
    fi
}

# 确认函数
confirm() {
    local message="$1"
    local default="${2:-y}"
    
    if [ "$SILENT_MODE" = true ]; then
        return 0
    fi
    
    while true; do
        local response=$(read_with_timeout "$message (y/n)" "$default" 10)
        case "$response" in
            [Yy]|[Yy][Ee][Ss]) return 0 ;;
            [Nn]|[Nn][Oo]) return 1 ;;
            *) log_warning "请输入 y 或 n" ;;
        esac
    done
}

# 系统检测函数
detect_system() {
    log_step "检测系统环境..."
    
    # 检测操作系统
    case "$(uname -s)" in
        Darwin)
            OS_TYPE="macos"
            PACKAGE_MANAGER="brew"
            ;;
        Linux)
            OS_TYPE="linux"
            if command -v apt-get >/dev/null 2>&1; then
                PACKAGE_MANAGER="apt"
            elif command -v yum >/dev/null 2>&1; then
                PACKAGE_MANAGER="yum"
            elif command -v dnf >/dev/null 2>&1; then
                PACKAGE_MANAGER="dnf"
            else
                log_error "不支持的Linux发行版"
                exit 1
            fi
            ;;
        *)
            log_error "不支持的操作系统: $(uname -s)"
            exit 1
            ;;
    esac
    
    # 检测架构
    case "$(uname -m)" in
        x86_64|amd64)
            ARCH_TYPE="x86_64"
            ;;
        arm64|aarch64)
            ARCH_TYPE="arm64"
            ;;
        *)
            log_error "不支持的系统架构: $(uname -m)"
            exit 1
            ;;
    esac
    
    # 检测Docker Compose命令
    if command -v docker-compose >/dev/null 2>&1; then
        DOCKER_COMPOSE_CMD="docker-compose"
    elif docker compose version >/dev/null 2>&1; then
        DOCKER_COMPOSE_CMD="docker compose"
    fi
    
    log_success "系统检测完成: $OS_TYPE ($ARCH_TYPE), 包管理器: $PACKAGE_MANAGER"
}

# 显示欢迎信息
show_welcome() {
    clear
    echo -e "${CYAN}"
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║    ██████╗ ██████╗ ███████╗███████╗██████╗ ██╗ ██████╗ ██╗  ██╗████████╗███████╗ ║
║   ██╔════╝██╔═══██╗╚══███╔╝██╔════╝██╔══██╗██║██╔════╝ ██║  ██║╚══██╔══╝██╔════╝ ║
║   ██║     ██║   ██║  ███╔╝ █████╗  ██████╔╝██║██║  ███╗███████║   ██║   ███████╗ ║
║   ██║     ██║   ██║ ███╔╝  ██╔══╝  ██╔══██╗██║██║   ██║██╔══██║   ██║   ╚════██║ ║
║   ╚██████╗╚██████╔╝███████╗███████╗██║  ██║██║╚██████╔╝██║  ██║   ██║   ███████║ ║
║    ╚═════╝ ╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝ ║
║                                                                              ║
║                    企业级AI工作平台智能部署工具                                ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
    
    log_header "欢迎使用 $SCRIPT_NAME v$SCRIPT_VERSION"
    echo
    log_info "本工具将帮助您快速部署 CozeRights + Coze Studio 企业级AI工作平台"
    log_info "支持完全自动化部署，包括依赖检查、环境配置、服务启动和健康检查"
    echo
}

# 显示主菜单
show_main_menu() {
    log_header "请选择部署模式："
    echo
    echo -e "${WHITE}1.${NC} ${GREEN}完整部署${NC}     - 部署 Coze Studio + CozeRights + 监控套件 ${YELLOW}[推荐]${NC}"
    echo -e "${WHITE}2.${NC} ${BLUE}仅 Coze Studio${NC} - 只部署开源的 Coze Studio 平台"
    echo -e "${WHITE}3.${NC} ${PURPLE}仅 CozeRights${NC}  - 只部署权限管理系统"
    echo -e "${WHITE}4.${NC} ${CYAN}开发环境${NC}     - 开发调试模式部署"
    echo -e "${WHITE}5.${NC} ${YELLOW}自定义部署${NC}   - 自定义选择组件"
    echo -e "${WHITE}6.${NC} ${RED}退出${NC}         - 退出部署工具"
    echo
}

# 处理命令行参数
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --silent|-s)
                SILENT_MODE=true
                DEPLOYMENT_MODE="complete"
                shift
                ;;
            --mode|-m)
                DEPLOYMENT_MODE="$2"
                shift 2
                ;;
            --no-rollback)
                ROLLBACK_ENABLED=false
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo
    echo "选项:"
    echo "  --silent, -s          静默模式，使用默认配置"
    echo "  --mode, -m MODE       部署模式 (complete|coze|cozerights|dev|custom)"
    echo "  --no-rollback         禁用回滚功能"
    echo "  --help, -h            显示此帮助信息"
    echo
    echo "部署模式:"
    echo "  complete              完整部署 (默认)"
    echo "  coze                  仅部署 Coze Studio"
    echo "  cozerights            仅部署 CozeRights"
    echo "  dev                   开发环境部署"
    echo "  custom                自定义部署"
    echo
}

# 系统资源检查
check_system_resources() {
    log_step "检查系统资源..."

    local cpu_cores=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "unknown")
    local memory_gb=""
    local disk_gb=""

    # 检查CPU核心数
    if [ "$cpu_cores" != "unknown" ] && [ "$cpu_cores" -lt 4 ]; then
        log_warning "CPU核心数不足: ${cpu_cores}核 (建议≥4核)"
        if ! confirm "是否继续部署？"; then
            exit 1
        fi
    else
        log_success "CPU核心数: ${cpu_cores}核"
    fi

    # 检查内存
    if [ "$OS_TYPE" = "macos" ]; then
        memory_gb=$(( $(sysctl -n hw.memsize) / 1024 / 1024 / 1024 ))
    else
        memory_gb=$(free -g | awk '/^Mem:/{print $2}')
    fi

    if [ "$memory_gb" -lt 8 ]; then
        log_warning "内存不足: ${memory_gb}GB (建议≥8GB)"
        if ! confirm "是否继续部署？"; then
            exit 1
        fi
    else
        log_success "内存: ${memory_gb}GB"
    fi

    # 检查磁盘空间
    disk_gb=$(df -BG . | awk 'NR==2{print $4}' | sed 's/G//')
    if [ "$disk_gb" -lt 50 ]; then
        log_warning "磁盘空间不足: ${disk_gb}GB (建议≥50GB)"
        if ! confirm "是否继续部署？"; then
            exit 1
        fi
    else
        log_success "磁盘空间: ${disk_gb}GB"
    fi
}

# 网络连接检查
check_network_connectivity() {
    log_step "检查网络连接..."

    local endpoints=(
        "https://hub.docker.com"
        "https://github.com"
        "https://registry-1.docker.io"
    )

    for endpoint in "${endpoints[@]}"; do
        if curl -s --connect-timeout 5 "$endpoint" >/dev/null; then
            log_success "网络连接正常: $endpoint"
        else
            log_warning "网络连接失败: $endpoint"
        fi
    done
}

# 检查命令是否存在
check_command() {
    if command -v "$1" >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# 安装依赖
install_dependencies() {
    log_step "检查和安装依赖..."

    local dependencies=("docker" "git" "curl" "jq")
    local missing_deps=()

    # 检查缺失的依赖
    for dep in "${dependencies[@]}"; do
        if ! check_command "$dep"; then
            missing_deps+=("$dep")
        else
            log_success "$dep 已安装"
        fi
    done

    # 安装缺失的依赖
    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_info "需要安装以下依赖: ${missing_deps[*]}"

        if confirm "是否自动安装缺失的依赖？"; then
            case "$PACKAGE_MANAGER" in
                brew)
                    for dep in "${missing_deps[@]}"; do
                        log_info "安装 $dep..."
                        brew install "$dep" || log_error "安装 $dep 失败"
                    done
                    ;;
                apt)
                    sudo apt-get update
                    for dep in "${missing_deps[@]}"; do
                        log_info "安装 $dep..."
                        case "$dep" in
                            docker)
                                curl -fsSL https://get.docker.com -o get-docker.sh
                                sudo sh get-docker.sh
                                sudo usermod -aG docker "$USER"
                                rm get-docker.sh
                                ;;
                            *)
                                sudo apt-get install -y "$dep"
                                ;;
                        esac
                    done
                    ;;
                yum|dnf)
                    for dep in "${missing_deps[@]}"; do
                        log_info "安装 $dep..."
                        case "$dep" in
                            docker)
                                curl -fsSL https://get.docker.com -o get-docker.sh
                                sudo sh get-docker.sh
                                sudo usermod -aG docker "$USER"
                                sudo systemctl enable docker
                                sudo systemctl start docker
                                rm get-docker.sh
                                ;;
                            *)
                                sudo "$PACKAGE_MANAGER" install -y "$dep"
                                ;;
                        esac
                    done
                    ;;
            esac
        else
            log_error "缺少必要依赖，无法继续部署"
            exit 1
        fi
    fi

    # 检查Docker Compose
    if [ -z "$DOCKER_COMPOSE_CMD" ]; then
        log_info "安装 Docker Compose..."
        if [ "$OS_TYPE" = "macos" ]; then
            brew install docker-compose
        else
            sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
            sudo chmod +x /usr/local/bin/docker-compose
        fi
        DOCKER_COMPOSE_CMD="docker-compose"
    fi
}

# 验证工具版本
verify_tool_versions() {
    log_step "验证工具版本..."

    # 检查Docker版本
    local docker_version=$(docker --version | grep -oE '[0-9]+\.[0-9]+' | head -1)
    if [ "$(echo "$docker_version >= 20.10" | bc -l 2>/dev/null || echo 0)" -eq 1 ]; then
        log_success "Docker版本: $docker_version"
    else
        log_warning "Docker版本较低: $docker_version (建议≥20.10)"
    fi

    # 检查Docker Compose版本
    local compose_version=$($DOCKER_COMPOSE_CMD --version | grep -oE '[0-9]+\.[0-9]+' | head -1)
    if [ "$(echo "$compose_version >= 2.0" | bc -l 2>/dev/null || echo 0)" -eq 1 ]; then
        log_success "Docker Compose版本: $compose_version"
    else
        log_warning "Docker Compose版本较低: $compose_version (建议≥2.0)"
    fi

    # 检查Docker服务状态
    if docker info >/dev/null 2>&1; then
        log_success "Docker服务运行正常"
    else
        log_error "Docker服务未运行，请启动Docker服务"
        exit 1
    fi
}

# 处理菜单选择
handle_menu_selection() {
    if [ -n "$DEPLOYMENT_MODE" ]; then
        case "$DEPLOYMENT_MODE" in
            complete)
                DEPLOY_COZE_STUDIO=true
                DEPLOY_COZERIGHTS=true
                DEPLOY_MONITORING=true
                ENVIRONMENT_TYPE="production"
                ;;
            coze)
                DEPLOY_COZE_STUDIO=true
                DEPLOY_COZERIGHTS=false
                DEPLOY_MONITORING=false
                ;;
            cozerights)
                DEPLOY_COZE_STUDIO=false
                DEPLOY_COZERIGHTS=true
                DEPLOY_MONITORING=false
                ;;
            dev)
                DEPLOY_COZE_STUDIO=true
                DEPLOY_COZERIGHTS=true
                DEPLOY_MONITORING=true
                ENVIRONMENT_TYPE="development"
                ;;
            custom)
                configure_custom_deployment
                ;;
        esac
        return
    fi

    while true; do
        show_main_menu
        local choice=$(read_with_timeout "请选择部署模式 (1-6)" "1" $TIMEOUT_SECONDS)

        case "$choice" in
            1)
                log_info "选择: 完整部署"
                DEPLOY_COZE_STUDIO=true
                DEPLOY_COZERIGHTS=true
                DEPLOY_MONITORING=true
                ENVIRONMENT_TYPE="production"
                break
                ;;
            2)
                log_info "选择: 仅部署 Coze Studio"
                DEPLOY_COZE_STUDIO=true
                DEPLOY_COZERIGHTS=false
                DEPLOY_MONITORING=false
                break
                ;;
            3)
                log_info "选择: 仅部署 CozeRights"
                DEPLOY_COZE_STUDIO=false
                DEPLOY_COZERIGHTS=true
                DEPLOY_MONITORING=false
                break
                ;;
            4)
                log_info "选择: 开发环境部署"
                DEPLOY_COZE_STUDIO=true
                DEPLOY_COZERIGHTS=true
                DEPLOY_MONITORING=true
                ENVIRONMENT_TYPE="development"
                break
                ;;
            5)
                log_info "选择: 自定义部署"
                configure_custom_deployment
                break
                ;;
            6)
                log_info "退出部署"
                exit 0
                ;;
            *)
                log_error "无效选择，请输入 1-6"
                ;;
        esac
    done
}

# 自定义部署配置
configure_custom_deployment() {
    log_header "自定义部署配置"
    echo

    # 选择组件
    if confirm "是否部署 Coze Studio？" "y"; then
        DEPLOY_COZE_STUDIO=true
    else
        DEPLOY_COZE_STUDIO=false
    fi

    if confirm "是否部署 CozeRights？" "y"; then
        DEPLOY_COZERIGHTS=true
    else
        DEPLOY_COZERIGHTS=false
    fi

    if confirm "是否部署监控套件 (Prometheus + Grafana)？" "y"; then
        DEPLOY_MONITORING=true
    else
        DEPLOY_MONITORING=false
    fi

    # 选择环境类型
    echo
    log_info "选择环境类型:"
    echo "1. 生产环境 (优化性能和安全性)"
    echo "2. 开发环境 (启用调试和开发工具)"

    local env_choice=$(read_with_timeout "请选择环境类型 (1-2)" "1" $TIMEOUT_SECONDS)
    case "$env_choice" in
        1) ENVIRONMENT_TYPE="production" ;;
        2) ENVIRONMENT_TYPE="development" ;;
        *) ENVIRONMENT_TYPE="production" ;;
    esac

    log_success "自定义配置完成"
}

# 配置端口和密码
configure_deployment_settings() {
    log_step "配置部署参数..."

    if [ "$SILENT_MODE" = false ]; then
        echo
        log_header "端口配置"

        if [ "$DEPLOY_COZE_STUDIO" = true ]; then
            COZE_PORT=$(read_with_timeout "Coze Studio 端口" "$DEFAULT_COZE_PORT" $TIMEOUT_SECONDS)
        fi

        if [ "$DEPLOY_COZERIGHTS" = true ]; then
            COZERIGHTS_PORT=$(read_with_timeout "CozeRights 端口" "$DEFAULT_COZERIGHTS_PORT" $TIMEOUT_SECONDS)
        fi

        if [ "$DEPLOY_MONITORING" = true ]; then
            GRAFANA_PORT=$(read_with_timeout "Grafana 端口" "$DEFAULT_GRAFANA_PORT" $TIMEOUT_SECONDS)
            PROMETHEUS_PORT=$(read_with_timeout "Prometheus 端口" "$DEFAULT_PROMETHEUS_PORT" $TIMEOUT_SECONDS)
        fi

        echo
        log_header "安全配置"
        MYSQL_PASSWORD=$(read_with_timeout "MySQL 密码" "$DEFAULT_MYSQL_PASSWORD" $TIMEOUT_SECONDS)
        POSTGRES_PASSWORD=$(read_with_timeout "PostgreSQL 密码" "$DEFAULT_POSTGRES_PASSWORD" $TIMEOUT_SECONDS)
        ADMIN_USERNAME=$(read_with_timeout "管理员用户名" "$DEFAULT_ADMIN_USERNAME" $TIMEOUT_SECONDS)
        ADMIN_PASSWORD=$(read_with_timeout "管理员密码" "$DEFAULT_ADMIN_PASSWORD" $TIMEOUT_SECONDS)
    else
        # 静默模式使用默认值
        COZE_PORT=$DEFAULT_COZE_PORT
        COZERIGHTS_PORT=$DEFAULT_COZERIGHTS_PORT
        GRAFANA_PORT=$DEFAULT_GRAFANA_PORT
        PROMETHEUS_PORT=$DEFAULT_PROMETHEUS_PORT
        MYSQL_PASSWORD=$DEFAULT_MYSQL_PASSWORD
        POSTGRES_PASSWORD=$DEFAULT_POSTGRES_PASSWORD
        ADMIN_USERNAME=$DEFAULT_ADMIN_USERNAME
        ADMIN_PASSWORD=$DEFAULT_ADMIN_PASSWORD
    fi

    log_success "部署参数配置完成"
}

# 创建目录结构
create_directory_structure() {
    log_step "创建目录结构..."

    local directories=(
        "logs/coze"
        "logs/cozerights"
        "logs/sync"
        "logs/nginx"
        "data/mysql"
        "data/postgres"
        "data/redis"
        "data/minio"
        "backups"
        "monitoring/prometheus"
        "monitoring/grafana/dashboards"
        "monitoring/grafana/datasources"
        "docker/nginx/ssl"
        "backend/configs"
        "backend/migrations"
    )

    for dir in "${directories[@]}"; do
        mkdir -p "$dir"
        show_progress $((++created_dirs)) ${#directories[@]} "创建目录: $dir"
    done

    log_success "目录结构创建完成"
}

# 生成配置文件
generate_configuration_files() {
    log_step "生成配置文件..."

    # 生成安全密钥
    local jwt_secret=$(openssl rand -base64 32)
    local encryption_key=$(openssl rand -base64 32)

    # 生成环境配置文件
    cat > .env << EOF
# CozeRights + Coze Studio 部署配置
# 生成时间: $(date)
# 部署模式: $DEPLOYMENT_MODE
# 环境类型: $ENVIRONMENT_TYPE

# 安全配置
JWT_SECRET=$jwt_secret
ENCRYPTION_KEY=$encryption_key

# 服务端口配置
COZE_PORT=$COZE_PORT
COZERIGHTS_PORT=$COZERIGHTS_PORT
GRAFANA_PORT=$GRAFANA_PORT
PROMETHEUS_PORT=$PROMETHEUS_PORT

# 数据库配置
MYSQL_ROOT_PASSWORD=$MYSQL_PASSWORD
POSTGRES_PASSWORD=$POSTGRES_PASSWORD

# 管理员配置
ADMIN_USERNAME=$ADMIN_USERNAME
ADMIN_PASSWORD=$ADMIN_PASSWORD

# 环境配置
ENVIRONMENT=$ENVIRONMENT_TYPE
LOG_LEVEL=$([ "$ENVIRONMENT_TYPE" = "development" ] && echo "debug" || echo "info")

# 外部服务配置
OPENAI_API_KEY=${OPENAI_API_KEY:-}
OPENAI_BASE_URL=${OPENAI_BASE_URL:-https://api.openai.com/v1}
EOF

    # 生成Docker Compose配置
    generate_docker_compose_config

    # 生成Nginx配置
    generate_nginx_config

    # 生成监控配置
    if [ "$DEPLOY_MONITORING" = true ]; then
        generate_monitoring_config
    fi

    log_success "配置文件生成完成"
}

# 生成Docker Compose配置
generate_docker_compose_config() {
    local compose_file="docker-compose.deploy.yml"

    cat > "$compose_file" << EOF
version: '3.8'

services:
EOF

    # Coze Studio服务
    if [ "$DEPLOY_COZE_STUDIO" = true ]; then
        cat >> "$compose_file" << EOF
  coze-server:
    build: .
    ports:
      - "$COZE_PORT:8888"
    environment:
      - LISTEN_ADDR=:8888
      - LOG_LEVEL=\${LOG_LEVEL}
      - MYSQL_DSN=root:\${MYSQL_ROOT_PASSWORD}@tcp(mysql:3306)/opencoze?charset=utf8mb4&parseTime=True
      - REDIS_ADDR=redis:6379
      - MINIO_ENDPOINT=minio:9000
      - MINIO_AK=minioadmin
      - MINIO_SK=minioadmin123
      - MINIO_SECURE=false
      - STORAGE_BUCKET=opencoze
EOF

        if [ "$DEPLOY_COZERIGHTS" = true ]; then
            cat >> "$compose_file" << EOF
      - COZERIGHTS_ENABLED=true
      - COZERIGHTS_URL=http://cozerights-server:8080
EOF
        fi

        cat >> "$compose_file" << EOF
    depends_on:
      - mysql
      - redis
      - minio
EOF

        if [ "$DEPLOY_COZERIGHTS" = true ]; then
            cat >> "$compose_file" << EOF
      - cozerights-server
EOF
        fi

        cat >> "$compose_file" << EOF
    volumes:
      - ./backend/conf:/app/conf
      - ./logs/coze:/app/logs
    restart: unless-stopped
    networks:
      - coze-network

EOF
    fi

    # CozeRights服务
    if [ "$DEPLOY_COZERIGHTS" = true ]; then
        cat >> "$compose_file" << EOF
  cozerights-server:
    build:
      context: ./backend
      dockerfile: Dockerfile.cozerights
    ports:
      - "$COZERIGHTS_PORT:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=\${POSTGRES_PASSWORD}
      - DB_NAME=cozerights
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=\${JWT_SECRET}
      - ENCRYPTION_KEY=\${ENCRYPTION_KEY}
      - LOG_LEVEL=\${LOG_LEVEL}
EOF

        if [ "$DEPLOY_COZE_STUDIO" = true ]; then
            cat >> "$compose_file" << EOF
      - COZE_MYSQL_DSN=root:\${MYSQL_ROOT_PASSWORD}@tcp(mysql:3306)/opencoze?charset=utf8mb4&parseTime=True
EOF
        fi

        cat >> "$compose_file" << EOF
    depends_on:
      - postgres
      - redis
    volumes:
      - ./backend/configs:/app/configs
      - ./logs/cozerights:/app/logs
    restart: unless-stopped
    networks:
      - coze-network

EOF
    fi

    # 添加基础设施服务
    add_infrastructure_services "$compose_file"
}

# 添加基础设施服务到Docker Compose
add_infrastructure_services() {
    local compose_file="$1"

    # MySQL (如果部署Coze Studio)
    if [ "$DEPLOY_COZE_STUDIO" = true ]; then
        cat >> "$compose_file" << EOF
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: \${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: opencoze
      MYSQL_USER: coze
      MYSQL_PASSWORD: \${MYSQL_ROOT_PASSWORD}
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./docker/volumes/mysql/schema.sql:/docker-entrypoint-initdb.d/01-schema.sql
    command: --default-authentication-plugin=mysql_native_password
    restart: unless-stopped
    networks:
      - coze-network

EOF
    fi

    # PostgreSQL (如果部署CozeRights)
    if [ "$DEPLOY_COZERIGHTS" = true ]; then
        cat >> "$compose_file" << EOF
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: cozerights
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: \${POSTGRES_PASSWORD}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/migrations:/docker-entrypoint-initdb.d
    restart: unless-stopped
    networks:
      - coze-network

EOF
    fi

    # Redis
    cat >> "$compose_file" << EOF
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes --maxmemory 1gb --maxmemory-policy allkeys-lru
    restart: unless-stopped
    networks:
      - coze-network

EOF

    # MinIO (如果部署Coze Studio)
    if [ "$DEPLOY_COZE_STUDIO" = true ]; then
        cat >> "$compose_file" << EOF
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin123
    volumes:
      - minio_data:/data
    restart: unless-stopped
    networks:
      - coze-network

EOF
    fi

    # 监控服务
    if [ "$DEPLOY_MONITORING" = true ]; then
        cat >> "$compose_file" << EOF
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "$PROMETHEUS_PORT:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
    restart: unless-stopped
    networks:
      - coze-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "$GRAFANA_PORT:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=\${ADMIN_PASSWORD}
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
    depends_on:
      - prometheus
    restart: unless-stopped
    networks:
      - coze-network

EOF
    fi

    # 添加卷和网络定义
    cat >> "$compose_file" << EOF
volumes:
EOF

    if [ "$DEPLOY_COZE_STUDIO" = true ]; then
        cat >> "$compose_file" << EOF
  mysql_data:
    driver: local
  minio_data:
    driver: local
EOF
    fi

    if [ "$DEPLOY_COZERIGHTS" = true ]; then
        cat >> "$compose_file" << EOF
  postgres_data:
    driver: local
EOF
    fi

    cat >> "$compose_file" << EOF
  redis_data:
    driver: local
EOF

    if [ "$DEPLOY_MONITORING" = true ]; then
        cat >> "$compose_file" << EOF
  prometheus_data:
    driver: local
  grafana_data:
    driver: local
EOF
    fi

    cat >> "$compose_file" << EOF

networks:
  coze-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
EOF
}

# 生成Nginx配置
generate_nginx_config() {
    mkdir -p docker/nginx

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
}

# 生成监控配置
generate_monitoring_config() {
    # Prometheus配置
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

    # Grafana数据源配置
    mkdir -p monitoring/grafana/datasources
    cat > monitoring/grafana/datasources/prometheus.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF

    # Grafana仪表板配置
    mkdir -p monitoring/grafana/dashboards
    cat > monitoring/grafana/dashboards/dashboard.yml << EOF
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF
}

# 执行部署
execute_deployment() {
    log_step "开始部署服务..."

    local total_steps=6
    local current_step=0

    # 步骤1: 构建镜像
    ((current_step++))
    show_progress $current_step $total_steps "构建Docker镜像..."

    if [ "$DEPLOY_COZE_STUDIO" = true ]; then
        log_info "构建Coze Studio镜像..."
        $DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml build coze-server || handle_deployment_error "Coze Studio镜像构建失败"
    fi

    if [ "$DEPLOY_COZERIGHTS" = true ]; then
        log_info "构建CozeRights镜像..."
        $DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml build cozerights-server || handle_deployment_error "CozeRights镜像构建失败"
    fi

    # 步骤2: 启动基础设施
    ((current_step++))
    show_progress $current_step $total_steps "启动基础设施服务..."

    local infrastructure_services=("redis")

    if [ "$DEPLOY_COZE_STUDIO" = true ]; then
        infrastructure_services+=("mysql" "minio")
    fi

    if [ "$DEPLOY_COZERIGHTS" = true ]; then
        infrastructure_services+=("postgres")
    fi

    $DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml up -d "${infrastructure_services[@]}" || handle_deployment_error "基础设施启动失败"

    # 步骤3: 等待数据库启动
    ((current_step++))
    show_progress $current_step $total_steps "等待数据库启动..."
    sleep 30

    # 步骤4: 启动应用服务
    ((current_step++))
    show_progress $current_step $total_steps "启动应用服务..."

    local app_services=()

    if [ "$DEPLOY_COZERIGHTS" = true ]; then
        app_services+=("cozerights-server")
    fi

    if [ "$DEPLOY_COZE_STUDIO" = true ]; then
        app_services+=("coze-server")
    fi

    if [ ${#app_services[@]} -gt 0 ]; then
        $DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml up -d "${app_services[@]}" || handle_deployment_error "应用服务启动失败"
    fi

    # 步骤5: 启动监控服务
    ((current_step++))
    show_progress $current_step $total_steps "启动监控服务..."

    if [ "$DEPLOY_MONITORING" = true ]; then
        $DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml up -d prometheus grafana || handle_deployment_error "监控服务启动失败"
    fi

    # 步骤6: 等待服务稳定
    ((current_step++))
    show_progress $current_step $total_steps "等待服务稳定..."
    sleep 20

    log_success "服务部署完成"
}

# 处理部署错误
handle_deployment_error() {
    local error_message="$1"
    log_error "$error_message"

    if [ "$ROLLBACK_ENABLED" = true ]; then
        if confirm "是否执行回滚？"; then
            execute_rollback
        fi
    fi

    exit 1
}

# 执行回滚
execute_rollback() {
    log_warning "开始执行回滚..."

    # 停止所有服务
    $DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml down -v 2>/dev/null || true

    # 清理镜像
    docker image prune -f 2>/dev/null || true

    # 清理配置文件
    rm -f docker-compose.deploy.yml .env 2>/dev/null || true

    log_success "回滚完成"
}

# 健康检查
perform_health_checks() {
    log_step "执行健康检查..."

    local total_checks=0
    local passed_checks=0

    # 等待服务完全启动
    sleep 30

    # 检查Coze Studio
    if [ "$DEPLOY_COZE_STUDIO" = true ]; then
        ((total_checks++))
        log_info "检查Coze Studio服务..."
        if curl -f "http://localhost:$COZE_PORT/health" >/dev/null 2>&1; then
            log_success "Coze Studio健康检查通过"
            ((passed_checks++))
        else
            log_error "Coze Studio健康检查失败"
        fi
    fi

    # 检查CozeRights
    if [ "$DEPLOY_COZERIGHTS" = true ]; then
        ((total_checks++))
        log_info "检查CozeRights服务..."
        if curl -f "http://localhost:$COZERIGHTS_PORT/health" >/dev/null 2>&1; then
            log_success "CozeRights健康检查通过"
            ((passed_checks++))
        else
            log_error "CozeRights健康检查失败"
        fi
    fi

    # 检查数据库连接
    if [ "$DEPLOY_COZE_STUDIO" = true ]; then
        ((total_checks++))
        log_info "检查MySQL连接..."
        if $DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml exec -T mysql mysql -u root -p"$MYSQL_PASSWORD" -e "SELECT 1;" >/dev/null 2>&1; then
            log_success "MySQL连接正常"
            ((passed_checks++))
        else
            log_error "MySQL连接失败"
        fi
    fi

    if [ "$DEPLOY_COZERIGHTS" = true ]; then
        ((total_checks++))
        log_info "检查PostgreSQL连接..."
        if $DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml exec -T postgres psql -U postgres -d cozerights -c "SELECT 1;" >/dev/null 2>&1; then
            log_success "PostgreSQL连接正常"
            ((passed_checks++))
        else
            log_error "PostgreSQL连接失败"
        fi
    fi

    # 检查监控服务
    if [ "$DEPLOY_MONITORING" = true ]; then
        ((total_checks++))
        log_info "检查Prometheus服务..."
        if curl -f "http://localhost:$PROMETHEUS_PORT/-/healthy" >/dev/null 2>&1; then
            log_success "Prometheus健康检查通过"
            ((passed_checks++))
        else
            log_error "Prometheus健康检查失败"
        fi

        ((total_checks++))
        log_info "检查Grafana服务..."
        if curl -f "http://localhost:$GRAFANA_PORT/api/health" >/dev/null 2>&1; then
            log_success "Grafana健康检查通过"
            ((passed_checks++))
        else
            log_error "Grafana健康检查失败"
        fi
    fi

    # 总结健康检查结果
    if [ $passed_checks -eq $total_checks ]; then
        log_success "所有健康检查通过 ($passed_checks/$total_checks)"
        return 0
    else
        log_warning "部分健康检查失败 ($passed_checks/$total_checks)"
        return 1
    fi
}

# 初始化数据
initialize_data() {
    log_step "初始化系统数据..."

    # 等待服务稳定
    sleep 10

    if [ "$DEPLOY_COZERIGHTS" = true ]; then
        log_info "创建默认管理员用户..."

        # 创建管理员用户
        local init_response=$(curl -s -X POST "http://localhost:$COZERIGHTS_PORT/api/v1/admin/init" \
            -H "Content-Type: application/json" \
            -d "{
                \"username\": \"$ADMIN_USERNAME\",
                \"email\": \"admin@example.com\",
                \"password\": \"$ADMIN_PASSWORD\"
            }" 2>/dev/null)

        if [ $? -eq 0 ]; then
            log_success "管理员用户创建成功"
        else
            log_warning "管理员用户可能已存在"
        fi
    fi

    log_success "数据初始化完成"
}

# 生成部署报告
generate_deployment_report() {
    local report_file="deployment-report-$(date +%Y%m%d_%H%M%S).txt"

    {
        echo "CozeRights + Coze Studio 部署报告"
        echo "=================================="
        echo "部署时间: $(date)"
        echo "部署模式: $DEPLOYMENT_MODE"
        echo "环境类型: $ENVIRONMENT_TYPE"
        echo
        echo "部署组件:"
        [ "$DEPLOY_COZE_STUDIO" = true ] && echo "  ✅ Coze Studio"
        [ "$DEPLOY_COZERIGHTS" = true ] && echo "  ✅ CozeRights"
        [ "$DEPLOY_MONITORING" = true ] && echo "  ✅ 监控套件"
        echo
        echo "服务访问地址:"
        [ "$DEPLOY_COZE_STUDIO" = true ] && echo "  🌐 Coze Studio:      http://localhost:$COZE_PORT"
        [ "$DEPLOY_COZERIGHTS" = true ] && echo "  🔐 CozeRights 管理:  http://localhost:$COZERIGHTS_PORT/admin"
        [ "$DEPLOY_MONITORING" = true ] && echo "  📊 Grafana 监控:     http://localhost:$GRAFANA_PORT"
        [ "$DEPLOY_MONITORING" = true ] && echo "  📈 Prometheus:       http://localhost:$PROMETHEUS_PORT"
        echo
        echo "默认账号信息:"
        echo "  用户名: $ADMIN_USERNAME"
        echo "  密码:   $ADMIN_PASSWORD"
        echo
        echo "系统信息:"
        echo "  操作系统: $OS_TYPE ($ARCH_TYPE)"
        echo "  Docker版本: $(docker --version)"
        echo "  Docker Compose: $DOCKER_COMPOSE_CMD"
        echo
        echo "配置文件:"
        echo "  环境配置: .env"
        echo "  Docker Compose: docker-compose.deploy.yml"
        echo "  部署报告: $report_file"
        echo
    } > "$report_file"

    log_success "部署报告已生成: $report_file"
}

# 显示部署结果
show_deployment_result() {
    clear
    echo -e "${GREEN}"
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║    🎉 部署完成！CozeRights + Coze Studio 企业级AI工作平台已就绪！              ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"

    log_header "🎯 部署成功！"
    echo

    log_header "📋 服务访问地址:"
    if [ "$DEPLOY_COZE_STUDIO" = true ]; then
        echo -e "  🌐 ${CYAN}Coze Studio:${NC}      http://localhost:$COZE_PORT"
    fi
    if [ "$DEPLOY_COZERIGHTS" = true ]; then
        echo -e "  🔐 ${PURPLE}CozeRights 管理:${NC}  http://localhost:$COZERIGHTS_PORT/admin"
    fi
    if [ "$DEPLOY_MONITORING" = true ]; then
        echo -e "  📊 ${YELLOW}Grafana 监控:${NC}     http://localhost:$GRAFANA_PORT (admin/$ADMIN_PASSWORD)"
        echo -e "  📈 ${BLUE}Prometheus:${NC}       http://localhost:$PROMETHEUS_PORT"
    fi
    echo

    log_header "🔑 默认管理员账号:"
    echo -e "  ${WHITE}用户名:${NC} $ADMIN_USERNAME"
    echo -e "  ${WHITE}密码:${NC}   $ADMIN_PASSWORD"
    echo

    log_header "📚 更多信息:"
    echo -e "  📖 ${CYAN}用户手册:${NC}   docs/USER_MANUAL.md"
    echo -e "  🔧 ${CYAN}API文档:${NC}    docs/API_REFERENCE.md"
    echo -e "  🚀 ${CYAN}部署指南:${NC}   docs/DEPLOYMENT_GUIDE.md"
    echo

    log_header "🛠️ 常用命令:"
    echo -e "  查看服务状态: ${YELLOW}$DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml ps${NC}"
    echo -e "  查看日志:     ${YELLOW}$DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml logs -f${NC}"
    echo -e "  停止服务:     ${YELLOW}$DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml down${NC}"
    echo -e "  重启服务:     ${YELLOW}$DOCKER_COMPOSE_CMD -f docker-compose.deploy.yml restart${NC}"
    echo

    log_success "🎊 恭喜！您的企业级AI工作平台已成功部署并可以使用了！"
    echo
    log_info "如需帮助，请查看文档或联系技术支持"
}

# 清理函数
cleanup() {
    log_info "正在清理临时文件..."
    # 这里可以添加清理逻辑
}

# 信号处理
trap cleanup EXIT
trap 'log_error "部署被中断"; exit 1' INT TERM

# 主函数
main() {
    # 解析命令行参数
    parse_arguments "$@"

    # 显示欢迎信息
    if [ "$SILENT_MODE" = false ]; then
        show_welcome
    fi

    # 系统检测
    detect_system

    # 系统资源检查
    check_system_resources

    # 网络连接检查
    check_network_connectivity

    # 安装依赖
    install_dependencies

    # 验证工具版本
    verify_tool_versions

    # 处理菜单选择
    handle_menu_selection

    # 配置部署参数
    configure_deployment_settings

    # 显示部署确认
    if [ "$SILENT_MODE" = false ]; then
        echo
        log_header "部署确认"
        echo -e "  部署模式: ${YELLOW}$DEPLOYMENT_MODE${NC}"
        echo -e "  环境类型: ${YELLOW}$ENVIRONMENT_TYPE${NC}"
        echo -e "  Coze Studio: $([ "$DEPLOY_COZE_STUDIO" = true ] && echo "${GREEN}是${NC}" || echo "${RED}否${NC}")"
        echo -e "  CozeRights: $([ "$DEPLOY_COZERIGHTS" = true ] && echo "${GREEN}是${NC}" || echo "${RED}否${NC}")"
        echo -e "  监控套件: $([ "$DEPLOY_MONITORING" = true ] && echo "${GREEN}是${NC}" || echo "${RED}否${NC}")"
        echo

        if ! confirm "确认开始部署？"; then
            log_info "部署已取消"
            exit 0
        fi
    fi

    # 开始部署流程
    log_header "${ROCKET} 开始部署 CozeRights + Coze Studio 企业级AI工作平台"
    echo

    # 创建目录结构
    create_directory_structure

    # 生成配置文件
    generate_configuration_files

    # 执行部署
    execute_deployment

    # 健康检查
    if ! perform_health_checks; then
        log_warning "部分服务可能存在问题，请检查日志"
    fi

    # 初始化数据
    initialize_data

    # 生成部署报告
    generate_deployment_report

    # 显示部署结果
    show_deployment_result

    log_success "部署流程完成！"
}

# 脚本入口点
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
