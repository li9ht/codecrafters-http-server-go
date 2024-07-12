package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	// Uncomment this block to pass the first stage
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	req := make([]byte, 1024)
	conn.Read(req)
	if strings.HasPrefix(string(req), "GET /user-agent") {
		lines := strings.Split(string(req), "\r\n")
		var userAgent string
		for _, line := range lines {
			if strings.HasPrefix(line, "User-Agent: ") {
				userAgent = strings.TrimPrefix(line, "User-Agent: ")
				break
			}
		}
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
		conn.Write([]byte(response))
		conn.Close()
		return
	}
	if strings.HasPrefix(string(req), "GET /echo/") {
		fullPath := strings.Split(string(req), " ")[1]
		path := strings.Split(fullPath, "/")[2]
		pathLength := len(path)
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", pathLength, path)))
		conn.Close()
		return
	}
	if !strings.HasPrefix(string(req), "GET / HTTP/1.1") {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		conn.Close()
		return
	}
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}
