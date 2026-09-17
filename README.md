# 🔐 Lockerin Backend API (Golang + Supabase)

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/Database-Supabase%20PostgreSQL-3ECF8E?style=for-the-badge&logo=supabase)](https://supabase.com)
[![Payment](https://img.shields.io/badge/Payment-Midtrans%20Snap-002B49?style=for-the-badge)](https://midtrans.com)
[![Documentation](https://img.shields.io/badge/API%20Docs-Swagger%20OpenAPI%203.0-85EA2D?style=for-the-badge&logo=swagger)](http://localhost:8080/swagger)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)

Modular-Monolith REST API berperforma tinggi untuk ekosistem **Lockerin Smart Locker**. Backend ini dirancang menggunakan bahasa Go (Golang) dan Supabase PostgreSQL sebagai pusat pengendali bisnis (*single source of truth*) yang menghubungkan aplikasi klien Flutter, Web Admin Dashboard, payment gateway, dan perangkat IoT Head Locker.

---

## 📑 Daftar Isi
- [Arsitektur Sistem & Ekosistem](#-arsitektur-sistem--ekosistem)
- [Integrasi Ekosistem (Lockerin-Web & Lockerin-Mobile)](#-integrasi-ekosistem)
- [Fitur Utama](#-fitur-utama)
- [Daftar Dependensi & Teknologi](#-daftar-dependensi--teknologi)
- [Struktur Direktori](#-struktur-direktori)
- [Panduan Instalasi & Setup](#-panduan-instalasi--setup)
  - [1. Prasyarat](#1-prasyarat)
  - [2. Clone Repository](#2-clone-repository)
  - [3. Konfigurasi Environment (.env)](#3-konfigurasi-environment-env)
  - [4. Setup Database Supabase & Migrasi](#4-setup-database-supabase--migrasi)
  - [5. Menjalankan Server](#5-menjalankan-server)
  - [6. Menjalankan dengan Docker](#6-menjalankan-dengan-docker-opsional)
- [Pengujian (Testing & Concurrency)](#-pengujian-testing)
- [Dokumentasi API & Swagger UI](#-dokumentasi-api--swagger-ui)
- [Keamanan & Best Practices](#-keamanan--best-practices)

---

## 🏛️ Arsitektur Sistem & Ekosistem

```text
┌──────────────────────────────────────┐       ┌──────────────────────────────────────┐
│        Lockerin-Mobile               │       │          Lockerin-Web                │
│    (Flutter iOS & Android)           │       │    (Admin & Operator Dashboard)      │
└──────────────────┬───────────────────┘       └──────────────────┬───────────────────┘
                   │                                              │
                   │ HTTPS REST (/api/v1/*)                       │ HTTPS REST (/api/v1/admin/*)
                   │ Bearer JWT Token                             │ Bearer JWT Token (Admin Role)
                   ▼                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                       LOCKERIN GOLANG BACKEND API SERVER                            │
│ ─────────────────────────────────────────────────────────────────────────────────── │
│  • Routing & Middleware: chi router, CORS, RequestID, Rate Limiting, RBAC           │
│  • Domain Logic: Rental State Machine, Dynamic Salted PIN Engine, Worker Expiry     │
│  • Concurrency Protection: Partial PostgreSQL Unique Index on active slot rentals   │
│  • Swagger OpenAPI 3.0 Interactive Documentation Engine                             │
└──────────┬───────────────────┬───────────────────┬───────────────────┬──────────────┘
           │                   │                   │                   │
           ▼                   ▼                   ▼                   ▼
┌────────────────────┐ ┌───────────────┐ ┌────────────────┐ ┌─────────────────────────┐
│ Supabase Cloud     │ │ Midtrans      │ │ Firebase Cloud │ │ IoT Head Locker Gateway │
│ PostgreSQL & Auth  │ │ Payment Snap  │ │ Messaging(FCM) │ │ Solenoid Hardware & ACK │
└────────────────────┘ └───────────────┘ └────────────────┘ └─────────────────────────┘
```

---

## 🔗 Integrasi Ekosistem

Backend ini berperan sebagai penghubung sentral (*orchestrator*) bagi platform Lockerin:

### 1. 🌐 Terhubung dengan **Lockerin-Web** (Admin Dashboard)
- **Role:** Menyediakan endpoint administrasi lengkap untuk manajemen lokasi stasiun loker, monitoring status slot (*available / reserved / occupied / maintenance*), rekap transaksi keuangan, pembuatan voucher promo diskon, dan audit logs aktivitas sistem.
- **Base Endpoint:** `/api/v1/admin/*`
- **Autentikasi:** Bearer JWT dengan validasi hak akses RBAC (`admin` / `operator`).
- **CORS Support:** CORS middleware telah diaktifkan untuk mengizinkan integrasi dari domain frontend web (misal: `http://localhost:3000`, `http://localhost:5173`, atau domain produksi).

### 2. 📱 Terhubung dengan **Lockerin-Mobile** (Flutter Client)
- **Role:** Melayani pengguna akhir untuk autentikasi profil, pencarian lokasi loker terdekat, reservasi slot, proses checkout & pembuatan *Snap Payment URL* Midtrans, pembukaan pintu loker menggunakan kode PIN / TOTP dinamis atau trigger remote Bluetooth/Internet, serta melihat riwayat sewa.
- **Base Endpoint:** `/api/v1/rentals`, `/api/v1/lockers`, `/api/v1/payments`, `/api/v1/auth`

### 3. 🤖 Terhubung dengan **IoT Head Locker Hardware**
- **Role:** Menerima heartbeat telemetry, sinkronisasi status pintu solenoid hardware, eksekusi perintah buka loker (*remote command*), serta validasi verifikasi ACK (Acknowledge) yang idempoten.
- **Base Endpoint:** `/api/v1/devices/*`

---

## ✨ Fitur Utama

- 🛡️ **Role-Based Access Control (RBAC):** Pemisahan otorisasi ketat antara `user`, `operator`, dan `admin`.
- ⚡ **Anti Double-Booking (Concurrency Protection):** Proteksi tingkat basis data dengan PostgreSQL partial unique index untuk menjamin dua pengguna tidak dapat memesan slot loker yang sama secara bersamaan.
- 💳 **Integrasi Midtrans Payment Gateway:** Mendukung pembuatan Snap Payment Token dan webhook otomatis yang idempoten untuk verifikasi transaksi berhasil, kedaluwarsa, atau dibatalkan.
- 🔑 **Dynamic Cryptographic PIN Engine:** Menghasilkan PIN 6-digit dengan SHA-256 + salt unik, dilengkapi rate limiting dan penguncian otomatis (*lockout*) setelah percobaan gagal berulang.
- ⏱️ **Background Worker Cleanup:** Routine otomatis untuk membatalkan reservasi yang melewati batas waktu (*TTL expired*) dan mengembalikan status slot menjadi tersedia (*available*).
- 📜 **Swagger UI OpenAPI 3.0 Terintegrasi:** Dokumentasi API interaktif yang bisa langsung diuji coba melalui browser tanpa instalasi tool tambahan.
- 🔄 **Dual Data Layer (PostgreSQL & In-Memory Fallback):** Mendukung koneksi live Supabase PostgreSQL serta In-Memory thread-safe repository untuk kemudahan demo dan pengujian lokal instan.

---

## 📦 Daftar Dependensi & Teknologi

Backend dibangun menggunakan ekosistem Go modern:

| Dependensi / Library | Versi | Deskripsi & Kegunaan |
| :--- | :--- | :--- |
| **[Go (Golang)](https://golang.org)** | `1.24+` | Bahasa pemrograman utama berkinerja tinggi |
| **[github.com/go-chi/chi/v5](https://github.com/go-chi/chi)** | `v5.3.2` | HTTP router modular yang ringan, cepat, dan idiomatik |
| **[github.com/go-chi/cors](https://github.com/go-chi/cors)** | `v1.2.2` | Middleware CORS untuk komunikasi dengan Web Frontend |
| **[github.com/jackc/pgx/v5](https://github.com/jackc/pgx)** | `v5.11.0` | Driver PostgreSQL & Connection Pooler performa tinggi |
| **[github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)** | `v5.3.1` | Pembuatan dan verifikasi token JSON Web Token (JWT) |
| **[github.com/google/uuid](https://github.com/google/uuid)** | `v1.6.0` | Generator UUID v4 untuk identitas entitas sistem |
| **[github.com/joho/godotenv](https://github.com/joho/godotenv)** | `v1.5.1` | Pemuat konfigurasi variabel lingkungan dari file `.env` |
| **[golang.org/x/crypto](https://golang.org/x/crypto)** | `v0.57.0` | Hashing kriptografi password & enkripsi PIN keamanan |

---

## 📁 Struktur Direktori

```text
golang/
├─ cmd/
│  └─ api/
│     └─ main.go                 # Entrypoint server & dependency injection
├─ internal/
│  ├─ config/                    # Loader konfigurasi environment (.env)
│  ├─ domain/                    # Entitas bisnis, DTOs, dan error response
│  ├─ auth/                      # JWT Manager & Token Verifier
│  ├─ middleware/                # CORS, RequestID, Logger, RateLimiter, RBAC, Recovery
│  ├─ repository/                # Data Access Layer
│  │  ├─ postgres/               # Supabase PostgreSQL repository (pgx/v5)
│  │  └─ inmemory/               # In-Memory thread-safe repository (Demo/Testing)
│  ├─ service/                   # Core Business Logic (Rental, Payment, PIN, Device)
│  ├─ handler/
│  │  └─ http/                   # HTTP REST Controllers (/api/v1/*)
│  ├─ integration/               # Eksternal Adapter (Supabase, Midtrans, FCM, IoT)
│  └─ worker/                    # Background Worker (Auto-expire overdue reservations)
├─ migrations/                   # Skema SQL Supabase PostgreSQL (001 s/d 008)
├─ docs/                         # OpenAPI 3.0 Contract & Swagger Embed
│  ├─ docs.go                    # Go embed untuk openapi.yaml
│  └─ openapi.yaml               # Spesifikasi lengkap OpenAPI 3.0
├─ tests/                        # Concurrency, Unit & Integration Tests
├─ .env.example                  # Template konfigurasi environment
├─ .gitignore                    # Git ignore rule
├─ Dockerfile                    # Container build multi-stage
├─ docker-compose.yml            # Docker Compose orchestration
└─ README.md                     # Dokumentasi proyek
```

---

## 🚀 Panduan Instalasi & Setup

Ikuti langkah-langkah di bawah ini untuk menginstal dan menjalankan proyek di lingkungan lokal:

### 1. Prasyarat
Pastikan software berikut telah terpasang di komputer Anda:
- **Go**: versi 1.20 atau lebih baru (disarankan Go 1.24+) [Download Go](https://go.dev/dl/)
- **Git**: [Download Git](https://git-scm.com/)
- Akun **Supabase** (Gratis di [supabase.com](https://supabase.com))
- *(Opsional)* Akun **Midtrans Sandbox** untuk testing pembayaran ([midtrans.com](https://midtrans.com))

---

### 2. Clone Repository
```bash
git clone https://github.com/AkmalYonan/Lockerin-Golang.git
cd Lockerin-Golang
```

---

### 3. Konfigurasi Environment (`.env`)
Salin file template `.env.example` menjadi `.env`:
```bash
cp .env.example .env
```

Buka file `.env` dan sesuaikan nilainya:
```env
# Server Configuration
APP_ENV=development
PORT=8080

# Supabase PostgreSQL Configuration
# Dapatkan connection string dari: Supabase Dashboard -> Project Settings -> Database -> URI (Session / Transaction Pooler)
SUPABASE_DB_URL=postgresql://postgres.your-project-ref:your-password@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres?sslmode=require
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_ANON_KEY=your-supabase-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-supabase-service-role-key

# JWT Secret untuk Autentikasi
JWT_SECRET=super-secret-lockerin-jwt-key-2026-production

# Midtrans Payment Gateway
MIDTRANS_SERVER_KEY=SB-Mid-server-your-midtrans-server-key
MIDTRANS_CLIENT_KEY=SB-Mid-client-your-midtrans-client-key
MIDTRANS_IS_PRODUCTION=false

# IoT Head Locker Hardware Security Key
IOT_DEVICE_SECRET_KEY=lockerin-head-locker-secret-2026

# Business Rules
RESERVATION_TTL_MINUTES=15
PIN_MAX_ATTEMPTS=5
PIN_EXPIRY_MINUTES=1440
```

> 💡 **Info Demo Mode:** Jika `SUPABASE_DB_URL` dikosongkan, server akan otomatis berjalan dalam **In-Memory Mode** dengan data dummy bawaan yang langsung siap digunakan tanpa database eksternal.

---

### 4. Setup Database Supabase & Migrasi

Jalankan skrip migrasi SQL secara berurutan di **SQL Editor** pada dashboard Supabase Cloud Anda:

1. `migrations/001_profiles.sql` — Tabel pengguna & role RBAC.
2. `migrations/002_locations.sql` — Data stasiun dan lokasi loker.
3. `migrations/003_lockers.sql` — Data loker dan slot.
4. `migrations/004_rentals.sql` — Tabel transaksi sewa & index proteksi double-booking.
5. `migrations/005_security_codes.sql` — Tabel kode PIN dinamis & tracking percobaan.
6. `migrations/006_payments.sql` — Tabel pembayaran & histori log webhook Midtrans.
7. `migrations/007_devices.sql` — Pendaftaran perangkat IoT Head Locker & logs audit perintah.
8. `migrations/008_notifications_promos_audit.sql` — Notifikasi, voucher diskon promo, dan audit trail.

---

### 5. Menjalankan Server

Unduh dependensi Go dan jalankan aplikasi:
```bash
# Download dependensi
go mod download

# Jalankan server API
go run ./cmd/api/main.go
```

Output terminal saat server berhasil aktif:
```text
==================================================
   LOCKERIN SMART LOCKER BACKEND SERVICE v1.0.0   
==================================================
[Database] Connected to Supabase PostgreSQL: pool ready
[Worker] Auto-expire reservation background worker started
[Docs] Swagger UI ready at: http://localhost:8080/swagger
[Server] Listening on http://localhost:8080
==================================================
```

---

### 6. Menjalankan dengan Docker (Opsional)

Jika ingin menjalankan aplikasi di dalam container Docker:
```bash
# Build dan jalankan dengan Docker Compose
docker compose up -d --build

# Melihat log
docker compose logs -f
```

---

## 🧪 Pengujian (Testing)

Proyek ini telah dilengkapi dengan unit test, flow test pembayaran, serta **Mandatory Concurrency Test** untuk menguji ketahanan sistem terhadap race condition ketika multiple user memesan slot loker yang sama:

```bash
# Menjalankan seluruh test suite
go test -v ./tests/...
```

Hasil pengujian yang diharapkan:
```text
=== RUN   TestAuthFlow
--- PASS: TestAuthFlow (0.00s)
=== RUN   TestPINSecurityAndLockout
--- PASS: TestPINSecurityAndLockout (0.00s)
=== RUN   TestPaymentWebhookIdempotencyAndFlow
--- PASS: TestPaymentWebhookIdempotencyAndFlow (0.21s)
=== RUN   TestMandatorySlotConcurrency
    concurrency_test.go:92: Concurrency Result: 1 Succeeded, 9 Failed out of 10 concurrent requests
--- PASS: TestMandatorySlotConcurrency (0.00s)
PASS
ok      lockerin-backend/tests  0.25s
```

---

## 📖 Dokumentasi API & Swagger UI

Dokumentasi API interaktif Swagger UI telah tertanam langsung di dalam binary Go:

- 🌐 **Swagger UI Interaktif:** [http://localhost:8080/swagger](http://localhost:8080/swagger) atau [http://localhost:8080/docs](http://localhost:8080/docs)
- 📄 **Spesifikasi OpenAPI 3.0 (YAML):** [http://localhost:8080/docs/openapi.yaml](http://localhost:8080/docs/openapi.yaml)
- 📝 **File Kontrak:** [docs/openapi.yaml](docs/openapi.yaml)

### Ringkasan Endpoint Utama:

| Modul | Method | Endpoint | Keterangan |
| :--- | :--- | :--- | :--- |
| **Auth** | `POST` | `/api/v1/auth/register` | Pendaftaran pengguna baru |
| **Auth** | `POST` | `/api/v1/auth/login` | Login dan perolehan JWT Token |
| **Auth** | `GET` | `/api/v1/auth/me` | Informasi profil pengguna saat ini |
| **Locations** | `GET` | `/api/v1/locations` | Daftar stasiun loker beserta slot tersedia |
| **Rentals** | `POST` | `/api/v1/rentals` | Reservasi slot loker (Concurrency Protected) |
| **Rentals** | `GET` | `/api/v1/rentals/{id}` | Detail informasi status sewa |
| **Payments** | `POST` | `/api/v1/payments/checkout` | Buat transaksi Midtrans Snap Token & URL |
| **Payments** | `POST` | `/api/v1/payments/webhook` | Webhook notifikasi status dari Midtrans |
| **PIN** | `GET` | `/api/v1/rentals/{id}/pin` | Ambil kode PIN akses loker (Terverifikasi) |
| **PIN** | `POST` | `/api/v1/rentals/{id}/unlock-pin` | Verifikasi PIN untuk buka pintu loker |
| **Devices** | `POST` | `/api/v1/devices/heartbeat` | Heartbeat & status sinkronisasi IoT |
| **Devices** | `POST` | `/api/v1/devices/ack` | Idempotent acknowledgment eksekusi solenoid |
| **Admin** | `GET` | `/api/v1/admin/dashboard` | Statistik performa, omzet, & okupansi loker |
| **Admin** | `POST` | `/api/v1/admin/lockers` | Tambah dan konfigurasi unit loker baru |

---

## 🔒 Keamanan & Best Practices

1. **Prinsip Zero-Trust Client:** Status pembayaran dan pembukaan loker tidak pernah dipercayakan dari input klien, melainkan divalidasi langsung melalui Webhook Midtrans dan verifikasi state backend.
2. **Kerahasiaan Kredensial:** Supabase Service-Role Key dan JWT Secret tidak pernah diekspos ke klien mobile/web.
3. **Idempotensi Webhook:** Webhook Midtrans dapat dipanggil berulang kali tanpa risiko duplikasi settlement atau pembuatan ganda kode PIN.
4. **Audit Trail:** Seluruh aksi kritis administratif dan eksekusi pembukaan hardware tercatat pada tabel `audit_logs` dan `device_commands`.

---

## 👥 Kontributor & Lisensi

- **Repository:** [Lockerin-Golang (GitHub)](https://github.com/AkmalYonan/Lockerin-Golang.git)
- **Frontend Admin:** [Lockerin-Web](https://github.com/AkmalYonan)
- **Mobile Client:** [Lockerin-Mobile (Flutter)](https://github.com/AkmalYonan)
- **Lisensi:** Didistribusikan di bawah Lisensi MIT.
