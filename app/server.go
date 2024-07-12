package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
    fmt.Println("Logs from your program will appear here!")

    l, err := net.Listen("tcp", "0.0.0.0:4221")
    if err != nil {
        fmt.Println("Failed to bind to port 4221")
        os.Exit(1)
    }
    defer l.Close()

    for {
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting connection: ", err.Error())
            continue
        }

        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    req := make([]byte, 1024)
    conn.Read(req)

    if strings.HasPrefix(string(req), "GET /user-agent") {
        sendUserAgentResponse(conn, req)
    } else if strings.HasPrefix(string(req), "GET /echo/") {
        sendEchoResponse(conn, req)
    } else if strings.HasPrefix(string(req), "GET / HTTP/1.1") {
        conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
    } else {
        conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
    }
}

func sendUserAgentResponse(conn net.Conn, req []byte) {
    userAgent := extractHeader(req, "User-Agent: ")
    response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
    conn.Write([]byte(response))
}

func sendEchoResponse(conn net.Conn, req []byte) {
    fullPath := strings.Split(string(req), " ")[1]
    path := strings.Split(fullPath, "/")[2]
    pathLength := len(path)
    conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", pathLength, path)))
}

func extractHeader(req []byte, headerName string) string {
    lines := strings.Split(string(req), "\r\n")
    for _, line := range lines {
        if strings.HasPrefix(line, headerName) {
            return strings.TrimPrefix(line, headerName)
        }
    }
    return ""
}