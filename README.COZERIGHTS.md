# Coze Studio + CozeRights 企业级AI工作平台

![CozeRights Logo](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/943f576df3424fa98580c2ad18946719~tplv-goo7wpa0wc-image.image)

<div align="center">
<p>
<a href="#项目概述">项目概述</a> •
<a href="#核心特性">核心特性</a> •
<a href="#快速开始">快速开始</a> •
<a href="#架构设计">架构设计</a> •
<a href="#部署指南">部署指南</a> •
<a href="#API文档">API文档</a>
</p>
<p>
  <img alt="License" src="https://img.shields.io/badge/license-apache2.0-blue.svg">
  <img alt="Go Version" src="https://img.shields.io/badge/go-%3E%3D%201.21-blue">
  <img alt="React Version" src="https://img.shields.io/badge/react-18.0-blue">
  <img alt="Build Status" src="https://img.shields.io/badge/build-passing-green">
</p>

[English](README.md) | 中文

</div>

## 🎯 项目概述

这是一个**革命性的企业级AI工作平台解决方案**，将开源的 [Coze Studio](https://github.com/coze-dev/coze-studio) 与强大的 **CozeRights 权限管理系统**深度集成，为企业提供安全、可控、可审计的AI开发和管理环境。

### 🌟 为什么选择我们？

- **🏢 企业就绪**: 完整的多租户架构，支持大规模企业部署
- **🔒 安全可控**: 细粒度权限控制，完整的审计追踪
- **💰 成本透明**: 智能计费系统，精确的使用量统计
- **📈 可观测性**: 实时监控告警，全方位的运营洞察
- **🚀 开箱即用**: 一键部署，完整的企业级功能

## ✨ 核心特性

### 🔐 企业级权限管理
- **多租户架构**: 完全隔离的租户数据和权限
- **细粒度控制**: 工作空间、资源、操作级别的权限管理
- **角色体系**: 灵活的角色定义和权限继承
- **动态策略**: 基于条件的智能权限控制

### 📊 完整审计追踪
- **操作记录**: 所有用户操作的详细记录
- **合规报告**: 满足企业合规要求的审计报告
- **数据导出**: 支持CSV、Excel等格式的数据导出
- **实时监控**: 异常操作的实时检测和告警

### 🚀 无缝Coze Studio集成
- **透明集成**: 无需修改Coze Studio代码的权限验证
- **API拦截**: 智能的API路径识别和权限检查
- **会话管理**: 统一的用户认证和会话管理
- **数据同步**: 自动的用户和工作空间数据同步

### 💰 智能计费系统
- **使用量统计**: 精确的API调用、存储、计算资源统计
- **灵活定价**: 多层级定价策略和配额管理
- **自动账单**: 定期生成详细的使用账单
- **成本控制**: 实时的成本监控和预算告警

### 📈 实时监控告警
- **系统指标**: CPU、内存、网络等系统资源监控
- **业务指标**: API调用量、错误率、响应时间等
- **智能告警**: 基于阈值和趋势的智能告警
- **多渠道通知**: 邮件、短信、Webhook等通知方式

## 🏗️ 架构设计

### 三层架构
```
┌─────────────────────────────────────────────────────────────┐
│                    第三层：高级功能层                        │
│  ┌─────────────────┬─────────────────┬─────────────────┐    │
│  │   动态策略引擎   │    计费系统     │   监控告警系统   │    │
│  │ Policy Engine   │ Billing System  │ Monitoring      │    │
│  └─────────────────┴─────────────────┴─────────────────┘    │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   第二层：Coze集成层                         │
│  ┌─────────────────┬─────────────────┬─────────────────┐    │
│  │   权限验证中间件 │   API路径映射   │   使用量统计     │    │
│  │ Auth Middleware │ API Mapping     │ Usage Tracking  │    │
│  └─────────────────┴─────────────────┴─────────────────┘    │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   第一层：基础权限层                         │
│  ┌─────────────────┬─────────────────┬─────────────────┐    │
│  │     RBAC系统    │    审计日志     │   多租户管理     │    │
│  │   RBAC System   │  Audit Logging  │ Multi-Tenancy   │    │
│  └─────────────────┴─────────────────┴─────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### 技术栈
- **后端**: Go 1.21+, Gin, GORM, Redis
- **前端**: React 18, TypeScript, Ant Design
- **数据库**: MySQL (Coze Studio), PostgreSQL (CozeRights)
- **缓存**: Redis
- **存储**: MinIO
- **监控**: Prometheus, Grafana
- **部署**: Docker, Docker Compose

## 🚀 快速开始

### 系统要求
- **CPU**: 4核心以上
- **内存**: 8GB以上（推荐16GB）
- **存储**: 50GB以上可用空间
- **软件**: Docker 20.10+, Docker Compose 2.0+

### 一键部署
```bash
# 1. 克隆仓库
git clone https://github.com/coze-dev/coze-studio.git
cd coze-studio

# 2. 运行部署脚本
chmod +x scripts/deploy.sh
./scripts/deploy.sh

# 3. 等待部署完成（约5-10分钟）
# 部署完成后会显示访问地址和默认账号
```

### 访问系统
- **Coze Studio**: http://localhost:8888
- **CozeRights 管理**: http://localhost:8080/admin
- **监控面板**: http://localhost:3000

### 默认账号
- **用户名**: admin
- **密码**: admin123

## 📚 文档导航

### 📖 用户文档
- [用户使用手册](docs/USER_MANUAL.md) - 详细的功能使用说明
- [部署指南](docs/DEPLOYMENT_GUIDE.md) - 生产环境部署指南
- [API参考](docs/API_REFERENCE.md) - 完整的API接口文档

### 🔧 开发文档
- [架构设计](docs/ARCHITECTURE.md) - 系统架构和设计原理
- [开发指南](docs/DEVELOPER_GUIDE.md) - 二次开发和定制指南
- [贡献指南](docs/CONTRIBUTING.md) - 如何参与项目贡献

### 📊 运维文档
- [监控配置](docs/MONITORING.md) - 监控和告警配置
- [故障排查](docs/TROUBLESHOOTING.md) - 常见问题和解决方案
- [性能优化](docs/PERFORMANCE.md) - 性能调优建议

## 🎯 功能特性

### 权限管理
- ✅ 多租户权限隔离
- ✅ 工作空间级权限控制
- ✅ 细粒度资源权限
- ✅ 角色和权限继承
- ✅ 动态权限策略

### 审计日志
- ✅ 完整操作记录
- ✅ 高级筛选和搜索
- ✅ 数据导出功能
- ✅ 合规报告生成
- ✅ 实时日志监控

### Coze Studio集成
- ✅ 无侵入式集成
- ✅ 智能API识别
- ✅ 统一用户认证
- ✅ 自动数据同步
- ✅ 权限透明验证

### 高级功能
- ✅ 智能计费系统
- ✅ 使用量统计分析
- ✅ 实时监控告警
- ✅ 自动化运维
- ✅ 企业级安全

## 🤝 社区和支持

### 获取帮助
- 📖 查看[用户手册](docs/USER_MANUAL.md)
- 🐛 提交[Issue](https://github.com/coze-dev/coze-studio/issues)
- 💬 参与[讨论](https://github.com/coze-dev/coze-studio/discussions)
- 📧 联系技术支持

### 贡献代码
我们欢迎社区贡献！请查看[贡献指南](docs/CONTRIBUTING.md)了解如何参与。

### 许可证
本项目采用 [Apache 2.0 许可证](LICENSE)。

## 🎉 致谢

感谢所有为这个项目做出贡献的开发者和用户！

特别感谢：
- [Coze Studio](https://github.com/coze-dev/coze-studio) 团队提供的优秀开源AI平台
- 所有测试用户和反馈者
- 开源社区的支持和贡献

---

**让企业AI开发变得更安全、更可控、更高效！** 🚀

<div align="center">
<p>如果这个项目对您有帮助，请给我们一个 ⭐️</p>
</div>
