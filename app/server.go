package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
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

	var isGzip bool = false
	headers, err := readHeaders(req)
	if err != nil {
		// Handle error, possibly send a 500 Internal Server Error response
		write500(conn)
		return
	}
	acceptEncoding, ok := headers["Accept-Encoding"]
	if ok && strings.Contains(acceptEncoding, "gzip") {
		isGzip = true
	}

    if strings.HasPrefix(string(req), "GET /user-agent") {
        sendUserAgentResponse(conn, req)
        return
    }
    if strings.HasPrefix(string(req), "GET /echo/") {
        sendEchoResponse(conn, req,isGzip)
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
        writeResponse(conn, "HTTP/1.1 200 OK\r\n\r\n", nil,isGzip)
        return
    }
    write404(conn)
}

func readHeaders(req []byte) (map[string]string, error) {
    headers := make(map[string]string)
    reqReader := bytes.NewReader(req)
    bufReader := bufio.NewReader(reqReader)

    _, err := bufReader.ReadString('\n')
    if err != nil {
        return nil, err 
    }

    for {
        line, err := bufReader.ReadString('\n')
        if err != nil {
            break 
        }
        line = strings.TrimSpace(line) 
        if line == "" {
            break 
        }

        parts := strings.SplitN(line, ":", 2)
        if len(parts) == 2 {
            key := strings.TrimSpace(parts[0])
            value := strings.TrimSpace(parts[1])
            headers[key] = value
        }
    }

    return headers, nil
}

func savePostFile(conn net.Conn, req []byte) {

	reqStr := string(req)
    contentLength := getContentLength(reqStr)
	bodyStartIndex := strings.Index(reqStr, "\r\n\r\n") + 4
    contentLengthInt, _ := strconv.Atoi(contentLength)
    requestBody := reqStr[bodyStartIndex : bodyStartIndex+contentLengthInt]

	filename := strings.TrimPrefix(strings.Split(string(req), " ")[1], "/files/")
	filePath := directory + filename
	err := ioutil.WriteFile(filePath, []byte(requestBody), 0644)
	if err != nil {
		log.Printf("Error writing file: %v", err)
		write500(conn)
		return
	}
	writeResponse(conn, "HTTP/1.1 201 Created\r\n\r\n", nil,false)
}

func getContentLength(reqStr string) string {
    lines := strings.Split(reqStr, "\r\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "Content-Length: ") {
            return strings.TrimPrefix(line, "Content-Length: ")
        }
    }
    return ""
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
    writeResponse(conn, responseHeader, fileContent,false)
}

func sendUserAgentResponse(conn net.Conn, req []byte) {
    userAgent := extractHeader(req, "User-Agent: ")
    response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
    conn.Write([]byte(response))
}

func sendEchoResponse(conn net.Conn, req []byte, useGzip bool) {
	fullPath := strings.Split(string(req), " ")[1]
	path := strings.Split(fullPath, "/")[2]
	// pathLength := len(path)
	// response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", pathLength, path)
	// conn.Write([]byte(response))
	header := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n"
	body := []byte(path)
	writeResponse(conn, header, body, useGzip)
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

func writeResponse(conn net.Conn, header string, body []byte,useGzip bool) {
	if useGzip {
		header += "Content-Encoding: gzip\r\n"
	}
	header += "Content-Encoding: text/plain\r\n"
	contentLength := len(body)
    header += "Content-Length: " + fmt.Sprint(contentLength) + "\r\n"
	header += "\r\n"

    conn.Write([]byte(header))
	conn.Write(body)
}