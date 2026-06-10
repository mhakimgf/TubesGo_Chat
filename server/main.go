package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// Client merepresentasikan satu koneksi aktif.
// Setiap field diperlukan untuk broadcast dan identifikasi.
type Client struct {
	conn     net.Conn    // koneksi TCP — digunakan untuk fmt.Fprintf(client.conn, ...)
	username string      // identitas unik, diisi saat registrasi
	room     string      // room aktif, default "lobby"
	send     chan string  // channel buffer — goroutine writeToClient membacanya
}

// State global server
var (
	clients = make(map[string]*Client)   // key: username
	rooms   = make(map[string][]*Client) // key: nama room
	mu      sync.RWMutex                 // proteksi akses ke kedua map
)

func main() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Println("Listening...")

	// Inisialisasi room default "lobby"
	rooms["lobby"] = []*Client{}

	// Loop tak terbatas menerima koneksi — bukan sekali Accept seperti echo-example
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection: %v\n", err)
			continue
		}
		fmt.Println("New connection accepted!")
		go handleClient(conn) // goroutine baru per client (WAJIB sesuai soal)
	}
}

// handleClient adalah goroutine utama per client.
// Menangani registrasi username, lalu loop baca pesan.
func handleClient(conn net.Conn) {
	reader := bufio.NewReader(conn) // pola dari contoh dosen

	// === FASE REGISTRASI ===
	fmt.Fprint(conn, "Enter username:\n")

	username, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return
	}
	username = strings.TrimSpace(username)

	// Validasi: username tidak boleh kosong
	if username == "" {
		fmt.Fprintln(conn, "[SERVER]: Username tidak boleh kosong.")
		conn.Close()
		return
	}

	// Cek uniqueness username
	mu.Lock()
	if _, exists := clients[username]; exists {
		mu.Unlock()
		fmt.Fprintln(conn, "[SERVER]: Username sudah dipakai.")
		conn.Close()
		return
	}

	// Buat client baru dan daftarkan
	client := &Client{
		conn:     conn,
		username: username,
		room:     "lobby",
		send:     make(chan string, 256),
	}
	clients[username] = client
	rooms["lobby"] = append(rooms["lobby"], client)
	mu.Unlock()

	// Jalankan goroutine penulis — satu-satunya goroutine yang menulis ke conn
	go writeToClient(client)

	// Kirim pesan selamat datang
	fmt.Fprintf(conn, "Welcome, %s! You are in room: lobby\n", username)

	// Broadcast notifikasi join ke semua client di lobby
	broadcast(client.room, "SERVER", fmt.Sprintf("[SERVER]: %s has joined!", username))

	// === DEFER: CLEANUP SAAT DISCONNECT ===
	defer func() {
		oldRoom := client.room

		mu.Lock()
		delete(clients, username)
		removeFromRoom(client)
		close(client.send) // sinyal writeToClient untuk berhenti
		mu.Unlock()

		conn.Close()

		// Broadcast notifikasi leave
		broadcast(oldRoom, "SERVER", fmt.Sprintf("[SERVER]: %s has left.", username))
	}()

	// === LOOP BACA PESAN ===
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			// ERROR/EOF -> break -> trigger defer -> cleanup otomatis
			break
		}

		message = strings.TrimSpace(message)

		if message == "" {
			continue // abaikan pesan kosong
		}

		// Cek apakah pesan adalah command (diawali '/')
		if strings.HasPrefix(message, "/") {
			handleCommand(client, message)
		} else {
			// Broadcast pesan biasa ke semua client di room yang sama
			broadcast(client.room, client.username,
				fmt.Sprintf("[%s]: %s", client.username, message))
		}
	}
}

// writeToClient adalah goroutine penulis.
// Satu-satunya goroutine yang boleh menulis ke client.conn.
// Ini memastikan tidak ada interleaved write ke koneksi TCP.
func writeToClient(client *Client) {
	for msg := range client.send {
		// range otomatis keluar saat channel ditutup (close(client.send))
		_, err := fmt.Fprintln(client.conn, msg)
		if err != nil {
			break // koneksi bermasalah
		}
	}
}

// broadcast mengirim pesan ke semua client di satu room,
// kecuali pengirim (sender).
func broadcast(room, sender, message string) {
	mu.RLock()
	targets := rooms[room]
	for _, c := range targets {
		if c.username != sender {
			// Non-blocking send: jika channel penuh (client lambat), skip
			select {
			case c.send <- message:
			default:
				// skip client yang lambat agar tidak memblok broadcast
			}
		}
	}
	mu.RUnlock()
}

// handleCommand memproses pesan yang diawali '/'.
func handleCommand(client *Client, cmd string) {
	parts := strings.Fields(cmd)

	switch parts[0] {
	case "/join":
		if len(parts) < 2 {
			fmt.Fprintln(client.conn, "[SERVER]: Usage: /join <nama_room>")
			return
		}
		newRoom := parts[1]
		oldRoom := client.room

		mu.Lock()
		removeFromRoom(client)
		client.room = newRoom
		rooms[newRoom] = append(rooms[newRoom], client)
		mu.Unlock()

		// Broadcast leave notification ke room lama
		broadcast(oldRoom, "SERVER",
			fmt.Sprintf("[SERVER]: %s left room %s", client.username, oldRoom))

		// Broadcast join notification ke room baru
		broadcast(newRoom, "SERVER",
			fmt.Sprintf("[SERVER]: %s joined room %s", client.username, newRoom))

		// Konfirmasi ke client
		fmt.Fprintf(client.conn, "[SERVER]: You are now in room: %s\n", newRoom)

	case "/leave":
		// /leave = kembali ke lobby
		handleCommand(client, "/join lobby")

	case "/rooms":
		mu.RLock()
		var roomNames []string
		for name, members := range rooms {
			roomNames = append(roomNames, fmt.Sprintf("%s(%d)", name, len(members)))
		}
		mu.RUnlock()
		fmt.Fprintln(client.conn, "[SERVER]: Active rooms: "+strings.Join(roomNames, ", "))

	case "/quit":
		// Tutup koneksi — trigger EOF di ReadString -> defer cleanup
		client.conn.Close()

	default:
		fmt.Fprintln(client.conn, "[SERVER]: Unknown command. Available: /join <room>, /leave, /rooms, /quit")
	}
}

// removeFromRoom menghapus client dari room-nya saat ini.
// HARUS dipanggil saat mu.Lock() sudah aktif oleh caller.
func removeFromRoom(client *Client) {
	list := rooms[client.room]
	for i, c := range list {
		if c.username == client.username {
			rooms[client.room] = append(list[:i], list[i+1:]...)
			break
		}
	}
	// Bersihkan room kosong (kecuali lobby)
	if len(rooms[client.room]) == 0 && client.room != "lobby" {
		delete(rooms, client.room)
	}
}
