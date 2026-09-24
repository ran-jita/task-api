# Task Management API

REST API untuk manajemen task multi-user, dibangun dengan Go + Echo + PostgreSQL.

## Tech Stack

- **Language**: Go
- **Framework**: Echo v4
- **Database**: PostgreSQL
- **Auth**: JWT (golang-jwt/jwt)
- **Migration**: golang-migrate

## Struktur Project

Mengikuti prinsip Clean Architecture — dependency mengalir satu arah (handler → usecase → repository interface), sehingga business logic (`usecase`) tidak bergantung pada detail database.

task-api/
├── cmd/api/ # entrypoint aplikasi
├── internal/
│ ├── entity/ # domain model (Task, User, TaskLog)
│ ├── usecase/ # business logic
│ ├── repository/ # interface akses data
│ │ └── postgres/ # implementasi konkret (Postgres)
│ ├── handler/http/ # HTTP handler + routing
│ └── middleware/ # JWT auth, request ID, logging, panic recovery
├── migrations/ # skema database (golang-migrate)


## Setup & Menjalankan

### Prasyarat
- Go 1.21+
- PostgreSQL 14+
- [golang-migrate CLI](https://github.com/golang-migrate/migrate)

### Langkah

```bash
# 1. Clone & masuk folder project
git clone https://github.com/ran-jita/task-api.git
cd task-api

# 2. Install dependency
go mod tidy

# 3. Buat database
createdb task_api

# 4. Copy environment variable
cp .env.example .env
# lalu sesuaikan DATABASE_URL dan JWT_SECRET di .env

# 5. Jalankan migration
migrate -path migrations -database "postgres://user:password@localhost:5432/task_api?sslmode=disable" up

# 6. Jalankan server
go run cmd/api/main.go
```

Server berjalan di `http://localhost:8080` (atau sesuai `PORT` di `.env`).

### Menjalankan Unit Test

```bash
go test ./... -race -v
```

Flag `-race` mengaktifkan Go race detector, terutama relevan untuk test race condition pada idempotency key (lihat `internal/usecase/idempotency_usecase_test.go`).

## API Endpoints

### Auth (public)
| Method | Endpoint | Deskripsi |
|---|---|---|
| POST | `/auth/register` | Registrasi user baru |
| POST | `/auth/login` | Login, mengembalikan JWT token |

### Tasks (butuh `Authorization: Bearer <token>`)
| Method | Endpoint | Deskripsi |
|---|---|---|
| POST | `/tasks` | Buat task baru (wajib header `Idempotency-Key`) |
| GET | `/tasks` | List task milik user, mendukung `?status=`, `?title=`, `?page=`, `?limit=` |
| GET | `/tasks/:id` | Detail satu task |
| PUT | `/tasks/:id` | Update task |
| DELETE | `/tasks/:id` | Hapus task |
| POST | `/tasks/:id/assign` | Assign task ke user lain (body: `{"assignee_id": "..."}`) |

## Keputusan Desain & Trade-off

Beberapa keputusan yang sengaja diambil, beserta alasannya:

- **Idempotency key disimpan di PostgreSQL, bukan in-memory/Redis** — supaya window 24 jam tetap valid walau server restart, dan tidak menambah dependency infra baru untuk skala kebutuhan ini.
- **Race-condition safety pada idempotency key mengandalkan unique constraint di database** (`Reserve` sebelum eksekusi), bukan locking manual di level aplikasi — pola *reserve-then-execute* memastikan operasi hanya dijalankan oleh request yang "menang" klaim key.
- **Assign task adalah satu-satunya jalur resmi untuk mengubah `assignee_id`** — endpoint update task biasa (`PUT /tasks/:id`) tidak bisa mengubah assignee, untuk menjaga audit trail (`task_logs`) tetap konsisten.
- **Delete bersifat hard delete**, bukan soft delete — karena soft delete akan menambah kompleksitas di setiap query GET (filter `deleted_at`).
- **Notifikasi assign dikirim setelah `Commit()` transaksi berhasil**, bukan sebelum — untuk menghindari kondisi user menerima notifikasi palsu jika transaksi ternyata gagal di tahap commit.

## Belum Diimplementasikan / Potensi Pengembangan

- Refresh token (saat ini JWT berlaku 24 jam tanpa mekanisme refresh)
- Rate limiting per user/IP