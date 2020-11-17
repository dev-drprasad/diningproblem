package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
)

type message struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

var seating = map[uint]string{1: "", 2: "", 3: "", 4: "", 5: ""}

func getDots(status string, position uint) []string {
	dots := []string{"........", "........", "........", "........", "........"}
	if len(status) > 8 {
		status = status[:8]
	}
	dots[position-1] = fmt.Sprintf("%8s", strings.ToUpper(status))
	return dots
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	var listenAddr string
	var debug bool
	flag.StringVar(&listenAddr, "listen", "", "listen address <ip>:<port> of this philosopher")
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.Parse()

	s, err := net.ResolveUDPAddr("udp4", listenAddr)
	if err != nil {
		log.Fatalln(err)
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		log.Fatalln(err)
	}
	defer connection.Close()

	buffer := make([]byte, 1024)

	for {
		if debug {
			log.Println("reading from udp")
		}
		n, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			log.Println("failed to read from udp", err)
		}
		b := buffer[0 : n-1]
		if debug {
			log.Println("message ", string(b))
		}
		var m message
		if err := json.Unmarshal(b, &m); err != nil {
			log.Println("error unmarshalling message", err)
		}

		if debug {
			log.Println("status is ", m.Status)
		}
		switch m.Status {
		case "dine":
			log.Printf("%8s %8s %8s %8s %8s", "1  ", "2  ", "3  ", "4  ", "5  ")
			break
		case "ping":
			for k, v := range seating {
				if v == "" {
					seating[k] = m.Name
					log.Println(m.Name+" reached to restaurant and sat at ", k)
					break
				}
			}
			break
		default:
			var pos uint
			for k, v := range seating {
				if m.Name == v {
					pos = k
					break
				}
			}

			log.Println(getDots(m.Status, pos))
		}

	}
}
