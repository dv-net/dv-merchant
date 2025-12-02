# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## [0.9.14] - 2025-13-04

- feat: added extended tx info in search response [DV-3827]
- feat: added endpoint to retrieve extended information about store enabled currencies [DV-3857]

## [0.9.13] - 2025-11-19
- Fix `FindLastWalletTransactions` sql query  [DV-3808]

## [0.9.12] - 2025-11-18
- Fix transaction ordering at `tx-find` method  [DV-3808]
- The readme has been updated and translated into Russian and Chinese [DV-3708]

## [0.9.11] - 2025-10-15
- fix: fixed universal groups separation by blockchain [DV-3645]
- fix: improve wallet address retrieval with mutex locking and retry logic
- refactor: AML validation errors (unsupported currency, invalid address) [DV-2896]
- fix: KuCoin minFunds incorrect behaviour on spot order creation [DV-3611] 
- fix: ByBit exchange balance duplicates causing incorrect total_usd calculation
- update: Frontend update
- feat: add endpoint to retrieve unconfirmed transfer transactions [DV-3733]


## [0.9.10] - 2025-10-15
- fix: incorrect Binance spot order rules calculation [DV-3598]
- fix: incorrect behaviour of logging [DV-3593]
- fix: encode analytics data into byte array before forwarding to KV cache [DV-3599]

## [0.9.9] - 2025-10-14
- fix: panic in Binance GetOrderRule when handling USDT pairs [DV-3558]
- feat: added proper logging for exchange clients and services [DV-3088]
- fix: added variation of KuCoin error on withdrawal confirmations lock [DV-3545]

## [0.9.8] - 2025-10-13
- fix: add error handling for empty withdrawal addresses and improve logging format [DV-3437]
- Added currency reference data about it being native, and contract address [DV-3456]
- Added new currencies to the supported currency pool (disabled by default for new stores) [DV-3319]
- Added new currency codes and enhance stablecoin handling in rate calculations [DV-3476]
- feat: update user context handling in wallet and address services [DV-3480]
- get minimal rate for currency pairs if not found for source [DV-3500]
- fix: handle htx rate limiting error [DV-3505]
- fix: locale forwarding on external endpoint [DV-3277]
- fix: removed duplicated ordering index [DV-3501]
- fix: fixed exchange withdrawal settings state restoration [DV-2945]
- fix: fixed exchange keys deletion behaviour [DV-3343]
- fix: changed exchange state SQL queries to correctly handle withdrawals/spot market orders [DV-2935]
- fix: sort currency order for form without api key [DV-3529]
- chore: reuse errors across multiple exchange clients [DV-3544]
- fix: exclude unconfirmed transactions for tx-find [DV-3528]

## [0.9.7] - 2025-09-22
- Fix rename ResetPasswordCode to Code for consistency in user notifications [DV-3403]
- enhance locale handling in wallet creation and update logging for system email [DV-3426]

## [0.9.6] - 2025-09-19
- Add validation for withdrawal exchange wallet by blockchain [DV-3326]
- Add log memory buffer [DV-3359]
- Fixed bug with withdrawals rules not being created on clean install via address book [DV-3339]
- Added per store settings API [DV-3278]
- Fix create store on registration user [DV-3392]
- Add Dockerfile [DV-3387]

## [0.9.5] - 2025-09-15
- Fix validation withdrawal exchange [DV-3362]
- Add last processing logs endpoint and update log response structure [DV-3379]

## [0.9.4] - 2025-09-15
### Note
This is the first public release of the project.  
All previous versions were private/internal and were never published.

- Added behavior handling for handling store being disabled [DV-3117]
- Accept browser locale from frontend and send emails to payers based on specific locale with fallback to user language or English [DV-3086]
- Added single column sorting in transaction history [DV-2480]
- Fixed request validation [DV-3077]
- Added blockchain icons into API response [DV-2978]
- Added currency labels to API response [DV-3128]
- Prevent sending wallet notification for disabled stores [DV-3170]
- Fixed mailer state setting immutability [DV-3166]
- Added address book API [DV-3069]
- Split API into single tags [DV-2578]
- Reworked private key file exporting [DV-3142]
- Added user crypto receipt notification type to notification types API [DV-3176]
- Fixed query bug with wallet address filtering in hot wallet summary [DV-3176]
- Fixed hot wallet paginated query ordering issue when sorted by balance [DV-3137]
- Enhance address book service to support TOTP for withdrawal rules and processing whitelist updates [DV-3191]
- Address book service includes withdrawal rule status in address responses [DV-3192]
- Updated email change request to accept string for verification code [DV-3194]
- Added locale migration for wallets [DV-3210]
- Sort wallet addresses by blockchain and native token in email notification [DV-3174]
- Mask response with CEX API keys [DV-3332]
- Add Handler for confirm selected wallet currency[DV-3354]

## [0.6.1] - 2025-08-11
- Internal/External API withdrawal route now takes a random store ID related to the user for internal requests, and uses the store ID from the XApiKey header for external requests. [DV-2704]
- Init withdraw from processing setting [DV-2666]
- Adds a new endpoint to retrieve notification history with pagination and filtering options [DV-2744]
- Fix validation numeric withdrawal from processing [DV-2683]
- Error message on transfer request [DV-2680]
- AML integration (BitOK, AMLBot) [DV-2537]
- Tron expenses status by hour [DV-2708]
- Transaction stats rework [DV-2747]
- Fixed minimum withdrawal amount calculation [DV-2792]
- Take out root setting merchant_pay_form_domain on dictionary [DV-2861]
- AmlBot risk level conversion fixed [DV-2855](https://mtask.org/issue/DV-2855)
- Refactoring internal method for make wallet pay form [DV-2869]
- Change swag annotation for params [DV-2859]
- Fix bug update use location [DV-2859]
- Fixed bug with notification history cleanup [DV-2886]
- Fix notification history filtering by user email [DV-2880]
- AML Keys check before update [DV-2808]
- Added new notification types to SMTP mailer [DV-2897]
- Added API for notification types retrieval [DV-2906]
- Processing wallets info fix [DV-2914]
- Processing wallets duplicates fix [DV-2894]
- Statistics timezones fix [DV-2907]
- Fixed bug with ByBit client not receiving deposit rules [DV-2926]
- Fixed bug with HTX client returning new error on locked balance [DV-2949]
- Added exchange ticker and chain labels for withdrawal page [DV-2988]
- Relate notification history to stores and users [DV-2919]
- Added L2 transfer estimation for EVM chains [DV-2820]
- TG unlink API [DV-1379]
- Adds anonymous telemetry to collect system information and transfer statistics [DV-2781](https://mtask.org/issue/DV-2781)
- Prevent application crash on nil order details from exchange [DV-2994]
- Git tags creation with same commit hash fixed [DV-2557]
- Traceback enabled by default in systemd [DV-3013]
- AML Attempts  [DV-2994]
- Fail KuCoin orders that are older then API limits [DV-3015]
- Add new notification type for user email change [DV-2985]
- Prevent validation error on user rate source update [DV-3024]
- Fixed bug with withdrawal rules not being applied correctly when changing current exchange [DV-3022]
- Added currency rates sorting [DV-3018]
- Added custom ctr linter [DV-3014]
- Sort exchanges by name in API [DV-3049]
- Migrated to external mustache templates [DV-3032]
- Added exclusion to multi-withdrawal API [DV-2893]
- Remove sending a letter of verification of the user during registration [DV-3055]
- Added cache for KuCoin exchange public API [DV-3060]
- Fix count_with_balance on hot wallets [DV-2911]
- Added KuCoin system rate limit handling [DV-3099]
- Added tx search by untrusted email [DV-2813]
- Added new receipt template support [DV-2569]

## [0.6.0] - 2025-07-17
- Import user mnemonic [DV-2349]
- Static dir file for frontend [DV-2800]

## [0.5.7] - 2025-07-14
- Dogecoin integration [DV-2457]
- Arbitrum integration [DV-2706]
- Arbitrum validation fix [DV-2718]
- Added internal API for withdrawal from processing wallets [DV-2704]
- Delete withdrawal from processing request [DV-2729]
- Add transaction untrusted email [DV-2742]

## [0.5.6] - 2025-07-03
- Fixed withdrawal rules amount validation [DV-2288]
- Added proper error handling for trading symbol being halted [DV-2474]
- Fix email tls smtp server [DV-2467]
- Fixed KuCoin spot order filled amount retrieval [DV-2475]
- Fixed manual withdraw err msg [DV-2213]
- Added initial exchange connection timestamp retrieval [DV-2289]
- Allowed unknown fields in config [DV-2099]
- Filter out offline exchange symbols [DV-2417]
- Changed verification email template [DV-2395]
- Send notification on user password change [DV-2492]
- Added migrations for OKX USDC delisting [DV-2505]
- Add timezone for user stats [DV-2421]
- Added additional information about dust balances on hot wallets [DV-2388]
- Transaction stats by date [DV-1979]
- Added GATE.IO exchange support [DV-1651]
- Multi select project currency [DV-2543]
- Added exclusion in hot, withdrawal wallet addressess [DV-2597]
- Added public payment form API untrusted email [DV-2548]
- Fixed bug with balances filtering by currency [DV-2586]
- Select summery with dust balances [DV-2595]
- Tron expenses statistics [DV-2460]
- Sort order currency on withdrawal rules Blockchain + code [DV-2443]
- Fixed filtering processing assets on external, internal API [DV-2586]
- Fixed comparison operators for amount and balance filters [DV-2513]
- Save untrusted email on public API request [DV-2642]
- Fixed tron statistics [DV-2669]
- Fixed transfers from processing prefetch [DV-2678]
- Added exposed root settings filtering [DV-3073]

## [0.5.5] - 2025-06-17
- Handle errors caused by user security action on exchange (Bitget) [DV-2331]
- Added support for exchange API testing without saving [DV-2296]
- Replaced Telegram support bot handle [DV-2373]
- Fixed KuCoin address generation [DV-2413]
- Transfer system transactions persist [DV-2031]

## [0.5.4] - 2025-06-05
- Update client email with wallets create [DV-2058]
- Sort processing wallets [DV-2198]
- Mailer settings update fix [DV-2280]
- Wallets create amount remove from response [DV-1671]
- Swagger fix [DV-923]

## [0.5.3] - 2025-05-27
- Dv rates source [DV-1919]
- Sort processing wallets [DV-1931]
- Multiple withdrawal for by low-balance rules [DV-1738]
- Currency rate for stablecoin [DV-2068]
- Added unique hash for exchange connection [DV-2050]
- Turnstile verification added [DV-2043]
- Binance withdrawal functionality fixed [DV-2101]
- Force transfer suspension on unknown exchange errors [DV-1966]
- Fix min balance in wallet balances API [DV-2523]

## [0.5.2] - 2025-04-22
- App versions with no cache API method [DV-1861]
- Archive store rework [DV-1846]
- Deposit statistics amount USD by currency [DV-1772]
- Added EVM approximate transfer count estimation [DV-1686]
- Add Polygon [DV-1885]
- Added approximate TRON transfer estimation data [DV-1890]
- Deposit stats fix [DV-1911]
- User transaction fix [DV-1887]
- Explorer links added for currencies [DV-1929]

## [0.5.1] - 2025-04-22

- App versions with no cache API method [DV-1715]
- Fixed incorrect amount of deposit addresses on exchanges [DV-1713]
- Added withdrawal disabling [DV-1660]
- Added quick-start root setting [DV-1662]
- Add profile start app [DV-1710]
- 2fa auth for settings added [DV-1668]
- Removed week statistics [DV-1675]
- Store archive API [DV-1363]
- Fix logger error [DV-1787]
- Wallets balance calc optimization [DV-1248]
- Added CSV, XLSX export for user financial actions reports [DV-1812]
- Eproxy v2 [DV-1785]
- BCH converter address [DV-1565]
- Fix updater connection error [DV-1824]
- Remove transaction [DV-1131]
- Fixed BitGet spot order submittion [DV-1853]
- Fix create address db-transaction [DV-1834]
- Fixed BitGet spo order submittion [DV-1853]
- Added proper HTX withdrawal handling [DV-1836]
- Archive store rework [DV-1846]

## [0.5.0] - 2025-04-10

- Renamed github.com/dv-net/dv-merchant to github.com/dv-net/dv-merchant  [DV-1500]
- Added proper error handling for exchange credentials testing [DV-1396]
- Hide dashboard balances below certain threshold [DV-1541]
- Replaced templater with mustache implementation [DV-1549]
- Added user email verification reminder [DV-1536]
- Updated API route for telegram link generation [DV-1597]
- Added filters to processing balances API [DV-1599]
- Remove unique request_id [DV-1705]

## [0.4.11] - 2025-03-31

- Fix withdrawal from processing wallet 

## [0.4.10] - 2025-03-28

- Change user password cmd tool [DV-1537]
- Integrate DAI [DV-1576]

## [0.4.9] - 2025-03-23

- Change user email cmd tool [DV-1257]
- Added stable repo publish [DV-955]
- Fix user settings [DV-1376]
- User notifications API [DV-1148]
- Updater [DV-1199]
- OKX, BitGet withdrawal retry with reduced amount on failure [DV-1428]
- Return retry reason in withdrawal history [DV-1511]
- Notifications batch update [DV-1470]
- Fix notification seed [DV-1451]
- Date filter added [DV-1384]
- Rounding txs amount [DV-1325]
- DV-updater dependency removed [DV-1374]
- Transfers prefetch fix [DV-1431]
- Fixed BitGet sub-client incorrect request assembly [DV-1523]
- Fixed backend IP retrieval [DV-1527]
- Fixed BitGet exchange order native and fiat retrieval [DV-1524]

## [0.4.8] - 2025-03-17

- Added transfer history handler [DV-888]
- Added text address private keys download option [DV-1247]
- Manual deletion of failed transfers [DV-1245]
- Exchange rule wallets balance sum [DV-1236]
- DV-updater dependency added [DV-1327]
- Field validator fix [DV-1303]
- Rate limits fix [DV-1170]
- Shop site added in store create route [DV-1136]
- Dev deploy fix [DV-1240]
- Global store webhook secret [DV-1149]
- Added Bitget exchange support [DV-1266]
- Added Binance executed order USD equivalent calculation [DV-1388]
- Fixed exchange withdrawal error handling bug [DV-1382]

## [0.4.7] - 2025-02-25

- Manual deletion of failed transfers [DV-921]
- Remember me [DV-612]
- Transfer step added [DV-639]
- Remove v3 routes [DV-743]
- Owner dvadmin auth [DV-1037]
- Auto disable transfers by failed withdrawal from exchange [DV-1083]
- Withdrawal settings toggle flag [DV-1025]
- Prevent unit-file replacement [DV-1120]
- External withdrawal from processing [DV-1130]
- Transfer tx_hash added [DV-1123]
- User cred verifications rework [DV-852]
- Tx and wallet search generic route added [DV-985]
- Extended transfers list route [DV-1150]
- Multi login fix [DV-1158]
- Deposit summary amount USD [DV-1195]
- BCH support [DV-1206]

## [0.4.6] - 2025-02-17

- Change redirect if system not install [DV-935]
- Sort summery balance [DV-1013]
- Fix admin routes permissions [DV-988]
- Fix store update [DV-874]
- Transaction type from processing removed [DV-699]
- Hot wallet transactions restore [DV-42]
- Multiple transfers [DV-906]
- Min amount usd for auto transfers config added [DV-659]
- Admin authentication rework [DV-1011]
- Fix webhooks duplicates [DV-1092]

## [0.4.5] - 2025-02-01

- Fixed CSV formatting for downloading user's hot wallet keys [DV-843]
- Added new transfer type 'cloud_delegate' [DV-703]
- Fix email send for get address [DV-905]
- Add log level [DV-586]
- Processing url setting fixed [DV-984]
- Log address activity download key [DV-482]
- Backend shutdown fix [DV-978]
- WH queue sync fixed [DV-781]

## [0.4.4] - Unreleased

## [0.4.3] - Unreleased

## [0.4.2] - 2025-01-23

- Update frontend build

## [0.4.1] - 2025-01-23

- Fix recalculate balance [DV-777]
- Email verification refactored [DV-757]
- Added new API method for hot wallet keys downloading [DV-460]
- Total used energy and bandwidth info for tron wallet [DV-776]
- Fix search hot wallet query [DV-775]
- Non-root executable [DV-786]
- System transactions [DV-446]
- Top up data by store [DV-599]

## [0.4.0] - 2025-01-17
- Wallet Address Activity log [DV-393]
- Dv-admin routes segregated [DV-709]
- Add tx_id on webhook history [DV-742]
- Make Processing list route public [DV-702]
- Move to home dir [DV-697]
- Withdrawal from processing webhook type [DV-392]

## [0.3.1] - 2024-12-26

- Integrated Binance exchange [DV-627]
- Move stage server [DV-561]
- Move dev server [DV-560]
- Add store rate scale [DV-30]

## [0.3.0] - 2024-12-16

- Add store info in find address [DV-452]
- Fixed deposit address request [DV-461]
- Update notional USD equivalent of exchange order [DV-396]
- Saving IP address on external wallet creation [DV-372]
- Remove unsupported pairs from rate response [DV-225]
- Merchant registration in admin service [DV-121]
- Added backend IP address response [DV-397]
- Added filters to hot wallet search [DV-379]
- Added additional data about tron resources [DV-379]
- Send token from hot-wallet to processing wallet on the backend [DV-304]
- Fixed user info object timestamp [DV-490]
- Added cold wallet balance retrieval with fixes [DV-490]
- Monitors removal [DV-404]

## [0.2.9] - 2024-12-04

- Added rate-limiting to outbound requests to exchanges [DV-298]
- Added response types for HTX clients (same as OKX) [DV-324]
- Unconfirmed deposit tx from mempool support [DV-227]
- Wallet id param added [DV-195]
- Manual withdraw from hot wallet [DV-221]
- Rework address response and add filter transaction by blockchain [DV-366]
- Save withdrawal settings in native token [DV-311]
- Add rate on withdrawal rules [DV-435]

## [0.2.8] - Unreleased

## [0.2.7] - 2024-11-22

- Added exchange chains API [DV-283]
- Fixed display of user exchange pairs [DV-244]
- Set default sender [DV-248]
- Log with monitor [DV-187]
- Remove stats last week [DV-252]
- Added exchanges withdrawals, pair swap, and exchange rate sources [DV-19]
- Remove header [DV-302]
- In requests to send notifications from dv.net, pass the domain in the \_backend_domain field [DV-267]
- Fix email users not update [DV-323]
- Save error messages from processing transfer status [DV-314]

## [0.2.6] - 2024-11-14

- Added handler for deposit address from exchange [DV-173]
- Make currency route is publish [DV-159]
- Added support version v1 and v3 [DV-186]
- Settings sort added [DV-127]
- Setting is_editable param added [DV-16]
- Search wallet transactions rework [DV-209]

## [0.2.5] - 2024-11-8

- Added available exchange pairs API [DV-2719]
- Added exchange rules API [DV-2697]
- Added exchange rate sources to API [DV-2702]
- Changes to transfer API [DV-2655]
- Added HTX exchange rate source [DV-2570]
- Added OKX exchange rate source [DV-2569]
- Added external processing balance API [DV-2601]
- Added external exchange balance API [DV-2602]
- Logger refactoring [DV-2588]
- Added monitors API and business logic [DV-2582]
- Removed strict validation on store update request [DV-2596]
- Remove go mod tidy from build [DV-2265]
- Remove frontend from filesystem [DV-2264]
- Rsync logs as defaults [DV-2409]
- Processing withdrawal rework [DV-2591]
- Address validation for processing withdrawal added [DV-2595]
- Fix processing withdrawal creation while transfers disabled [DV-2592]
- User transaction filter for min amount usd added [DV-2635]
- Fix webhook send by store available currency [DV-2638]
- Setting refactoring [DV-2621]
- Fix webhook delay increase with i/o errors [DV-2667]
- Dv-net notification sender added [DV-2442]
- Notifier refactoring [DV-2617]
- WH history sort for tx-find [DV-2629]
- External currencies rate API method added [DV-2600]
- Public and external tags added to swagger [DV-2650]
- Withdrawal from processing external info added [DV-2603]
- Notifications hotfix [DV-2720]
- Hide processing wallets by target blockchains [DV-2627]
- Request body for send test wh added [DV-2640]
- USDC Tron support disabled [DV-2641]
- Store user email with external wallet creation [DV-2718]
- Fix checking processing response code [DV-2723]
- Hotfix mailer init [DV-2744]

## [0.2.4] - Unreleased

## [0.2.3] - Unreleased

## [0.2.2] - 2024-10-24

- Added user invite logic [DV-2181]
- Withdrawal wallet addresses list fix added [DV-2394]
- Refactored mailing service to support persistent history, retries, multiple channels [DV-2396]
- Added Cache-Control header for Cloudflare [DV-2464]
- Added currently used and balance API for exchanges [DV-2436]
- Added blockchain info to errors [DV-2478]
- Uniq withdrawal_wallet_addresses fix [DV-2511]
- Prefetch transfers data fix [DV-2460]
- Fix distinct address for cold wallet processing [DV-2510]
- Withdrawal from processing wallet added [DV-2403]
- Dictionaries API method added [DV-2375]
- Webhook history API [DV-2446]
- Get processing wallets info refactoring [DV-2479]
- Processing callback domain update API method added [DV-2185]
- API For store webhook emulate added [DV-2474]
- Manual wh send fix [DV-2537]
- Webhook check if success by body [DV-2553]
- Processing withdrawal native currency setting removed [DV-2548]
- Fix transfers created_at [DV-2558]
- Fix webhook_send_history [DV-2749]

## [0.2.1] - 2024-10-15

- Withdrawal wallet addresses list fix added [DV-2394]
- Added cold wallet addresses total balance API [DV-2365]
- Public tx find amount_usd hotfix [DV-2369]
- Unconfirmed transactions created_at fix added [DV-2390]
- Save message if have failed transfer [DV-2400]
- Change logic transfer send [DV-2399]
- Resend transfer if not resource [DV-2411]
- Receipt id in transaction/{hash} fix [DV-2423]
- Processing cold wallet creation fix added [DV-2303]
- Fix tron processing wallet resources fetch [DV-2393]
- Fix store update validation [DV-2280]
- Amount & currency parameters for pay URL [DV-2420]
- Fix store webhook queue increment delay [DV-2455]
- Find wallet info with transactions method added [DV-2348]
- Currency rates API rework [DV-2433]
- Withdrawal addresses validation added [DV-2457]

## [0.2.0] - 2024-10-09

- Added hot wallet addresses total balance API [DV-2364]
- Public tx find amount_usd hotfix [DV-2369]
- Update settings API added [DV-2358]
- Custom huobi API client [DV-2356]
- Fix webhook url [DV-2369]
- Public tx find amount_usd hotfix [DV-2368]
- Check migrations before merging to stage branch [DV-2343]
- Check if db schema is ready before service start [DV-2348]
- External route for store currencies list [DV-2346]
- Moved mailer settings to database [DV-2293]
- Added new settings for payment form, merchant domain [DV-2289]
- Notification service reworked [DV-2308]
- Collapse unconfirmed transactions worker [DV-2333]
- Setting for type transfer [DV-2350]
- Update native token for deposit webhook [DV-2345]
- Add okx test method [DV-2286]
- Manual handler for webhooks [DV-2278]
- User settings route [DV-2336]
- Payform urls [DV-2338]
- Unified responses across handlers [DV-2291]
- Rollback migrations for casbin [DV-2251]
- Combined all migrations into single unit [DV-2269]
- Removed unnecessary data from transaction search response [DV-2282]
- Fully refactored generation of Swagger API documentation [DV-1936]
- Added registration state changing functionality [DV-2204]
- Patch withdrawal addresses validation hotfix [DV-2266]
- Exchange integration fix [DV-2284]
- Callback domain trailing slash hotfix [DV-2300]
- TX_info by hash API method added [DV-2237]
- Wallet info added in tx_info [DV-2247]
- User stores relation added [DV-2294]
- Withdrawal to address only one at a time on the blockchain [DV-2328]
- Public API for find recent wallet transactions [DV-2311]
- Transfers mode disable API added [DV-2177]
- Store currencies duplicates removed [DV-2299]
- Exchange errors formatting [DV-2329]

## [0.1.5] - 2024-09-30

- Update go-releaser [DV-2227]
- Reformatting config [DV-2156]
- Fix root user wallets route [DV-2202]
- Fix coingate response test [DV-2159]
- Use in-memory storage if redis is unavailable [DV-2157]
- Added deposit statistics API [DV-2121]
- Changed wallet balance API response [DV-2222]
- Changed user auth API response [DV-2218]
- Sign added for processing requests [DV-2052]
- Static files embedding [DV-2183]
- Added prometheus, health, pprof metrics [DV-2200]
- GET route for external wallet creation adde [DV-2150]
- Stage column for transfers added [DV-2231]
- Deposit statistics sort date desc added [DV-2230]
- API for manual withdrawal [DV-2201]
- HTX exchange integration added [DV-2165]
- Parameter backend domain for processing callback url added [DV-2184]
- Fixed root user registration in uninitialized processing [DV-2250]
- Update withdrawal wallets list API added [DV-2241]
- Callback domain empty string supported + custom url validation added [DV-2258]
- Seeds embedded to binary [DV-2253]

## [0.1.4] - 2024-09-18

- Added root/admin API methods [DV-2168]
- Two-factor authentication added to withdrawal wallets [DV-2138]
- Withdrawal wallets status as enum [DV-2139]
- Add transfer logic and transfer API [DV-2133]
- Add frontend to go-releaser [DV-2163]
- Domain layer is not depends on redis [DV-2144]
- Added processing wallet assets API [DV-2111]
- Added user transaction response pagination object [DV-2175]
- Added hot wallets summary API [DV-2119]

## [0.1.3] - 2024-09-11

- Add go-releaser [DV-2147]
- Management withdrawal wallet and address [DV-2022]
- User transactions filter added [DV-2126]
- Email verification [DV-1868]
- Password reset mechanism [DV-1861]
- Multi-language mailing support [DV-1860]
- Added email notifications implementation with future support for Telegram [DV-1859]
- Email notifications refactored, added i18n support [DV-1866]

## [0.1.2] - 2024-09-02

- Fix db pool connection in backend [DV-2098]
- Root user updates [DV-2043]
- Host added for processing callback url [DV-2095]
- Api for wallets with balance [DV-2003]
- Casbin flow updates [DV-2001]
- Casbin change adapter [DV-2095]
- Updated seeds (USDC) [DV-2106]

## [0.1.1] - 2024-08-28

- Remove NATS and change logic retry webhook [DV-1877]
- Public API for payment form [DV-1883]

## [0.1.0] - 2024-08-22

- Deposit processing [DV-1873]

## [0.0.3] - 2024-08-20

- Update processing API [DV-2042]
- Added new API method for wallet balances [DV-1989]
- Added caching for settings [DV-1793]
- Added new API methods for web-installer [DV-1829]
- Added body/param x-api-key support [DV-1929]

## [0.0.2] - 2024-08-13

- Added dev ci-cd [DV-1899]
- Frontend initialized. [DV-1717]
- Database seed tool. [DV-1569]
- Source of exchange rates service. [DV-1571]
- Currency conversion service. [DV-1573]
- GitLab pipeline. [DV-1578]
- Notification service. [DV-1581]
- Event listener. [DV-1582]
- Role-based access model. [DV-1589](
- CLI `config` command. [DV-1602]
- Send webhook to merchants service. [DV-1593]
- Database migrations tool. [DV-1603]
- 2FA API fixes. [DV-1918]
- Added TRX hot wallets generation [DV-1839]
- Added BTC hot wallets generation [DV-1843]
- Dirty wallets logic implemented [DV-1967]
- Fixed ID mismatch [DV-1952]
- Added new API methods for whitelists [DV-1984]
- Added API methods for seeds, private keys [DV-1846]
- Added new API method for webhook [DV-1906]
- Added new API method for whitelists, changed store response [DV-1988]
- Fixed store creation method returning nil data [DV-1987]

## [0.0.1] - 2024-07-01

- Initial commit
