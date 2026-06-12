package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	//melakukan connect ke port 9090 sesuai dengan yang akan di listen oleh server
	conn, err := net.Dial("tcp", ":9090")
	//jika terdapat error maka tampilkan pesan errornya menggunakan os.stderr dan hentikan si program
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to server: %v\n", err)
		os.Exit(1)
	}
	//defer digunakan untuk memastikan bahwa koneksi akan ditutup sebelum fungsinya selesai.
	//memastikan bahwa koneksi dari client sudah ditutup
	defer conn.Close()

	fmt.Println("Connected to server!")

	//membuat buffer untuk conn
	connReader := bufio.NewReader(conn)

	//ReadString digunakan untuk membaca suatu teks sampai menemukan karakter yang ditulis di parameter
	//bagian ini digunakan untuk mengambil pesan/teks "enter username" di function handleClient
	//kenapa pake \n karena di server tuh pake fmt.Fprintln yang di belakangnya pasti ada \n
	prompt, err := connReader.ReadString('\n')
	//jika terdapat error maka akan menampilkan pesan error dan menghentikan si program
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from server: %v\n", err)
		os.Exit(1)
	}
	//print teks yang didapatkan dari server
	fmt.Print(prompt)

	//membuat local reader yang akan membaca input user dari terminal (os.stdin)
	localReader := bufio.NewReader(os.Stdin)
	//masukan input yang ditulis user ke dalam variabel username
	username, err := localReader.ReadString('\n')
	//jika terdapat error maka tampilkan pesan error dan hentikan si program
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading username: %v\n", err)
		os.Exit(1)
	}

	//jika tidak ada error, kirimkan username tersebut ke server dengan menggunakan .Write
	//pesan dikirim dalam byte karena tcp hanya bisa membaca byte
	conn.Write([]byte(username))

	//Sama seperti proses sebelumnya. gunakan ReadString untuk mengambil pesan yang diberikan oleh server
	//ReadString di bagian ini dipakai untuk ngambil response validasi username di server yang ada dua skenario
	//skenario sukses(welcome .... you are in ....) atau skenario gagal(username sudah dipakai/tidak boleh kosong)
	response, err := connReader.ReadString('\n')
	//Jika ada error, tampilin pesan dan hentikan program
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading server response: %v\n", err)
		os.Exit(1)
	}

	//print response yang didapat
	fmt.Print(response)

	//Cek response dari server, jika username sudah dipakai maka akan hentikan program
	if strings.Contains(response, "[SERVER]: Username sudah dipakai") {
		os.Exit(1)
	}

	//menggunakan goroutine untuk receivemessage
	go receiveMessages(connReader)

	//memanggil fungsi sendMessages
	sendMessages(conn)
}

// Functon receiveMessages berguna buat membaca pesan dari server dan ditampilkan ke terminal
func receiveMessages(reader *bufio.Reader) {

	//infinite loop untuk membaca pesan-pesan yang diterima dari server
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected from server.")
			os.Exit(0)
		}
		fmt.Print(message)
	}
}

// Function sendMessages berguna buat baca input user(os.Stdin) dan mengirim pesan tersebut ke server
func sendMessages(conn net.Conn) {
	//localReader dipakai untuk mengambil input dari user
	localReader := bufio.NewReader(os.Stdin)

	//infinite loop yang digunakan untuk menangkap semua pesan yang diketikan oleh user
	for {
		//cara bacanya adalah dengan mengambil pesan sampai ke \n
		message, err := localReader.ReadString('\n')
		if err != nil {
			break
		}
		//mengirimkan pesan ke server dengan menggunakan Write
		//pesan dikirim dalam byte karena tcp hanya bisa membaca byte
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending message.")
			break
		}
	}
}
