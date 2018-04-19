package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"time"
)

const (
	identPattern = "^\\s*([0-9]{1,5})\\s*,\\s*([0-9]{1,5})\\s*$"
	maxPort      = 0xFFFF
	minPort      = 0x0
)

var (
	re               *regexp.Regexp
	unknownError     = []byte("0 , 0 : ERROR : UNKNOWN-ERROR\r\n")
	invalidPortError = []byte("0 , 0 : ERROR : INVALID-PORT\r\n")
)

func strToPortNumber(val string) (uint16, error) {
	intVal, err := strconv.ParseUint(val, 10, 16)
	return uint16(intVal), err
}

func generateResponse(conn net.Conn, fromPortStr string, toPortStr string) []byte {
	var r *rand.Rand
	var randValue uint32
	var randValueHex string
	var err error

	_, err = strToPortNumber(fromPortStr)
	if err != nil {
		log.Printf("Could not parse fromPort: %v", err)
		return invalidPortError
	}

	_, err = strToPortNumber(toPortStr)
	if err != nil {
		log.Printf("Could not parse toPort: %v", err)
		return invalidPortError
	}

	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	randValue = r.Uint32()
	randValueHex = strconv.FormatUint(uint64(randValue), 16)
	return []byte(fmt.Sprintf("%s , %s : USERID : UNIX : %s\r\n", fromPortStr, toPortStr, randValueHex))
}

func handleRequest(conn net.Conn, req string) {
	var response []byte
	matches := re.FindStringSubmatch(req)
	if matches == nil || len(matches) != 3 {
		log.Printf("Invalid request: %s", req)
		response = unknownError
	} else {
		response = generateResponse(conn, matches[1], matches[2])
	}

	conn.Write(response)
}

func handleConnection(conn net.Conn) {
	buffer := bufio.NewReader(conn)
	request, err := buffer.ReadString('\n')
	if err != nil {
		log.Printf("Couldn't read request from client: %v", err)
		return
	}
	handleRequest(conn, request)
	conn.Close()
}

func main() {
	re = regexp.MustCompile(identPattern)

	ln, err := net.Listen("tcp", ":113")
	if err != nil {
		log.Fatalf("Couldn't open a listen socket: %v", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Couldn't accept connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}
