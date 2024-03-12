package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Config struct {
	Host        string
	Port        int
	ReadTimeout Duration
}

var config Config

const configFileName = "config.json"

func main() {
	// create default config
	if _, err := os.Stat(configFileName); err != nil {
		log.Println("Configuration file not found, generating a default one...")
		f, err := os.Create(configFileName)
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		enc.Encode(Config{Host: "0.0.0.0", Port: 7890, ReadTimeout: Duration{time.Duration(3 * time.Second)}})
		os.Exit(0)
	}

	// setup logging
	logFile, err := os.Create("log.txt")
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.Println("Started logging")

	// read config
	f, err := os.Open(configFileName)
	if err != nil {
		log.Fatalln(err)
	}
	d, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}

	json.Unmarshal(d, &config)

	log.Printf("Read config: %+v", config)

	// start server
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		log.Fatalln(err)
	}
	defer listen.Close()

	log.Printf("Started a TCP server on %s", listen.Addr())

	// listen for connections and handle them
	connectionId := 0
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %s", err)
			continue
		}
		connectionId += 1
		log.Printf("Got a connection from %s. Assigning connId %d", conn.RemoteAddr(), connectionId)
		go handleConnection(conn, connectionId)
	}
}

func handleConnection(conn net.Conn, connId int) {
	defer conn.Close()
	log.Printf("[%d] Handling client connection", connId)

	ctx, cancel := context.WithTimeout(context.Background(), config.ReadTimeout.Duration)
	defer cancel()
	deadline, _ := ctx.Deadline()

	conn.SetReadDeadline(deadline)

	const stringToReturn = "\nСервер написан Романовым Д.И. М3О-107Б-23"
	// leave some space for our string in case the input is too big
	buffer := make([]byte, 1024, 1024+len(stringToReturn))

	bytesRead, err := conn.Read(buffer)
	if err != nil && err != io.EOF {
		log.Printf("[%d] Failed to read from client: %s", connId, err)
		return
	}

	log.Printf("[%d] Read client data: %s", connId, buffer[:bytesRead])

	runes := bytes.Runes(buffer[:bytesRead])

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	copy(buffer, cases.Title(language.English).String(strings.ToLower(string(runes))))

	// shrink the buffer to desired length
	buffer = buffer[:bytesRead+len(stringToReturn)]

	log.Printf("[%d] Reversed the input string: %s", connId, buffer[:bytesRead])
	copy(buffer[bytesRead:], []byte(stringToReturn))
	log.Printf("[%d] Appended to the input string: %s", connId, buffer)

	time.Sleep(4 * time.Second)

	log.Printf("[%d] Sending a response back", connId)
	bytesSent, err := conn.Write(buffer)
	if err != nil && err != io.EOF {
		log.Printf("[%d] Failed to write response: %s", connId, err)
		return
	}
	log.Printf("[%d] Wrote %d bytes", connId, bytesSent)
}
