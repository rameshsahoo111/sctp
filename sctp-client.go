package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"syscall"
)

func main() {
	// 1. Define command-line flags
	serverIP := flag.String("ip", "127.0.0.1", "SCTP server IP address (IPv4 or IPv6)")
	port := flag.Int("port", 5000, "SCTP server port")
	lport := flag.Int("lport", 0, "Local source port to bind to (0 for ephemeral)")
	msg := flag.String("msg", "Ping from native Go SCTP Client", "Message to send")
	flag.Parse()

	// Parse the provided IP address
	ip := net.ParseIP(*serverIP)
	if ip == nil {
		log.Fatalf("Invalid IP address provided: %s", *serverIP)
	}

	var fd int
	var err error
	var sockaddr syscall.Sockaddr

	// 2. Dynamically create an IPv4 or IPv6 socket based on the input IP
	if ip4 := ip.To4(); ip4 != nil {
		// It's an IPv4 address
		fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_SCTP)
		if err != nil {
			log.Fatalf("Failed to create IPv4 SCTP socket: %v", err)
		}

		// Bind to a specific local port if requested
		if *lport > 0 {
			bindSA := &syscall.SockaddrInet4{Port: *lport}
			// Leaving bindSA.Addr empty defaults to 0.0.0.0 (all interfaces)
			if err := syscall.Bind(fd, bindSA); err != nil {
				log.Fatalf("Failed to bind local IPv4 port %d: %v", *lport, err)
			}
		}

		sa := &syscall.SockaddrInet4{Port: *port}
		copy(sa.Addr[:], ip4)
		sockaddr = sa
	} else {
		// It's an IPv6 address
		fd, err = syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, syscall.IPPROTO_SCTP)
		if err != nil {
			log.Fatalf("Failed to create IPv6 SCTP socket: %v", err)
		}

		// Bind to a specific local port if requested
		if *lport > 0 {
			bindSA := &syscall.SockaddrInet6{Port: *lport}
			// Leaving bindSA.Addr empty defaults to :: (all interfaces)
			if err := syscall.Bind(fd, bindSA); err != nil {
				log.Fatalf("Failed to bind local IPv6 port %d: %v", *lport, err)
			}
		}

		sa := &syscall.SockaddrInet6{Port: *port}
		copy(sa.Addr[:], ip)
		sockaddr = sa
	}
	
	// Ensure the socket is always closed when the function exits
	defer syscall.Close(fd)

	if *lport > 0 {
		fmt.Printf("Bound to local port %d, connecting to %s:%d...\n", *lport, *serverIP, *port)
	} else {
		fmt.Printf("Using ephemeral local port, connecting to %s:%d...\n", *serverIP, *port)
	}

	// 3. Connect to the server
	if err := syscall.Connect(fd, sockaddr); err != nil {
		log.Fatalf("Connect failed: %v", err)
	}
	fmt.Println("Connected successfully!")

	// 4. Send the message payload
	payload := []byte(*msg)
	_, err = syscall.Write(fd, payload)
	if err != nil {
		log.Fatalf("Write failed: %v", err)
	}
	fmt.Printf("Sent: %s\n", *msg)

	// 5. Read the response (with a 3-second receive timeout)
	tv := syscall.Timeval{Sec: 3, Usec: 0}
	if err := syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv); err != nil {
		log.Printf("Warning: Failed to set receive timeout: %v", err)
	}

	buf := make([]byte, 1024)
	n, err := syscall.Read(fd, buf)
	if err != nil {
		log.Fatalf("Read failed (Timeout/Drop?): %v", err)
	}

	fmt.Printf("Server Reply: %s\n", string(buf[:n]))
}
