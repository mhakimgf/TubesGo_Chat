# Go Chat Application

Aplikasi chat berbasis jaringan TCP menggunakan Go. Mendukung multi-client, room system, dan broadcast pesan.

## Fitur

### Fitur Wajib
- ✅ Multi-client: server menerima banyak koneksi secara bersamaan
- ✅ Broadcast: pesan dikirim ke semua client lain yang terhubung
- ✅ Identitas pengirim: setiap pesan disertai username pengirim

### Fitur Opsional
- ✅ Validasi uniqueness username
- ✅ Notifikasi ketika client baru bergabung (beserta identitasnya)
- ✅ Notifikasi ketika client meninggalkan server (beserta identitasnya)
- ✅ Fitur room: `/join` dan `/leave`; pesan hanya tersebar di room yang sama

## Cara Menjalankan

### 1. Jalankan Server
```bash
go run server/main.go
```

### 2. Jalankan Client (buka terminal baru untuk setiap client)
```bash
go run client/main.go
```

## Perintah yang Tersedia

| Perintah | Keterangan |
|----------|------------|
| `/join <nama_room>` | Pindah ke room tertentu |
| `/leave` | Kembali ke room lobby |
| `/rooms` | Lihat daftar room aktif |
| `/quit` | Keluar dari server |

## Struktur Proyek

```
├── server/
│   └── main.go      # Server TCP dengan goroutine per client
├── client/
│   └── main.go      # Client TCP dengan goroutine penerima pesan
└── README.md
```

## Teknologi

- Bahasa: Go (Golang)
- Hanya menggunakan standard library (`net`, `bufio`, `fmt`, `os`, `strings`, `sync`)
- Tidak menggunakan library eksternal
