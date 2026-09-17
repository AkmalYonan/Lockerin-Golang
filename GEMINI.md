# GEMINI.md — Lockerin Backend Golang + Supabase

## Mission
Build Lockerin as a secure modular-monolith Go API backed by Supabase PostgreSQL. Flutter Mobile and Web Admin must consume the same Golang API.

## Paths
- Master: `E:\AKMAL\AKMAL MAINDATA\PROJECT\lockerin`
- Backend: `E:\AKMAL\AKMAL MAINDATA\PROJECT\lockerin\golang`

## Non-Negotiable Architecture
```text
Flutter ─────┐
             ├── HTTPS REST ──> Golang API ──> Supabase PostgreSQL
Web Admin ───┘                     │
                                  ├── Supabase Auth
                                  ├── Supabase Storage (optional)
                                  ├── Redis
                                  ├── Midtrans
                                  ├── FCM
                                  └── Head Locker
```

- Never let Flutter/Web Admin connect directly to PostgreSQL for business operations.
- Never expose Supabase service-role key to clients.
- Golang owns business rules, authorization, payment state, rental state, and IoT commands.
- Supabase is the persistent data platform.

## Supabase Usage
Recommended:
- Supabase PostgreSQL for application data.
- Supabase Auth for user identity.
- Supabase Storage for optional images.
- Supabase Realtime only as an optional event channel; do not use it to bypass backend authorization.

Environment:
- `SUPABASE_URL`
- `SUPABASE_ANON_KEY` only if needed server-side for public/auth flows.
- `SUPABASE_SERVICE_ROLE_KEY` server-only.
- `SUPABASE_DB_URL` or equivalent server database connection string.

Never commit these secrets.

## Database
Prefer migrations under:
```text
migrations/
  001_profiles.sql
  002_locations.sql
  003_lockers.sql
  004_rentals.sql
  005_payments.sql
  006_devices.sql
  ...
```

Use PostgreSQL constraints, foreign keys, indexes, and transactions.

For active slot protection, use a database constraint/index strategy that prevents multiple active rentals for the same slot.

## Suggested Backend Structure
```text
golang/
├─ cmd/api/
├─ internal/
│  ├─ config/
│  ├─ auth/
│  ├─ middleware/
│  ├─ domain/
│  ├─ handler/http/
│  ├─ service/
│  ├─ repository/
│  ├─ integration/
│  │  ├─ supabase/
│  │  ├─ midtrans/
│  │  ├─ fcm/
│  │  └─ locker_device/
│  └─ worker/
├─ migrations/
├─ docs/
├─ tests/
├─ .env.example
└─ README.md
```

## Auth
If Supabase Auth is used:
1. Client registers/logs in through the approved Golang API.
2. Golang interacts with/validates Supabase Auth.
3. Supabase user UUID becomes the stable identity.
4. `profiles` stores Lockerin-specific fields and role.
5. Backend enforces role authorization.

Do not implement a second password database when Supabase Auth is selected.

## API Rules
- `/api/v1`.
- Consistent response envelope.
- Request ID/correlation ID.
- Typed errors.
- Context propagation.
- Validation at boundaries.
- OpenAPI is the contract.

## Business State
Rental:
`draft -> reserved -> awaiting_payment -> paid -> active -> completed`
Failure:
`reserved -> expired`
`awaiting_payment -> cancelled`

Slot:
`available -> reserved -> occupied`
`occupied -> available`
Any operational state may become `maintenance/offline`.

Payment:
`pending -> settlement`
`pending -> deny/cancel/expire`
`settlement -> refund` when supported.

## IoT
Device auth is separate from user auth.
Every command needs:
- command ID
- locker ID
- slot ID
- action
- correlation ID
- timeout/expiry

ACK must be idempotent.

## Security
- Never log Supabase service key.
- Never log access/refresh tokens.
- Never log plaintext PIN.
- Never return internal secrets.
- Rate-limit auth and PIN.
- Validate payment webhook.
- Use RBAC.
- Audit critical mutations.

## Testing
Mandatory concurrency test:
two users attempt to reserve the same slot simultaneously; only one succeeds.

Test:
- auth
- reservation
- payment webhook idempotency
- PIN
- rental transitions
- device command/ACK
- RBAC

## Do Not
- Do not use Supabase directly from Flutter/Web for business data.
- Do not expose service-role key.
- Do not create a separate Node backend.
- Do not store passwords yourself if Supabase Auth is enabled.
- Do not trust client payment/rental status.
