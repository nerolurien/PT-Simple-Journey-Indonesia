# Simple Journey Helpdesk Monorepo

Sistem helpdesk tiket internal untuk operasional PT Simple Journey Indonesia. Repositori ini dibangun dalam arsitektur monorepo yang terdiri dari backend berbasis Go dan antarmuka web berbasis Vue 3.

## Kebutuhan Sistem

- Docker dan Docker Compose (metode deployment utama), ATAU
- Go 1.22+
- Node.js 20+
- PostgreSQL 16

## Cara Menjalankan Aplikasi

### 1. Menjalankan Menggunakan Docker Compose (Rekomendasi)

Cukup jalankan satu perintah berikut di direktori root proyek:

```bash
docker compose up --build
```

Perintah ini akan secara otomatis:
1. Menjalankan container PostgreSQL 16 dan menunggu status healthcheck siap.
2. Membangun dan menjalankan backend Go, mengeksekusi migrasi skema tabel secara idempoten, serta memverifikasi data seeder awal.
3. Membangun frontend Vue 3 dan menyajikannya via reverse proxy Nginx pada port 80 dan port 5173.

Akses aplikasi pada browser:
- Frontend: `http://localhost` atau `http://localhost:5173`
- Backend API langsung: `http://localhost:8080`
- Dokumentasi OpenAPI: file `openapi.yaml` pada root repositori

### 2. Menjalankan Secara Manual (Development Lokal)

Jika ingin menjalankan tanpa Docker:

1. Buat database PostgreSQL bernama `simple_journey_helpdesk`.
2. Salin `.env.example` menjadi `.env` di root atau atur environment variable:
   ```bash
   cp .env.example .env
   ```
3. Jalankan backend:
   ```bash
   cd backend
   go run ./cmd/api
   ```
   (Skema database dan seeder data awal akan otomatis dieksekusi saat backend pertama kali dijalankan).
4. Jalankan frontend:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   Buka `http://localhost:5173` pada browser.

## Akun Pengguna Awal (Seed Data)

Database secara otomatis diisi dengan akun pengguna awal untuk masing-masing peran operasional:

| Peran | Nama Lengkap | Email | Password | Hak Akses Utama |
| --- | --- | --- | --- | --- |
| Admin | Budi Santoso | admin@simplejourney.co.id | Password123! | Mengelola semua tiket, mengubah role pengguna, menutup tiket dari status open |
| Agent | Nadia Putri | agent@simplejourney.co.id | Password123! | Menangani tiket, mengubah status (in_progress, resolved), membuat catatan internal |
| Agent | Rizky Pratama | rizky.pratama@simplejourney.co.id | Password123! | Menangani tiket operasional |
| Requester | Siti Rahmawati | requester@simplejourney.co.id | Password123! | Membuat tiket, melihat tiket miliknya sendiri, membatalkan/menutup tiket miliknya |
| Requester | Dewi Lestari | dewi.lestari@simplejourney.co.id | Password123! | Akun pemohon kedua (untuk validasi IDOR) |

Catatan: Pada halaman login, tersedia tombol pintas untuk mengisi kredensial seed secara instan.

## Ringkasan Endpoint API

Spesifikasi lengkap tersedia dalam format OpenAPI 3.0 pada file `openapi.yaml`.

### Autentikasi & Pengguna
- `POST /api/auth/register` - Pendaftaran akun pemohon baru (role default: requester).
- `POST /api/auth/login` - Autentikasi dan penerbitan token JWT HS256.
- `GET /api/me` - Informasi profil pengguna yang sedang login.
- `GET /api/users` - Daftar seluruh pengguna (Khusus Admin).
- `PATCH /api/users/{id}/role` - Pembaruan peran pengguna (Khusus Admin, tidak dapat mengubah role akun sendiri).

### Tiket, Komentar & Lampiran
- `POST /api/tickets` - Pembuatan tiket bantuan baru (Status awal: open).
- `GET /api/tickets` - Daftar tiket dengan filter status, prioritas, pencarian judul (`q`), sorting whitelist, dan pagination. Otomatis dibatasi ke tiket milik sendiri jika requester.
- `GET /api/tickets/{id}` - Detail tiket, riwayat komentar, dan daftar lampiran berkas. Catatan internal otomatis disaring untuk requester.
- `PATCH /api/tickets/{id}` - Pembaruan status (sesuai matriks transisi), prioritas, atau penugasan teknisi.
- `POST /api/tickets/{id}/comments` - Penambahan komentar publik atau catatan internal (catatan internal hanya diizinkan untuk agent/admin).
- `POST /api/tickets/{id}/attachments` - Unggah berkas lampiran pada tiket (batas 5 MB, validasi tipe file ganda).
- `GET /api/tickets/{id}/attachments/{att_id}` - Unduh berkas lampiran dengan autentikasi dan proteksi IDOR.

### Notifikasi & Statistik (Fitur Tambahan)
- `GET /api/notifications/stream` - Stream notifikasi real-time via Server-Sent Events (SSE). Mendukung token query param atau Bearer header.
- `GET /api/admin/stats` - Agregasi metrik helpdesk, distribusi prioritas/status, SLA breach, dan leaderboard agen (Khusus Admin).

## Menjalankan Unit & Integration Test

Seluruh pengujian backend ditulis menggunakan package standar `testing` Go tanpa memerlukan database eksternal yang aktif:

```bash
cd backend
go test ./... -v
```

Pengujian mencakup:
1. Validasi alur transisi status yang sah untuk tiap peran.
2. Penolakan transisi status ilegal (422) dan pelanggaran hak peran (403).
3. Proteksi IDOR dan kontrol akses kepemilikan tiket.
4. Penapisan catatan internal agar tidak bocor ke role requester.
5. Proteksi pembaruan peran dan pencegahan admin mengubah role sendiri.
6. Pengecekan strict algoritma JWT dan token-bucket rate limiter.
7. Perhitungan SLA breach untuk tiket urgent yang belum ditangani vs tiket non-urgent.
8. Keamanan upload berkas: penolakan ekstensi terlarang (.exe, .sh, .html, .svg), penolakan MIME terselubung, penolakan file melebihi batas 5 MB, serta proteksi IDOR pada unduhan lampiran.

---

## Catatan Arsitektur & Keputusan Teknis

### 1. Struktur Kode Berlapis dan Pemilihan Database
Untuk backend, saya menyusun struktur kode ke dalam lapisan yang memiliki tanggung jawab tunggal: `cmd/` untuk entrypoint eksekusi, `internal/domain` untuk entitas inti dan logika bisnis murni, `internal/repository` untuk interaksi basis data terparameterisasi, `internal/service` untuk orkestrasi use-case dan aturan akses, serta `internal/handler` dan `internal/middleware` untuk penanganan HTTP.

Saya menggunakan PostgreSQL 16 bersama library `pgx/v5` dan `pgxpool`. PostgreSQL dipilih karena integritas relasionalnya yang ketat melalui foreign key constraints (`ON DELETE RESTRICT` dan `ON DELETE CASCADE`), index performa, serta konsistensi transaksi ACID yang penting untuk status tiket bantuan.

### 2. Implementasi Keamanan, Validasi Status & Fitur Tambahan
- **Matriks Transisi Status:** Logika transisi tiket dipusatkan pada fungsi `ValidateStatusTransition` di domain layer. Saya membedakan secara tegas antara transisi yang secara state machine memang tidak valid (dikembalikan sebagai HTTP 422 Unprocessable Entity) dan transisi yang secara alur ada namun pengguna tidak berhak melakukannya (dikembalikan sebagai HTTP 403 Forbidden).
- **Pencegahan IDOR (Insecure Direct Object References):** Pada level service, jika pengguna ber-role `requester`, query pencarian daftar tiket secara paksa menambahkan filter `requester_id = current_user_id`. Saat mengambil detail tiket, mengunggah lampiran, atau mengunduh berkas, kepemilikan tiket divalidasi terlebih dahulu sebelum operasi dijalankan.
- **Pembersihan Catatan Internal (OWASP A01):** Catatan internal (`is_internal = true`) tidak disembunyikan menggunakan manipulasi CSS di frontend, melainkan disaring langsung pada service layer Go sebelum data dienkode ke JSON. Dengan demikian, data internal tidak pernah keluar melalui jaringan ke klien pemohon.
- **Pencegahan SQL Injection pada Sorting:** Kolom pengurutan data tiket diperiksa terhadap whitelist map (`created_at`, `updated_at`, `priority`). Jika nilai input tidak cocok, sistem otomatis menggunakan default `t.created_at DESC` tanpa melakukan interpolasi string mentah ke dalam query SQL.
- **Keamanan Unggah Berkas (OWASP File Upload Security):** Untuk fitur bonus lampiran berkas, saya menerapkan validasi lapis ganda: whitelist ekstensi (`.png`, `.jpg`, `.jpeg`, `.pdf`, `.txt`, `.log`) dikombinasikan dengan magic-byte sniffing melalui `http.DetectContentType` pada 512 byte pertama (menolak executable dan script terselubung seperti file HTML yang disamarkan). Batas ukuran 5 MB ditegakkan via `http.MaxBytesReader`. Untuk mencegah path traversal, file disimpan dengan nama acak UUID pada direktori non-publik (`./storage/attachments`) dengan permission ketat, dan disajikan dengan header `Content-Disposition: attachment` serta `X-Content-Type-Options: nosniff`.
- **Aturan SLA Tiket Urgent:** Tiket berprioritas `urgent` yang masih berstatus `open` melebihi ambang batas (`SLA_URGENT_HOURS`, default 2 jam) otomatis ditandai sebagai `is_sla_breached = true` dan memunculkan badge indikator merah "Terlambat (SLA)" pada antarmuka. Saat tiket berpindah ke `in_progress`, timestamp `first_responded_at` dicatat sebagai bukti audit waktu respon awal.
- **Notifikasi Real-Time via Server-Sent Events (SSE):** Hub in-memory berbasis channels dan `sync.RWMutex` menyiarkan event tiket secara aman ke pengguna yang sedang online. Agen & Admin menerima alert tiket baru, sedangkan pemohon menerima update status dan komentar balasan secara langsung. Catatan internal secara ketat tidak memicu notifikasi ke akun pemohon.
- **Halaman Statistik Admin:** Dashboard analitik metrik tiket, visualisasi distribusi status & prioritas murni CSS, penghitungan tiket terlambat SLA, durasi rata-rata penyelesaian penanganan, dan leaderboard agen teratas.

### 3. Kompromi Teknis Berdasarkan Batasan Waktu
Dalam batas waktu pengerjaan tugas teknis ini, saya mengambil beberapa keputusan penyederhanaan:
- **Rate Limiter & SSE Hub In-Memory:** Rate limiter token bucket dan SSE Hub diimplementasikan secara in-memory per alamat instance dengan goroutine pembersih data usang berkala. Pendekatan ini menghindari ketergantungan Redis untuk deployment satu instance, namun siap dipindahkan ke Redis Pub/Sub jika aplikasi di-deploy secara multi-node.
- **Penyimpanan Berkas Lokal Terisolasi:** Berkas lampiran disimpan pada volume direktori lokal terisolasi dalam container Docker. Pada skala enterprise, ini dapat dialihkan ke layanan Object Storage S3 yang kompatibel.

### 4. Rencana Pengembangan Lanjutan
Apabila diberikan waktu pengembangan lebih lanjut, perbaikan yang akan saya prioritaskan adalah:
1. **SLA Escalation Worker:** Membuat background worker terjadwal untuk mengirimkan notifikasi eskalasi otomatis ke email manager jika tiket berstatus urgent belum ditangani menjelang batas SLA.
2. **Object Storage Integration:** Integrasi penyimpanan berkas ke AWS S3 atau MinIO menggunakan mekanisme presigned URL.
3. **Multi-Instance SSE Pub/Sub:** Menghubungkan SSE Hub dengan Redis Pub/Sub untuk sinkronisasi notifikasi real-time lintas replika container horizontal.
