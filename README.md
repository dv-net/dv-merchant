<div align="center">

## 🚀 DV.net Merchant Backend
<br>

[🇬🇧 English](README.md) • [🇷🇺 Русский](docs/README.ru.md) • [🇨🇳 中文](docs/README.zh.md)

[Website](https://dv.net) • [Docs](https://docs.dv.net) • [API](https://docs.dv.net/en/operations/post-v1-external-wallet.html) • [Support](https://dv.net/#support)

</div>

---

## 💡 Overview

**DV.net Merchant Backend** is a high-load, self-hosted payments platform for accepting, processing, and sending cryptocurrency. The stack is fully open source, runs on your own infrastructure, and keeps you in control of every transaction.

> 🔒 **Non-custodial** — private keys always stay on your side
>
> ⚡ **High-performance** — Go 1.24, Fiber v3, PostgreSQL & Redis
>
> 🌐 **Wide coverage** — multiple blockchains and centralized exchanges
>
> 🧱 **Modular** — clean architecture with delivery → service → storage

---

## ✨ Highlights

**🎯 Business capabilities**
- ✅ Accept and send crypto without mandatory KYC/KYB
- ✅ Notifications, webhooks, and flexible event routing
- ✅ Fee management plus TRON/EVM resource optimization
- ✅ Integrations with major CEXs (Binance, OKX, HTX, KuCoin, Bybit, etc.)

**🔧 Technical features**
- ✅ Fiber v3 HTTP API with Casbin-based RBAC
- ✅ Async workers and schedulers in `internal/app`
- ✅ Service layer with DI and business logic (`internal/service`)
- ✅ PostgreSQL / Redis repositories (`internal/storage`)
- ✅ Automated SQL generation (`sqlc`, `pgxgen`)
- ✅ Rich helper packages in `pkg` (clients, retry, OTP, AML)

---

## 🧭 Architecture at a Glance

```text
cmd/                CLI entrypoints (server, migrations, utilities)
configs/            Config templates and Casbin policies
internal/app        App bootstrap and background jobs
internal/delivery   HTTP handlers, middleware, routing
internal/service    Business logic, integrations, events
internal/storage    PostgreSQL/Redis repositories
pkg/                External clients and shared libraries
sql/                SQL modules, migrations, code generation
```

Diagrams and Swagger specs live in `docs/` (`swagger.yaml`, `swagger.json`).

---

## 🚀 Getting Started

**Self-hosted install (one command)**
```bash
sudo bash -c "$(curl -fsSL https://dv.net/install.sh)"
```

**Developer Docker bundle**
```bash
git clone --recursive https://github.com/dv-net/dv-bundle.git
cd dv-bundle && cp .env.example .env
docker compose up -d
```

**Manual backend build**
```bash
git clone https://github.com/dv-net/dv-merchant.git
cd dv-merchant

make update-frontend
make build
```

The binary `github.com/dv-net/dv-merchant` will appear in `.bin/` once the build finishes.

---

## 🧪 Development & Testing

**Pre-commit checklist**
- Run linting and formatting to keep the codebase consistent.
- Execute unit tests and make sure critical flows are covered.
- Add or update tests when shipping new features or fixes.

```bash
# Static analysis & formatting
make lint
go fmt ./...

# Unit tests
make test
```

> ℹ️ Extended workflows (`make run`, Docker Compose, etc.) are documented in the [`dv-bundle`](https://github.com/dv-net/dv-bundle) repo (`README.md`) and on https://docs.dv.net.

---

## 🛠 CLI Commands

- `.bin/dv-merchant start` — run the HTTP API server.
- `.bin/dv-merchant migrate up|down` — apply or roll back DB migrations.
- `.bin/dv-merchant seed up|down` — load or drop seed data.
- `.bin/dv-merchant config` — validate config and generate env/flags.
- `.bin/dv-merchant permission` — manage roles and Casbin policies.
- `.bin/dv-merchant transactions` — tooling for transaction operations.
- `.bin/dv-merchant users` — manage users from the console.

---

## 📚 Documentation

- 📖 [Full guide](https://docs.dv.net) — installation, configuration, scenarios.
- 🔌 [API reference](https://docs.dv.net/en/operations/post-v1-external-wallet.html) — request/response schemas.
- 🧾 [Swagger spec](docs/swagger.yaml) — shipped with the repository.

---

## 🔐 Security Features

1. 🔓 Non-custodial design — you control keys and addresses.
2. 🧠 Multisig support and TRON resource delegation.
3. 🛡️ Casbin RBAC with flexible `configs/rbac_*` policies.
4. 📜 Full audit trail: events, logging, Prometheus metrics.

---

## 🤝 Contributing

```bash
# Before submitting a PR
make lint
go test ./...
```

- ⭐ Star the repo if it helps your project.
- 🐛 Report bugs via Issues.
- 💡 Propose new features and use cases.
- 🔧 Pull Requests are welcome!

---

## 💝 Donations

Support the development of the project with crypto:

> <img src="docs/assets/icons/coins/IconUsdt.png" width="17"> **USDT (Tron)** — `TCB4bYYN5x1z9Z4bBZ7p3XxcMwdtCfmNdN`

> <img src="docs/assets/icons/coins/IconBtcBitcoin.png" width="17"> **Bitcoin** — `bc1qemvhkgzr4r7ksgxl8lv0lw7mnnthfc6990v3c2`

> <img src="docs/assets/icons/coins/IconTrxTron.png" width="17"> **TRON (TRX)** — `TCB4bYYN5x1z9Z4bBZ7p3XxcMwdtCfmNdN`

> <img src="docs/assets/icons/coins/IconEthEthereum.png" width="17"> **Ethereum** — `0xf1e4c7b968a20aae891cc18b1d5836b806691d47`

🔗 Other networks and tokens (BNB Chain, Arbitrum, Polygon, Litecoin, Dogecoin, Bitcoin Cash, etc.) are available at **[payment form](https://cloud.dv.net/pay/store/208ec77f-d516-46b9-b280-3c12e1a75071/donate)**

---

## 📞 Contact

<div align="center">

**Telegram:** [@dv_net_support_bot](https://t.me/dv_net_support_bot) • **Telegram Chat:** [@dv_net_support_chat](https://t.me/dv_net_support_chat) • **Discord:** [discord.gg/Szy2XGsr](https://discord.gg/Szy2XGsr)

**Email:** [support@dv.net](https://dv.net/#support) • **Website:** [dv.net](https://dv.net) • **Documentation:** [docs.dv.net](https://docs.dv.net)

</div>

---

<div align="center">

**© 2026 DV.net** • [DV Technologies Ltd.](https://dv.net)

*Built with ❤️ for the crypto community*

</div>
