<div align="center">

## 🚀 DV.net Merchant Backend
<br>

[🇬🇧 English](../README.md) • [🇷🇺 Русский](README.ru.md) • [🇨🇳 中文](README.zh.md)

[Веб-сайт](https://dv.net) • [Документация](https://docs.dv.net) • [API](https://docs.dv.net/en/operations/post-v1-external-wallet.html) • [Поддержка](https://dv.net/#support)

</div>

---

## 💡 О проекте

**DV.net Merchant Backend** — высоконагруженная back-end платформа для приёма, обработки и отправки криптовалюты. Система полностью открытая, self-hosted и позволяет контролировать каждый аспект платежей без посредников и скрытых комиссий.

> 🔒 **Некастодиально** — приватные ключи остаются на вашей стороне
>
> ⚡ **Производительно** — Go 1.24, Fiber v3, PostgreSQL & Redis
>
> 🌐 **Широкая поддержка** — множество блокчейнов и централизованных бирж
>
> 🧱 **Модульно** — чистая архитектура: delivery → service → storage

---

## ✨ Возможности

**🎯 Бизнес-функции**
- ✅ Приём и отправка криптовалют без KYC/KYB
- ✅ Нотификации, вебхуки и гибкая маршрутизация событий
- ✅ Управление комиссиями, оптимизация ресурсов TRON и EVM
- ✅ Интеграции с CEX (Binance, OKX, HTX, KuCoin, Bybit и др.)

**🔧 Технические особенности**
- ✅ HTTP API на базе Fiber v3 и Casbin RBAC
- ✅ Асинхронные обработчики и планировщики внутри `internal/app`
- ✅ Слой сервисов с DI и бизнес-логикой (`internal/service`)
- ✅ Репозитории поверх PostgreSQL и Redis (`internal/storage`)
- ✅ Автоматическая генерация SQL (`sqlc`, `pgxgen`)
- ✅ Набор вспомогательных пакетов в `pkg` (клиенты, retry, OTP, AML)

---

## 🧭 Архитектура

```text
cmd/                CLI и точки входа (сервер, миграции, утилиты)
configs/            Шаблоны конфигов и Casbin политики
internal/app        Инициализация приложения, фоновые задачи
internal/delivery   HTTP-эндпоинты, middleware, маршрутизация
internal/service    Бизнес-логика, интеграции, события
internal/storage    Репозитории PostgreSQL/Redis, миграции
pkg/                Внешние клиенты и общие библиотеки
sql/                SQL-модули, миграции и генерация кода
```

Диаграммы и Swagger доступны в каталоге `docs/` (`swagger.yaml`, `swagger.json`).

---

## 🚀 Запуск

**Self-hosted установка (1 команда)**
```bash
sudo bash -c "$(curl -fsSL https://dv.net/install.sh)"
```

**Docker-бандл для разработки**
```bash
git clone --recursive https://github.com/dv-net/dv-bundle.git
cd dv-bundle && cp .env.example .env
docker compose up -d
```

**Локальная сборка бекенда**
```bash
git clone https://github.com/dv-net/dv-merchant.git
cd dv-merchant

make update-frontend
make build
```

После сборки бинарь `github.com/dv-net/dv-merchant` появится в `.bin/`.

---

## 🧪 Разработка и тестирование

**Как проверять изменения перед коммитом**
- Запустите линтеры и форматирование, чтобы убедиться в едином стиле кода.
- Выполните модульные тесты, убедитесь, что покрыты критичные сценарии.
- При необходимости дополните тестами новые фичи или багфиксы.

```bash
# Статический анализ и форматирование
make lint
go fmt ./...

# Модульные тесты
make test
```

> ℹ️ Расширенные сценарии (`make run`, Docker Compose и т.п.) описаны в репозитории [`dv-bundle`](https://github.com/dv-net/dv-bundle) (файл `README.md`) и на https://docs.dv.net.

---

## 🛠 CLI команды

- `.bin/dv-merchant start` — запуск HTTP API.
- `.bin/dv-merchant migrate up|down` — миграции БД.
- `.bin/dv-merchant seed up|down` — начальные данные.
- `.bin/dv-merchant config` — валидация конфигов и генерация env.
- `.bin/dv-merchant permission` — управление ролями и политиками.
- `.bin/dv-merchant transactions` — инструменты по операциям.
- `.bin/dv-merchant users` — управление пользователями.

---

## 📚 Документация

- 📖 [Полный гайд](https://docs.dv.net) — установка, настройка, сценарии.
- 🔌 [API Reference](https://docs.dv.net/en/operations/post-v1-external-wallet.html) — схемы запросов и ответы.
- 🧾 [Swagger](swagger.yaml) — доступен в репозитории.

---

## 🔐 Безопасность

1. 🔓 Некастодиальная модель — ключи и адреса контролируете вы.
2. 🧠 Поддержка мультисиг и делегирования ресурсов TRON.
3. 🛡️ RBAC на Casbin + гибкие политики в `configs/rbac_*`.
4. 📜 Полный аудит: события, логирование, прометей-метрики.

---

## 🤝 Вклад в проект

```bash
# Перед PR
make lint
go test ./...
```

- ⭐ Поставьте звезду репозиторию, если проект полезен.
- 🐛 Сообщайте об ошибках через Issues.
- 💡 Предлагайте новые фичи и сценарии.
- 🔧 Pull Requests приветствуются!

---

## 💝 Donations

Поддержите развитие проекта криптовалютой:

> <img src="assets/icons/coins/IconUsdt.png" width="17"> **USDT (Tron)** — `TCB4bYYN5x1z9Z4bBZ7p3XxcMwdtCfmNdN`

> <img src="assets/icons/coins/IconBtcBitcoin.png" width="17"> **Bitcoin** — `bc1qemvhkgzr4r7ksgxl8lv0lw7mnnthfc6990v3c2`

> <img src="assets/icons/coins/IconTrxTron.png" width="17"> **TRON (TRX)** — `TCB4bYYN5x1z9Z4bBZ7p3XxcMwdtCfmNdN`

> <img src="assets/icons/coins/IconEthEthereum.png" width="17"> **Ethereum** — `0xf1e4c7b968a20aae891cc18b1d5836b806691d47`

🔗 Другие сети и токены (BNB Chain, Arbitrum, Polygon, Litecoin, Dogecoin, Bitcoin Cash и др.) доступны в **[форме оплаты](https://cloud.dv.net/pay/store/208ec77f-d516-46b9-b280-3c12e1a75071/donate)**

---

## 📞 Контакты

<div align="center">

**Telegram:** [@dv_net_support_bot](https://t.me/dv_net_support_bot) • **Чат в Telegram:** [@dv_net_support_chat](https://t.me/dv_net_support_chat) • **Discord:** [discord.gg/QUXNWZdy](https://discord.gg/QUXNWZdy)

**Email:** [support@dv.net](https://dv.net/#support) • **Сайт:** [dv.net](https://dv.net) • **Документация:** [docs.dv.net](https://docs.dv.net)

</div>

---

<div align="center">

**© 2026 DV.net** • [DV Technologies Ltd.](https://dv.net)

*Сделано с ❤️ для криптосообщества*

</div>
