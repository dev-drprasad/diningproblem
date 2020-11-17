package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"golang.org/x/net/netutil"
)

var forkWaitDelay = [2]int{400, 4000}

func randDuration(minmax [2]int) time.Duration {
	r := rand.Intn(minmax[1]-minmax[0]) + minmax[0]
	return time.Duration(r) * time.Millisecond
}

type Fork struct {
	address   string
	free      bool
	available *chan bool
}

func NewFork(listenAddr string) *Fork {
	c := make(chan bool, 1)
	f := Fork{
		address:   listenAddr,
		free:      true,
		available: &c,
	}
	go func() { *f.available <- true }()
	go f.initListener(listenAddr)
	return &f
}

func (f *Fork) initListener(address string) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln("Why can't I have address ? ", err)
	}
	defer l.Close()

	l = netutil.LimitListener(l, 2)

	for {
		log.Println("waiting for new connections")
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go func(c *net.Conn) {
			conn := *c
			defer conn.Close()
			retries := 3
			var message string
			remoteAddr := conn.RemoteAddr().String()
			var rd *bufio.Reader
			for {
				log.Println("reading from ", remoteAddr)
				rd = bufio.NewReader(conn)
				message, err = rd.ReadString('\n')
				if err != nil {
					log.Println("read from "+remoteAddr+" failed : ", err)
					return
					conn.Close()
					if err == io.EOF {
						return
					} else {
						if retries == 0 {
							return
						}
						retries--
						continue
					}
				} else {
					break
				}
			}

			message = strings.TrimSpace(message)
			s := strings.SplitN(message, ":", 2)
			from, message := s[0], s[1]
			displayname := from + "(" + remoteAddr + ")"
			log.Println(displayname+" ->: ", string(message))

			switch message {
			case "NEED FORK":
				select {
				case <-*f.available:
					log.Println("fork availble. giving it to ", from)
					if n, err := conn.Write([]byte("TAKE IT\n")); err == nil {
						log.Printf("fork given to %s. wrote %d bytes", from, n)
					} else {
						// make it available again, if philospoher can't take it
						go func() {
							*f.available <- true
						}()
						log.Println("error giving to philosopher : ", err)
					}
					return
				case <-time.After(randDuration(forkWaitDelay)):
					log.Println("fork not available. try again")
					return
				}
			case "PUT DOWN":
				go func() {
					*f.available <- true
				}()
				log.Println("fork claimed from ", from)
				_, err := conn.Write([]byte("OK\n"))
				if err != nil {
					panic(err)
				}
			default:
			}
		}(&c)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	var listenAddr string

	flag.StringVar(&listenAddr, "listen", "", "listen address <ip>:<port> of this philosopher")
	flag.Parse()

	if listenAddr == "" {
		log.Fatalln("I know I am not a living thing, but still I need an address")
	}

	stopCh := make(chan struct{}, 1)

	NewFork(listenAddr)

	<-stopCh
}
