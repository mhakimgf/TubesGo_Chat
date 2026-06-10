package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Koneksi ke server — pola dari contoh dosen (net.Dial)
	conn, err := net.Dial("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to server: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connected to server!")

	connReader := bufio.NewReader(conn) // pola dari contoh dosen

	// === FASE REGISTRASI ===
	// Baca prompt "Enter username:" dari server
	prompt, err := connReader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from server: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(prompt)

	// Baca username dari stdin
	localReader := bufio.NewReader(os.Stdin) // pola dari contoh dosen
	username, err := localReader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading username: %v\n", err)
		os.Exit(1)
	}

	// Kirim username ke server
	conn.Write([]byte(username)) // pola dari contoh dosen

	// Baca balasan server (Welcome atau Username sudah dipakai)
	response, err := connReader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading server response: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(response)

	// Cek apakah username ditolak
	if strings.Contains(response, "[SERVER]: Username sudah dipakai") {
		os.Exit(1)
	}

	// === GOROUTINE: TERIMA PESAN DARI SERVER ===
	go receiveMessages(conn, connReader)

	// === LOOP KIRIM PESAN (goroutine utama) ===
	sendMessages(conn)
}

// receiveMessages terus membaca pesan dari server dan menampilkannya ke terminal.
func receiveMessages(conn net.Conn, reader *bufio.Reader) {
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected from server.")
			os.Exit(0)
		}
		fmt.Print(message)
	}
}

// sendMessages membaca input dari stdin dan mengirimkannya ke server.
func sendMessages(conn net.Conn) {
	localReader := bufio.NewReader(os.Stdin) // pola dari contoh dosen

	for {
		message, err := localReader.ReadString('\n')
		if err != nil {
			break
		}
		_, err = conn.Write([]byte(message)) // pola dari contoh dosen
		if err != nil {
			fmt.Println("Error sending message.")
			break
		}
	}
}
