# PRD — Lockerin Backend API (Golang + Supabase)

## 1. Ringkasan
LockerIn Backend adalah pusat API untuk Flutter Mobile dan Web Admin. Golang menjadi satu-satunya API bisnis yang diakses kedua client. Database utama menggunakan **Supabase PostgreSQL**.

Path:
`E:\AKMAL\AKMAL MAINDATA\PROJECT\lockerin\golang`

Arsitektur:
`Flutter / Web Admin -> Golang API -> Supabase PostgreSQL`
Tambahan integrasi: Redis, Midtrans, FCM, Head Locker/IoT.

## 2. Keputusan Arsitektur
- **Backend:** Go modular monolith.
- **Database:** Supabase PostgreSQL.
- **Auth:** Backend menyediakan endpoint auth. Untuk implementasi identity, rekomendasi menggunakan Supabase Auth sebagai identity provider; Golang memvalidasi token dan menerapkan RBAC/business rules. Jangan expose Supabase service-role key ke Flutter/Web.
- **Storage:** Supabase Storage bila membutuhkan upload gambar lokasi/promo/profile.
- **Realtime:** Supabase Realtime dapat digunakan sebagai pendukung sinkronisasi internal/admin, tetapi status bisnis tetap divalidasi melalui Golang.
- **Flutter dan Web Admin tidak boleh query Supabase langsung untuk data bisnis.**
- **Redis:** cache, TTL reservation, rate limit, distributed lock bila diperlukan.
- **Midtrans:** seluruh secret dan webhook diproses server-side.
- **IoT:** device tidak berbicara langsung dengan client.

## 3. Tujuan
- Satu API untuk Flutter dan Web Admin.
- Supabase menjadi persistent database.
- Menjamin konsistensi reservation/rental/payment.
- Menyediakan gateway aman ke Head Locker.
- Menyediakan histori transaksi lengkap dan audit.

## 4. Scope MVP

### Authentication
- Register.
- Login.
- Refresh session/token.
- Logout.
- Profile.
- Role: `user`, `admin`, `operator`, `viewer`.
- Jika Supabase Auth dipakai, `auth.users.id` menjadi identity ID dan tabel profile/domain menyimpan data aplikasi.

### Locations
- CRUD lokasi locker.
- Nama, alamat, latitude, longitude.
- Status lokasi.
- Operating hours.
- Tarif/default configuration.

### Lockers & Slots
- Locker/head locker.
- Device ID.
- Firmware/version.
- Heartbeat.
- Slot A1-E5 atau konfigurasi slot dinamis.
- Status: `available`, `reserved`, `occupied`, `maintenance`, `offline`, `unknown`.

### Rental
- Quote.
- Reserve slot dengan TTL.
- Payment.
- Start rental.
- Open locker.
- Active rental/countdown.
- Close locker.
- Expired rental.
- History.

### Security Code
- Generate one-time PIN.
- Hash PIN.
- TTL.
- Attempt limit.
- Lockout.
- Verify.
- Audit tanpa menyimpan plaintext PIN.

### Payment
- Create Midtrans transaction.
- Payment status.
- Webhook.
- Idempotency.
- Mapping payment -> rental -> user -> location -> locker -> slot.
- Refund/cancel bila diperlukan.

### Notifications
- FCM token.
- Promo.
- Payment status.
- Rental reminder.
- Rental expiry.

### Admin
- Dashboard.
- Users.
- Locations.
- Lockers.
- Slots.
- Pricing.
- Promos.
- Transactions.
- Devices.
- Audit logs.

## 5. Supabase Database Design

### `auth.users`
Dikelola Supabase Auth. Jangan membuat ulang password storage sendiri bila Supabase Auth digunakan.

### `profiles`
- `id` UUID PK/FK ke `auth.users.id`
- `name`
- `phone`
- `role`
- `status`
- timestamps

### `locations`
- id
- name
- address
- latitude
- longitude
- status
- operating_hours
- timestamps

### `lockers`
- id
- location_id
- code
- name
- device_id
- status
- firmware_version
- last_seen_at
- timestamps

### `locker_slots`
- id
- locker_id
- slot_code
- size
- status
- current_rental_id
- timestamps

### `rentals`
- id
- user_id
- location_id
- locker_id
- slot_id
- started_at
- expires_at
- ended_at
- status
- price_snapshot
- timestamps

### `security_codes`
- id
- rental_id
- pin_hash
- expires_at
- attempts
- max_attempts
- used_at
- timestamps

### `payments`
- id
- rental_id
- provider
- provider_transaction_id
- amount
- status
- payment_type
- paid_at
- raw_reference/metadata yang aman
- timestamps

### `device_commands`
- id
- locker_id
- slot_id
- command
- correlation_id
- status
- requested_at
- acknowledged_at
- timestamps

### `notifications`
- id
- user_id
- type
- title
- body
- read_at
- created_at

### `fcm_devices`
- id
- user_id
- token
- platform
- last_seen_at

### `promos`
- id
- title
- description
- code
- discount_type
- discount_value
- starts_at
- ends_at
- status

### `audit_logs`
- id
- actor_type
- actor_id
- action
- entity_type
- entity_id
- metadata
- created_at

## 6. Database Rules
- Gunakan UUID.
- Gunakan foreign key dan constraint di PostgreSQL.
- Gunakan transaction untuk state transition penting.
- Terapkan unique constraint agar satu slot tidak memiliki rental aktif ganda.
- Harga transaksi disimpan sebagai snapshot.
- Nominal rupiah menggunakan integer/bigint atau numeric sesuai kebutuhan, bukan floating point.
- Timestamp disimpan timezone-aware.
- RLS Supabase tetap dapat digunakan sebagai defense-in-depth, tetapi client tidak bypass Golang untuk business data.
- Service role key hanya berada di server.
- Migration database menjadi source-controlled artifact.

## 7. API Kontrak

### Auth
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/me`

### Locations
- `GET /api/v1/locations`
- `GET /api/v1/locations/{id}`
- `GET /api/v1/locations/{id}/availability`

### Lockers
- `GET /api/v1/lockers/{id}`
- `GET /api/v1/lockers/{id}/slots`

### Rental
- `POST /api/v1/rentals/quote`
- `POST /api/v1/rentals/reserve`
- `POST /api/v1/rentals/{id}/start`
- `POST /api/v1/rentals/{id}/open`
- `POST /api/v1/rentals/{id}/close`
- `GET /api/v1/rentals`
- `GET /api/v1/rentals/{id}`

### Security
- `POST /api/v1/rentals/{id}/security-code/verify`

### Payment
- `POST /api/v1/payments`
- `GET /api/v1/payments/{id}`
- `POST /api/v1/payments/webhook/midtrans`

### Notifications
- `POST /api/v1/devices/fcm-token`
- `GET /api/v1/notifications`
- `PATCH /api/v1/notifications/{id}/read`

### Head Locker
- `POST /api/v1/device/heartbeat`
- `POST /api/v1/device/status`
- `POST /api/v1/device/commands/{id}/ack`

### Admin
- `GET /api/v1/admin/dashboard`
- CRUD `/api/v1/admin/users`
- CRUD `/api/v1/admin/locations`
- CRUD `/api/v1/admin/lockers`
- CRUD `/api/v1/admin/slots`
- CRUD `/api/v1/admin/pricing`
- CRUD `/api/v1/admin/promos`
- `GET /api/v1/admin/transactions`
- `GET /api/v1/admin/devices`
- `GET /api/v1/admin/audit-logs`

## 8. Business Rules
1. Backend adalah source of truth.
2. Slot hanya memiliki satu rental aktif.
3. Reservation memiliki TTL.
4. Payment sukses ditentukan dari backend/webhook provider, bukan client.
5. Harga historis tidak berubah ketika pricing berubah.
6. Open command hanya boleh jika rental/payment valid.
7. PIN hanya disimpan dalam bentuk hash.
8. Device offline membuat slot tidak bookable.
9. Device command harus idempotent.
10. Semua aksi kritis masuk audit log.

## 9. Security
- Supabase service-role key hanya di backend.
- Validasi Supabase JWT/session di backend.
- RBAC server-side.
- Rate limit login/PIN.
- HTTPS.
- Input validation.
- Parameterized query.
- Secret melalui environment/secret manager.
- Jangan log token, PIN, password, server key, atau payment secret.

## 10. Acceptance Criteria
- Flutter dan Web Admin memakai API Golang yang sama.
- User tersimpan/teridentifikasi melalui Supabase Auth + profiles.
- Lokasi/locker/slot dapat dikelola.
- Availability dapat dibaca Flutter.
- Reservation aman terhadap race condition.
- Payment tersimpan lengkap.
- Head Locker dapat mengirim status.
- Backend dapat mengirim open command + menerima ACK.
- Transaction detail lengkap.
- Audit log tersedia.

## 11. Out of Scope MVP
- Microservices.
- Dynamic pricing kompleks.
- Multi-country.
- Loyalty.
- Offline-first penuh.

## 12. Definition of Done
- Supabase project/schema siap.
- Migration SQL tersimpan di repository.
- OpenAPI tersedia.
- `.env.example`.
- Docker/local setup.
- Unit + integration tests.
- CI build/test/lint.
- README deployment.
