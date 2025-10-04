# Steganography on Audio Files with Multiple LSB Method

Aplikasi steganografi untuk menyisipkan dan mengekstrak berkas rahasia ke/dari file MP3 menggunakan metode Multiple-LSB, dilengkapi opsi enkripsi Vigenere, pemakaian key untuk penentuan posisi bit, perhitungan kapasitas, dan evaluasi kualitas dengan PSNR. Proyek ini dibuat sebagai Minor Assignment II (Tucil II) IF4020 Kriptografi Semester I 2025/2026.

Frontend statis tersedia pada root server dan berinteraksi dengan backend REST sederhana berbasis Go.

## Daftar Isi

- [Nama & Deskripsi Program](#nama--deskripsi-program)
- [Tech Stack](#tech-stack)
- [Dependensi](#dependensi)
- [Cara Menjalankan](#cara-menjalankan)
- [Fitur Utama](#fitur-utama)
- [API Endpoints](#api-endpoints)
- [Struktur Proyek](#struktur-proyek)
- [Contoh Pemakaian Singkat](#contoh-pemakaian-singkat)
- [Daftar Anggota](#daftar-anggota)
- [Lisensi](#lisensi)

## Nama & Deskripsi Program

Steganography on Audio Files with Multiple LSB Method — alat untuk menyembunyikan berkas (PDF, TXT, gambar, dsb.) ke dalam MP3 melalui teknik manipulasi bit paling rendah (LSB) dan membaca kembali pesan tersembunyi tersebut. Opsi tambahan meliputi:

- Penyisipan metadata (nama asli file, tipe, ukuran, konfigurasi LSB, dll.)
- Enkripsi Vigenere atas payload sebelum disisipkan
- Pemakaian kunci untuk menentukan posisi bit agar lebih teracak
- Perhitungan kapasitas maksimum yang tersedia
- Perhitungan PSNR untuk menilai kualitas hasil penyisipan

## Tech Stack

- Backend: Go 1.21 (net/http, encoding/json, dll. — standar library)
- Frontend: HTML, CSS, JavaScript (vanilla) yang disajikan statis dari folder `static/`

## Dependensi

- Go 1.21 atau lebih baru
- Tidak ada dependensi pihak ketiga (hanya Go standard library)

## Cara Menjalankan

1) Pastikan Go sudah terpasang (>= 1.21).
2) Jalankan perintah build dan eksekusi berikut dari root proyek:

```bash
go build main.go
./main
```

3) Akses antarmuka web pada:

- Frontend: http://localhost:8080
- API Health: http://localhost:8080/api/health

Server secara default berjalan di port 8080 dan akan membuat folder sementara `./temp` untuk pemrosesan file.

## Fitur Utama

- Penyisipan (embed) berkas rahasia ke MP3 via LSB (1–4 bit)
- Ekstraksi (extract) berkas rahasia beserta metadata
- Opsi enkripsi Vigenere dan pemakaian key untuk penentuan posisi bit
- Perhitungan kapasitas penyisipan (byte dan format human-readable)
- Perhitungan PSNR untuk membandingkan file MP3 asli vs hasil embed
- Frontend sederhana untuk unggah file dan uji cepat

## API Endpoints

Base URL: `http://localhost:8080`

- GET `/api/health` — Cek status server
- POST `/api/embed` — Sisipkan berkas ke MP3
	- Form fields: `mp3_file` (file), `secret_file` (file), `key` (string), `use_encryption` ("true"/"false"), `use_key_for_position` ("true"/"false"), `method` ("lsb"/"header", default `lsb`), `lsb_bits` (1–4, default 1)
- POST `/api/extract` — Ekstrak berkas dari MP3
	- Form fields: `mp3_file` (file), `key` (string, opsional — wajib bila saat embed memakai enkripsi)
- POST `/api/capacity` — Hitung kapasitas embed
	- Form fields: `mp3_file` (file), `method` ("lsb"/"header"), `lsb_bits` (1–4 untuk `lsb`)
- POST `/api/psnr` — Hitung PSNR antara MP3 asli dan hasil
	- Form fields: `original_file` (file), `modified_file` (file)

Header hasil ekstraksi (bila tersedia metadata):

- `X-Original-Filename`, `X-File-Type`, `X-Secret-Size`, `X-Used-Encryption`, `X-Used-Key-Position`, `X-LSB-Bits`

## Struktur Proyek

```
.
├── main.go
├── go.mod
├── internal/
│   ├── crypto/           # Enkripsi Vigenere
│   ├── handlers/         # HTTP handlers (embed, extract, capacity, psnr, health)
│   ├── middleware/       # CORS
│   ├── models/           # Tipe request/response (jika diperlukan)
│   └── stego/            # Logika LSB, header stego, metadata
├── static/               # Frontend statis (HTML, JS)
└── test/                 # Berkas uji contoh (mp3 & payload)
```

## Contoh Pemakaian Singkat

- Uji cepat melalui frontend: buka `http://localhost:8080`, unggah file MP3 dan berkas rahasia, atur `LSB bits`, `key`, dan opsi enkripsi sesuai kebutuhan, lalu klik Embed/Extract.
- Direktori `test/` menyediakan contoh MP3 dan beberapa payload untuk pengujian.

## Daftar Anggota

| NIM        | Nama                 |
|------------|----------------------|
| 13522150   | Nama Anggota 1       |
| 13522158   | Nama Anggota 2       |



## Lisensi

Proyek ini dirilis di bawah lisensi MIT — lihat berkas `LICENSE`.
