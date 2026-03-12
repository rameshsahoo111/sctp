package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
)

func main() {
	port := flag.Int("port", 5000, "SCTP server port to listen on")
	flag.Parse()

	// 1. Create a raw OS socket for SCTP
	// AF_INET = IPv4, SOCK_STREAM = Stream connection, IPPROTO_SCTP (132) = SCTP Protocol
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_SCTP)
	if err != nil {
		log.Fatalf("Failed to create SCTP socket: %v", err)
	}
	defer syscall.Close(fd)

	// 2. Bind the socket to 0.0.0.0 and the specified port
	addr := &syscall.SockaddrInet4{Port: *port}
	copy(addr.Addr[:], []byte{0, 0, 0, 0}) // 0.0.0.0

	if err := syscall.Bind(fd, addr); err != nil {
		log.Fatalf("Bind failed: %v", err)
	}

	// 3. Start listening (the second argument is the backlog queue size)
	if err := syscall.Listen(fd, 128); err != nil {
		log.Fatalf("Listen failed: %v", err)
	}

	fmt.Printf("SCTP server listening on 0.0.0.0:%d (via syscall)\n", *port)

	// 4. Accept loop
	for {
		clientFd, clientAddr, err := syscall.Accept(fd)
		if err != nil {
			log.Printf("Accept failed: %v", err)
			continue
		}

		go handleClient(clientFd, clientAddr)
	}
}

func handleClient(fd int, addr syscall.Sockaddr) {
	// Always close the file descriptor when done
	defer syscall.Close(fd)

	// Extract the client IP/Port from the raw Sockaddr struct
	var clientIP string
	if inet4, ok := addr.(*syscall.SockaddrInet4); ok {
		clientIP = fmt.Sprintf("%d.%d.%d.%d:%d",
			inet4.Addr[0], inet4.Addr[1], inet4.Addr[2], inet4.Addr[3], inet4.Port)
	}

	// Extract the server's local IP/Port for this specific connection
	var serverIP string
	localAddr, err := syscall.Getsockname(fd)
	if err == nil {
		if inet4, ok := localAddr.(*syscall.SockaddrInet4); ok {
			serverIP = fmt.Sprintf("%d.%d.%d.%d:%d",
				inet4.Addr[0], inet4.Addr[1], inet4.Addr[2], inet4.Addr[3], inet4.Port)
		}
	} else {
		serverIP = "unknown-ip"
	}

	// Read data from the socket
	buf := make([]byte, 1024)
	n, err := syscall.Read(fd, buf)
	if err != nil {
		log.Printf("Read failed: %v", err)
		return
	}

	fmt.Printf("Received from %s: %s\n", clientIP, string(buf[:n]))

	// Fetch hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}

	// Send reply including Server Hostname, Server IP, and Client IP
	replyMsg := fmt.Sprintf("Hello from SCTP Server! My Hostname: %s, My IP: %s | Your Client IP: %s\n", hostname, serverIP, clientIP)
	
	_, err = syscall.Write(fd, []byte(replyMsg))
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
}
