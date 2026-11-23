-- Irreversible data fix: tokens are not turned back into never-expiring ones.
SELECT 1;

fix(security): address findings from external security report

  Authorization (IDOR/BOLA):
  - wallet: POST /wallet returns only wallets of stores the user has access to; client-supplied store_ids are intersected with user_stores
  - wallet: POST /wallet/restore/:walletId checks store access; foreign and missing wallets both return 404 (no UUID existence oracle)
  - receipts: GET /receipts/:id is scoped to the user's stores (404 otherwise)
  - webhook: POST /webhook/send-test checks access to the webhook's store
  - address book: DELETE of a simple address checks entry ownership
Authentication:
  - register response uses an explicit DTO; password hash, remember_token and dvnet_token are no longer returned
  - password reset and console password change revoke all user tokens, reset runs password change and revocation in one transaction
  - auth tokens always expire: 24h, 30 days with remember_me; migration bounds existing never-expiring tokens to 30 days
  - add POST /auth/logout revoking the current bearer token
  - rate limit /auth/login (per IP + email) and /auth/register(-root) (per IP)

One-time codes (pkg/otp):
  - one active code per identifier/purpose, a new code replaces the old one
  - max 5 wrong attempts per TTL window, not reset by issuing a new code;
constant-time comparison
  - reset-password: validate password confirmation before consuming the code, return the same error for unknown email and wrong code

Key-value counters:
    - redis: IncrementCounterWithLimit is atomic (MULTI: SET NX PX + INCR);
the WATCH-based version let concurrent callers pass the limit
  - in-memory: atomic counter creation/increment, Delete also removes counters

Withdrawals and secrets:
    - partial unique index on queued withdrawal_from_processing_wallets.request_id, a concurrent duplicate returns the existing duplicate request error
    - shared per-user rate limit (5/min) on seed phrase and private key export


Dependencies:
    - bump github.com/jackc/pgx/v5 to v5.13.0 (CVE-2026-33815, CVE-2026-33816, CVE-2026-41889)

Other:
  - CHANGELOG: drop links to the lost mtask.org domain
  - regenerate swagger docs

Deploy notes:
  - before migrating, check for duplicate queued withdrawals:
SELECT request_id, count(*) FROM withdrawal_from_processing_wallets
WHERE transfer_id IS NULL AND request_id IS NOT NULL
GROUP BY request_id HAVING count(*) > 1;
- OTP codes issued before the deploy become invalid (key format changed)