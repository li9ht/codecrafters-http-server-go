package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

var directory string

func init() {
    flag.StringVar(&directory, "directory", "/default/path", "dir")
}

func main() {
	flag.Parse()
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

    req := make([]byte, 4096)
    conn.Read(req)

    if strings.HasPrefix(string(req), "GET /user-agent") {
        sendUserAgentResponse(conn, req)
        return
    }
    if strings.HasPrefix(string(req), "GET /echo/") {
        sendEchoResponse(conn, req)
        return
    }
    if strings.HasPrefix(string(req), "GET /files/") {	
		if directory == "" {
			log.Fatal("You must specify a directory with --directory")
		}
        sendFileResponse(conn, req)
        return
    }
	if strings.HasPrefix(string(req), "POST /files/") {
		if directory == "" {
			log.Fatal("You must specify a directory with --directory")
		}
		savePostFile(conn, req)
		return
	}
    if strings.HasPrefix(string(req), "GET / HTTP/1.1") {
        writeResponse(conn, "HTTP/1.1 200 OK\r\n\r\n", nil)
        return
    }
    write404(conn)
}

func savePostFile(conn net.Conn, req []byte) {
	filename := strings.TrimPrefix(strings.Split(string(req), " ")[1], "/files/")
	filePath := directory + filename
	fileContent := strings.Split(string(req), "\r\n\r\n")[1]
	err := ioutil.WriteFile(filePath, []byte(fileContent), 0644)
	if err != nil {
		log.Printf("Error writing file: %v", err)
		write500(conn)
		return
	}
	writeResponse(conn, "HTTP/1.1 201 Created\r\n\r\n", nil)
}

func sendFileResponse(conn net.Conn, req []byte) {
	
    filename := strings.TrimPrefix(strings.Split(string(req), " ")[1], "/files/")
    filePath := directory + filename
	// log.Printf("Accessing file: %s", filePath)

    fileInfo, err := os.Stat(filePath)
    if err != nil {
		// log.Printf("Error: %v", err)
        write404(conn)
        return
    }

    fileContent, err := ioutil.ReadFile(filePath)
    if err != nil {
        write500(conn)
        return
    }

    responseHeader := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", fileInfo.Size())
    writeResponse(conn, responseHeader, fileContent)
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

func write500(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
}

func write404(conn net.Conn) {
	const response = "HTTP/1.1 404 Not Found\r\n\r\n"
    _, err := conn.Write([]byte(response))
    if err != nil {
        log.Printf("Error sending 404 response: %v", err)
    }
}

func writeResponse(conn net.Conn, header string, body []byte) {
    conn.Write([]byte(header))
    if body != nil {
        conn.Write(body)
    }
}