<div align="center">

## 🚀 DV.net Merchant Backend
<br>

[🇬🇧 English](../README.md) • [🇷🇺 Русский](README.ru.md) • [🇨🇳 中文](README.zh.md)

[官网](https://dv.net) • [文档](https://docs.dv.net) • [API](https://docs.dv.net/en/operations/post-v1-external-wallet.html) • [支持](https://dv.net/#support)

</div>

---

## 💡 项目简介

**DV.net Merchant Backend** 是一款高性能的自托管后端平台，用于接收、处理和发送加密货币支付。系统完全开源，部署在自己的基础设施上即可，无需第三方托管或隐藏费用。

> 🔒 **非托管** — 私钥始终掌握在您手中
>
> ⚡ **高性能** — 基于 Go 1.24、Fiber v3、PostgreSQL 与 Redis
>
> 🌐 **广泛集成** — 支持多个公链与中心化交易所
>
> 🧱 **模块化架构** — delivery → service → storage 的干净分层

---

## ✨ 功能亮点

**🎯 业务能力**
- ✅ 无需 KYC/KYB 即可收发加密货币
- ✅ 丰富的通知、Webhook 与事件路由机制
- ✅ 手续费管理，TRON / EVM 资源优化
- ✅ 对接主流交易所（Binance、OKX、HTX、KuCoin、Bybit 等）

**🔧 技术特性**
- ✅ 基于 Fiber v3 与 Casbin RBAC 的 HTTP API
- ✅ `internal/app` 内置异步处理器与调度器
- ✅ `internal/service` 提供依赖注入与业务逻辑层
- ✅ `internal/storage` 封装 PostgreSQL / Redis 仓储
- ✅ SQL 自动生成 (`sqlc`, `pgxgen`)
- ✅ `pkg` 目录提供客户端、重试、OTP、AML 等通用库

---

## 🧭 架构概览

```text
cmd/                CLI 入口（服务、迁移、工具）
configs/            配置模板与 Casbin 策略
internal/app        应用初始化与后台任务
internal/delivery   HTTP 路由、处理器与中间件
internal/service    业务逻辑、集成与事件
internal/storage    PostgreSQL/Redis 仓储层
pkg/                外部客户端与通用库
sql/                SQL 模块、迁移与代码生成
```

架构图与 Swagger 文件位于 `docs/`（`swagger.yaml`, `swagger.json`）。

---

## 🚀 快速启动

**自托管一键安装**
```bash
sudo bash -c "$(curl -fsSL https://dv.net/install.sh)"
```

**本地开发 Docker 套件**
```bash
git clone --recursive https://github.com/dv-net/dv-bundle.git
cd dv-bundle && cp .env.example .env
docker compose up -d
```

**手动构建后端**
```bash
git clone https://github.com/dv-net/dv-merchant.git
cd dv-merchant

make update-frontend
make build
```

构建完成后可在 `.bin/` 中找到可执行文件 `github.com/dv-net/dv-merchant`。

---

## 🧪 开发与测试流程

**提交代码前的检查清单**
- 运行代码格式化与静态分析，确保风格一致。
- 执行单元测试，覆盖关键业务流程。
- 若新增功能或修复缺陷，请补充相应测试用例。

```bash
# 静态分析与格式化
make lint
go fmt ./...

# 单元测试
make test
```

> ℹ️ 更多场景（`make run`、Docker Compose 等）详见 [`dv-bundle`](https://github.com/dv-net/dv-bundle) 仓库的 `README.md` 以及 https://docs.dv.net。

---

## 🛠 CLI 命令

- `.bin/dv-merchant start` — 启动 HTTP API
- `.bin/dv-merchant migrate up|down` — 数据库迁移
- `.bin/dv-merchant seed up|down` — 初始化种子数据
- `.bin/dv-merchant config` — 配置校验与 ENV 生成
- `.bin/dv-merchant permission` — 角色与策略管理
- `.bin/dv-merchant transactions` — 交易辅助工具
- `.bin/dv-merchant users` — 用户管理命令

---

## 📚 文档资源

- 📖 [完整指南](https://docs.dv.net) — 安装、配置与常见场景
- 🔌 [API 参考](https://docs.dv.net/en/operations/post-v1-external-wallet.html) — 请求/响应格式
- 🧾 [Swagger](swagger.yaml) — 仓库内的接口定义

---

## 🔐 安全特性

1. 🔓 非托管模式 — 私钥与地址由您掌控。
2. 🧠 支持多签与 TRON 资源委托。
3. 🛡️ 基于 Casbin 的 RBAC，可通过 `configs/rbac_*` 灵活配置。
4. 📜 完整审计：事件、日志与 Prometheus 指标。

---

## 🤝 贡献方式

```bash
# 提交 PR 前
make lint
go test ./...
```

- ⭐ 如果项目对您有帮助，欢迎点亮 Star。
- 🐛 通过 Issues 反馈问题。
- 💡 提出新功能或业务场景建议。
- 🔧 欢迎提交 Pull Request！

---

## 💝 捐赠

用加密货币支持项目发展：

> <img src="assets/icons/coins/IconUsdt.png" width="17"> **USDT (Tron)** — `TCB4bYYN5x1z9Z4bBZ7p3XxcMwdtCfmNdN`

> <img src="assets/icons/coins/IconBtcBitcoin.png" width="17"> **Bitcoin** — `bc1qemvhkgzr4r7ksgxl8lv0lw7mnnthfc6990v3c2`

> <img src="assets/icons/coins/IconTrxTron.png" width="17"> **TRON (TRX)** — `TCB4bYYN5x1z9Z4bBZ7p3XxcMwdtCfmNdN`

> <img src="assets/icons/coins/IconEthEthereum.png" width="17"> **Ethereum** — `0xf1e4c7b968a20aae891cc18b1d5836b806691d47`

🔗 其他网络和代币（BNB Chain、Arbitrum、Polygon、Litecoin、Dogecoin、Bitcoin Cash 等）可通过 **[支付表单](https://cloud.dv.net/pay/store/208ec77f-d516-46b9-b280-3c12e1a75071/donate)** 捐赠

---

## 📞 联系我们

<div align="center">

**Telegram:** [@dv_net_support_bot](https://t.me/dv_net_support_bot) • **Telegram 群组:** [@dv_net_support_chat](https://t.me/dv_net_support_chat) • **Discord:** [discord.gg/Szy2XGsr](https://discord.gg/Szy2XGsr)

**邮箱:** [support@dv.net](https://dv.net/#support) • **网站:** [dv.net](https://dv.net) • **文档:** [docs.dv.net](https://docs.dv.net)

</div>

---

<div align="center">

**© 2026 DV.net** • [DV Technologies Ltd.](https://dv.net)

*以热忱与初心服务加密社区*

</div>
