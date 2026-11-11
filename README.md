# Internal Helpdesk Ticketing System

Repositori ini berisi pengerjaan technical test Shift Engineer - Fullstack untuk PT Simple Journey Indonesia. Aplikasi dibangun dalam satu monorepo yang terdiri dari backend REST API menggunakan Go, antarmuka web menggunakan Vue 3, dan database PostgreSQL.

## Cara Menjalankan Aplikasi

### 1. Menggunakan Docker Compose
Pastikan Docker dan Docker Compose sudah berjalan di mesin lokal. Jalankan perintah berikut dari root direktori:

```bash
docker compose up --build
```

Saat container backend pertama kali menyala, sistem akan otomatis menjalankan migrasi tabel dan mengisi data awal (seed data).
- Frontend (Web UI): http://localhost atau http://localhost:5173 (diatur pada docker-compose.yml)
- Backend REST API: http://localhost:8080
- Spesifikasi OpenAPI: tersedia di file openapi.yaml pada root repositori.

### 2. Menjalankan Secara Lokal (Tanpa Docker)

Backend:
1. Buat database PostgreSQL baru dan salin .env.example menjadi .env, lalu sesuaikan koneksi database.
2. Masuk ke folder backend dan jalankan server:
```bash
cd backend
go mod download
go run ./cmd/api/main.go
```

Frontend:
1. Masuk ke folder frontend dan instal dependensi:
```bash
cd frontend
npm install
npm run dev
```

## Akun Seed (Data Awal)
Migrasi awal sudah menyertakan akun untuk masing-masing peran agar alur aplikasi bisa langsung diuji:

| Peran | Email | Password | Hak Akses Utama |
| --- | --- | --- | --- |
| Admin | admin@simplejourney.co.id | Password123! | Semua akses Agent + kelola pengguna, ubah role, dan lihat statistik |
| Agent | agent@simplejourney.co.id | Password123! | Lihat semua tiket, assign tiket, ubah status/prioritas, buat catatan internal |
| Requester | requester@simplejourney.co.id | Password123! | Buat tiket, lihat & komentar di tiket milik sendiri, batalkan/reopen tiket sendiri |

(Catatan: Jika terdapat akun requester kedua di seeder untuk pengujian isolasi akses/IDOR, gunakan kredensial yang tertera pada file seeder).

## Daftar Endpoint API
Dokumentasi lengkap skema request/response tersedia di openapi.yaml. Berikut ringkasan endpoint yang tersedia:

| Method | Path | Akses | Keterangan |
| --- | --- | --- | --- |
| POST | /api/auth/register | Publik | Registrasi akun baru (otomatis berperan sebagai requester) |
| POST | /api/auth/login | Publik | Login dan mendapatkan token JWT |
| GET | /api/me | Login | Mengambil profil pengguna yang sedang aktif |
| GET | /api/tickets | Login | Daftar tiket (mendukung pagination, filter status/prioritas/assignee, cari judul, sort) |
| POST | /api/tickets | Login | Membuat tiket bantuan baru |
| GET | /api/tickets/{id} | Pemilik, Agent, Admin | Detail tiket beserta daftar komentar dan lampiran |
| PATCH | /api/tickets/{id} | Sesuai aturan transisi | Mengubah status tiket, prioritas, atau agen yang menangani |
| POST | /api/tickets/{id}/comments | Pemilik, Agent, Admin | Menambahkan komentar biasa atau catatan internal (is_internal) |
| POST | /api/tickets/{id}/attachments | Pemilik, Agent, Admin | Mengunggah file lampiran pada tiket (maksimal 5MB) |
| GET | /api/tickets/{id}/attachments/{att_id} | Pemilik, Agent, Admin | Mengunduh file lampiran tiket |
| GET | /api/notifications/stream | Login | Koneksi Server-Sent Events (SSE) untuk notifikasi real-time |
| GET | /api/users | Admin | Melihat daftar seluruh pengguna |
| PATCH | /api/users/{id}/role | Admin | Mengubah peran pengguna (requester, agent, admin) |
| GET | /api/admin/stats | Admin | Ringkasan metrik tiket, beban kerja agen, dan kepatuhan SLA |

## Menjalankan Pengujian (Test)
Pengujian di sisi backend mencakup validasi matriks transisi status tiket, pembatasan akses antar peran (RBAC & pencegahan IDOR), pemisahan catatan internal agar tidak bocor ke requester, serta perhitungan SLA.

Untuk menjalankan seluruh test:
```bash
cd backend
go test -v ./...
```

## Catatan Arsitektur & Keputusan Teknis

### 1. Keputusan Arsitektur & Struktur Kode
Pada sisi backend, saya memisahkan kode ke dalam empat lapisan utama: domain (model dan aturan bisnis murni), repository (query database), service (orkestrasi logika bisnis dan kontrol akses), serta handler (transport HTTP dan validasi input). Pendekatan ini saya pilih agar aturan krusial seperti matriks perpindahan status tiket dapat diuji langsung lewat unit test tanpa harus menyalakan server HTTP maupun database.

Untuk database, saya menggunakan PostgreSQL dengan query SQL eksplisit (tanpa ORM berat) supaya query filter kombinasi, pencarian ILIKE, dan pagination pada daftar tiket lebih mudah dikontrol performanya. Kolom pengurutan (sort_by dan order) divalidasi menggunakan daftar putih (whitelist map) di level repository untuk menutup celah SQL Injection pada klausa ORDER BY.

Di sisi keamanan akses:
- **Validasi Kepemilikan (Anti-IDOR)**: Pengguna dengan peran requester dikunci di level query dan service agar hanya bisa membaca, mengomentari, dan mengunduh lampiran dari tiket yang requester_id-nya sesuai dengan ID di token JWT mereka.
- **Catatan Internal (is_internal)**: Saat requester mengambil detail tiket (GET /api/tickets/{id}), komentar dengan flag is_internal = true langsung difilter di sisi backend sebelum diserialisasi ke JSON, bukan sekadar disembunyikan di tampilan Vue.
- **Validasi Lampiran File**: Upload file dibatasi maksimal 5MB menggunakan http.MaxBytesReader dan divalidasi tipe kontennya menggunakan pembacaan magic bytes (http.DetectContentType), kemudian disimpan dengan nama UUID acak di luar folder publik untuk mencegah eksekusi script maupun path traversal.
- **Notifikasi Real-Time & SLA**: Untuk fitur notifikasi real-time, saya memilih Server-Sent Events (SSE) dibanding WebSocket karena arah komunikasinya satu arah (dari server ke klien saat ada pembaruan tiket/komentar), sehingga jauh lebih ringan dan otomatis berjalan di atas koneksi HTTP standar. Aturan SLA untuk tiket urgent dihitung berdasarkan selisih waktu sejak tiket dibuat (created_at) jika statusnya masih open melewati batas waktu yang ditentukan (default 2 jam).

### 2. Hal yang Sengaja Disederhanakan (Trade-off)
Karena batasan waktu pengerjaan take-home test, beberapa komponen saya buat sesederhana mungkin selama fungsionalitas utamanya terpenuhi:
- **Penyimpanan File Lokal**: File lampiran tiket masih disimpan di dalam local disk / Docker volume pada container backend, belum dipisahkan ke layanan object storage eksternal.
- **In-Memory SSE Hub & Rate Limiter**: Pengelolaan koneksi aktif SSE dan pembatasan request (rate limiting) pada endpoint autentikasi masih disimpan di memori aplikasi Go (sync.RWMutex). Ini sudah cukup untuk berjalan di satu container, namun state-nya tidak terbagi jika backend dijalankan lebih dari satu instance (horizontal scaling).
- **Manajemen Sesi JWT**: Autentikasi menggunakan satu stateless JWT dengan masa berlaku terbatas tanpa mekanisme tabel refresh token / pencabutan token (token revocation) di database.

### 3. Perbaikan Jika Ada Waktu Lebih
Jika proyek ini diteruskan ke lingkungan produksi, beberapa hal yang akan saya tambahkan adalah:
- Memindahkan penyimpanan lampiran ke S3-compatible storage (seperti AWS S3 atau MinIO) menggunakan mekanisme presigned URL agar beban I/O file tidak membebani server API.
- Mengintegrasikan Redis untuk kebutuhan Pub/Sub pada SSE (agar notifikasi tetap sinkron di banyak instance backend) serta untuk menyimpan counter rate limiter.
- Menambahkan background worker terjadwal untuk mengirimkan pengingat otomatis ke agen sebelum batas waktu SLA tiket urgent terlewati, bukan hanya menandai status terlambat saat data tiket diambil.
- Menambahkan riwayat audit (audit log) terpisah yang mencatat setiap perubahan status, prioritas, dan perpindahan assignee pada tiket.
