package main

import (
	"bufio"
	udpclient "diningproblem/udpclient"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// time to retry if something didn't work
var retrywait time.Duration = 500 * time.Millisecond

// philosopher will take min-max milliseconds to eat and think
var laziness = [2]int{200, 500}

// no of times to eat and think before done
var cycles int = 15

type Stalker struct {
	address   string
	theirAddr string
	debug     bool
	udpclient.Client
}

func randDuration(minmax [2]int) time.Duration {
	r := rand.Intn(minmax[1]-minmax[0]) + minmax[0]
	return time.Duration(r) * time.Millisecond
}

func (s *Stalker) stalk(status string) {
	if err := s.Init(s.address); err != nil {
		log.Println("stalking failed. ignoring error", err)
	}

	m := map[string]string{"name": s.theirAddr, "status": status}
	b, _ := json.Marshal(m)
	if s.debug {
		log.Println("sending ", string(b))
	}
	err := s.Send(b)
	if err != nil && s.debug {
		log.Println("stalk failed ", err)
	}
}

func makeConn(address string) (net.Conn, error) {
	return net.Dial("tcp", address)
}

func makeConnIndefinite(address string) *net.Conn {
	var c net.Conn
	var err error

	for {
		c, err = makeConn(address)
		if err != nil {
			log.Println("😠 this person alway comes late : ", err)
		} else {
			break
		}
		time.Sleep(retrywait)
	}
	return &c
}

type Philosopher struct {
	name          string
	address       string
	initiator     bool
	debug         bool
	leftforkaddr  string
	rightforkaddr string

	leftfork  chan bool
	rightfork chan bool

	neighborconn  *net.Conn
	leftforkconn  *net.Conn
	rightforkconn *net.Conn
}

func NewPhilosopher(name string, listenAt string, rightForkAt string, leftForkAt string, neightborAt string, initiator bool, debug bool) *Philosopher {

	p := Philosopher{
		name: name, address: listenAt,
		initiator:     initiator,
		debug:         debug,
		leftforkaddr:  leftForkAt,
		rightforkaddr: rightForkAt,
	}

	// don't wait for friends and forks. lets settle down first
	go p.listen(listenAt)

	leftfork := make(chan bool, 1)
	rightfork := make(chan bool, 1)
	p.leftfork = leftfork
	p.rightfork = rightfork

	p.neighborconn = makeConnIndefinite(neightborAt)
	return &p
}

func (p *Philosopher) listen(address string) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln("Why can't I have address ? ", err)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("❗️ couldn't listen for new conn : ", err)
			time.Sleep(retrywait)
			continue
		}

		remoteAddr := c.RemoteAddr().String()
		for {
			message, err := bufio.NewReader(c).ReadString('\n')
			if err != nil {
				log.Println("❗️ couldn't read from conn : ", err)
				c.Close()
				if err == io.EOF {
					break
				}
				time.Sleep(retrywait)
				continue
			}

			message = strings.TrimSpace(message)
			if p.debug {
				log.Println(remoteAddr+" ->: ", message)
			}
			switch message {
			case "ALL JOINED ?":
				if p.initiator {
					c.Write([]byte("OK\n"))
					log.Println("🙌 Every joined at diner. Time to eat 😋")
					go p.nudge("LET's EAT")
					allJoined = true // this stops asking "ALL JOINED ?". Look in main function
					stalker.stalk("dine")
				} else {
					if p.neighborconn != nil {
						p.nudge(message)
						if _, err := c.Write([]byte("OK\n")); err != nil {
							panic(err)
						}
					} else {
						log.Println("🙄 skipping. no neighbour present to ask ", message)
						if _, err := c.Write([]byte("OK\n")); err != nil {
							panic(err)
						}
					}
				}
			case "LET's EAT":
				// assuming all neighbor's available
				if !p.initiator {
					if p.neighborconn != nil {
						p.nudge(message)
					}
				}
				go p.dine()
				if _, err := c.Write([]byte("OK\n")); err != nil {
					panic(err)
				}
			default:
				if _, err := c.Write([]byte("OK\n")); err != nil {
					panic(err)
				}
			}

		}
	}

}

func (p *Philosopher) putdownfork(forkaddr string) error {
	var forkconn net.Conn
	var err error
	retries := 3
	for {
		forkconn, err = makeConn(forkaddr)
		if err != nil {
			if retries == 0 {
				panic(err)
			} else {
				time.Sleep(retrywait)
				retries--
				continue
			}
		}
		break
	}
	defer (forkconn).Close()

	retries = 3
	message := "PUT DOWN"
	for {
		n, err := fmt.Fprintf(forkconn, p.name+":"+message+"\n")
		if err != nil {
			log.Println("❗️ failed to return fork. retrying...")
			time.Sleep(retrywait)
			if retries == 0 {
				return err
			}
			retries--
		} else {
			if p.debug {
				log.Println("wrote "+fmt.Sprintf("%d", n)+" bytes to conn for ", message)
			}
			break
		}
	}

	log.Println("⏱  waiting for response from " + (forkconn).RemoteAddr().String() + " for " + message)
	inmessage, err := bufio.NewReader(forkconn).ReadString('\n')
	if err != nil {
		log.Println((forkconn).RemoteAddr().String()+" ->: reply error for "+message+" : ", err)
		return err
	}

	inmessage = strings.TrimSpace(inmessage)
	if p.debug {
		log.Println((forkconn).RemoteAddr().String() + " ->: reply for " + message + " : " + inmessage)
	}

	return nil
}

func (p *Philosopher) getfork(forkaddr string) error {
	var forkconn net.Conn
	var err error
	retries := 3
	for {
		forkconn, err = makeConn(forkaddr)
		if err != nil {
			if retries == 0 {
				panic(err)
			} else {
				time.Sleep(retrywait)
				retries--
				continue
			}
		}
		break
	}
	defer (forkconn).Close()

	remoteAddr := (forkconn).RemoteAddr().String()

	message := "NEED FORK"
	_, err = fmt.Fprintf(forkconn, p.name+":"+message+"\n")
	if err != nil {
		return err
	}

	log.Println("⏳ waiting from fork " + remoteAddr + " for reply for " + message)
	inmessage, err := bufio.NewReader(forkconn).ReadString('\n')
	if err != nil {
		return err
	}

	inmessage = strings.TrimSpace(inmessage)
	if p.debug {
		log.Println(remoteAddr + " ->: " + inmessage)
	}

	log.Println("get fork returning")
	return nil
}

func (p *Philosopher) nudge(outmessage string) error {
	if p.debug {
		log.Println("saying to negihbor ", outmessage)
	}
	_, err := fmt.Fprintf(*p.neighborconn, outmessage+"\n")
	if err != nil {
		return err
	}

	message, err := bufio.NewReader(*p.neighborconn).ReadString('\n')
	if err != nil {
		return err
	}

	message = strings.TrimSpace(message)
	if p.debug {
		log.Println((*p.neighborconn).RemoteAddr().String() + " ->: reply for " + outmessage + " : " + message)
	}

	return nil
}

func (p *Philosopher) think() {
	stalker.stalk("thinking")
	time.Sleep(randDuration(laziness))
}

func (p *Philosopher) eat() {
	stalker.stalk("eating")
	time.Sleep(randDuration(laziness))
}

func (p *Philosopher) getforks() {
	stalker.stalk("waiting")
	var hasrightfork bool

	log.Println("⏳ waiting for right fork")
	err := p.getfork(p.rightforkaddr)
	if err != nil {
		log.Println("get right fork failed ", err)
	} else {
		hasrightfork = true
	}

	if hasrightfork {
		log.Println("👍🏼 got right fork. going for left fork")
		log.Println("⏳ waiting for left fork")
		errr := p.getfork(p.leftforkaddr)
		if errr != nil {
			log.Println("get left fork failed ", errr)
		} else {
			return
		}
		log.Println("👍🏼 got left fork")
	}

	if err != nil {
		p.think()
		p.getforks()
	} else if hasrightfork {
		if err := p.putdownfork(p.rightforkaddr); err != nil {
			panic(err)
		}
		p.think()
		p.getforks()
	}
}

func (p *Philosopher) putdownforks() {
	log.Println("🍴 putting down my forks")

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		if err := p.putdownfork(p.leftforkaddr); err != nil {
			panic(err)
		}
		wg.Done()
	}(&wg)

	go func(wg *sync.WaitGroup) {
		if err := p.putdownfork(p.rightforkaddr); err != nil {
			panic(err)
		}
		wg.Done()
	}(&wg)
	wg.Wait()
	log.Println("🤲 my hands are free now")
}

func (p *Philosopher) dine() {
	for i := 0; i < cycles; i++ {
		log.Println("🤔 thinking..., thinking...")

		p.think()
		log.Println("😋 Aaa! My brain tired. Lets give work to mouth.")

		log.Println("🍴 Asking for forks")
		p.getforks()

		log.Println("😍 Time to eat something...")
		p.eat()

		log.Println("🧆 I ate what served. Lets order new items and wait...")
		p.putdownforks()
	}
	log.Println("🥱 enough for the day. lets pay bill and go home and sleep")
	stalker.stalk("done")
	os.Exit(0)
}

var stalker *Stalker

var allJoined bool

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	rand.Seed(time.Now().UnixNano())

	var name string
	var listenAddr string
	var forkAddr string
	var fork2Addr string
	var neighbourAddr string
	var stalkerAddr string
	var theInitiator bool
	var debug bool
	flag.StringVar(&name, "name", "", "name  of this philosopher")
	flag.StringVar(&listenAddr, "listen", "", "listen address <ip>:<port> of this philosopher")
	flag.StringVar(&forkAddr, "fork", "", "address of fork")
	flag.StringVar(&fork2Addr, "fork2", "", "address of 2nd fork")
	flag.StringVar(&neighbourAddr, "neighbour", "", "address of next sittting philoshopher")
	flag.StringVar(&stalkerAddr, "stalker", "", "address of stalker (printer)")
	flag.BoolVar(&theInitiator, "the-initiator", false, "")
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.Parse()

	if name == "" {
		log.Fatalln("How others supposed to call me ? I need a name")
	}

	if listenAddr == "" {
		log.Fatalln("If there is no  address, others can't talk to me")
	}

	if forkAddr == "" || fork2Addr == "" {
		log.Fatalln("Obviously, I need 2 forks")
	}

	if neighbourAddr == "" {
		log.Fatalln("I need a friend to talk to, if I get bored with thinking")
	}

	if stalkerAddr == "" {
		log.Fatalln("Someone need to tell history to next generations")
	}

	log.Println("My name: ", name)
	log.Println("My address: ", listenAddr)
	log.Println("My fork: ", forkAddr)
	log.Println("My fork2: ", fork2Addr)
	log.Println("My neighbour: ", neighbourAddr)
	if theInitiator {
		log.Println("I am not shy to start conversation")
	}

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	stalker = &Stalker{address: stalkerAddr, debug: debug, theirAddr: listenAddr}
	stalker.stalk("ping")

	p := NewPhilosopher(name, listenAddr, forkAddr, fork2Addr, neighbourAddr, theInitiator, debug)
	log.Println("✨ I am ready")
	if theInitiator {
		for {
			if !allJoined {
				log.Println("🙌 initiating conversation")
				err := p.nudge("ALL JOINED ?")
				if err != nil {
					log.Println("⚠️ nudge failed ", err)
				}
				time.Sleep(retrywait)
			} else {
				break
			}
		}
	}

	<-stopCh
}
