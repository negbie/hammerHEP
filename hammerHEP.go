package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// Hammer container
type Hammer struct {
	conn net.Conn
	ch   chan Packet
	rate int
}

// Packet payload and length
type Packet struct {
	payload []byte
	length  int
}

// NewHammer setup
func NewHammer(prot string, addr string, port string) (*Hammer, error) {

	switch prot {
	case "ipfix":
		dest, err := net.ResolveTCPAddr("tcp", addr+":"+port)
		if err != nil {
			println("Resolve transport failed:", err.Error())
			os.Exit(1)
		}

		c, err := net.DialTCP("tcp", nil, dest)
		if err != nil {
			println("Dial failed:", err.Error())
			os.Exit(1)
		}
		if err != nil {
			return nil, err
		}

		h := &Hammer{
			conn: c,
			ch:   make(chan Packet, 15000),
			rate: 15000,
		}
		return h, nil
	default:
		dest, err := net.ResolveUDPAddr("udp", addr+":"+port)
		if err != nil {
			println("Resolve transport failed:", err.Error())
			os.Exit(1)
		}

		c, err := net.DialUDP("udp", nil, dest)
		if err != nil {
			println("Dial failed:", err.Error())
			os.Exit(1)
		}
		if err != nil {
			return nil, err
		}

		h := &Hammer{
			conn: c,
			ch:   make(chan Packet, 15000),
			rate: 15000,
		}
		return h, nil

	}

}

// Hammer time
func (h *Hammer) start() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		var p Packet
		defer wg.Done()
		for {
			p = <-h.ch
			h.conn.Write(p.payload[:p.length])
		}
	}()

	time.Sleep(1 * time.Second)

	wg.Add(1)
	go func() {
		defer wg.Done()
		h.send()
	}()
	wg.Wait()
}

func (h *Hammer) send() {
	var (
		limit   <-chan time.Time
		packets = h.make()
	)

	if h.rate > 0 {
		limit = time.Tick(time.Duration(1000000/(h.rate)) * time.Microsecond)
	}

	for {
		for i := range packets {
			<-limit
			h.ch <- packets[i]
		}
	}
}

func (h *Hammer) make() []Packet {
	packets := []Packet{}
	msg := [][]byte{}

	for i := 0; i < len(msg); i++ {
		data := msg[i]

		payload := make([]byte, 4096)
		copy(payload[:], data)
		packets = append(packets, Packet{payload: payload[:len(data)], length: len(data)})
	}
	return packets
}

func main() {
	var (
		wg   sync.WaitGroup
		port = flag.String("port", "9060", "Port to send packets")
		addr = flag.String("addr", "localhost", "Address to send packets")
		rate = flag.Int("rate", 1, "How many packets per second to send")
		prot = flag.String("prot", "hep", "Supported protocols are hep,ipfix")
	)
	flag.Parse()

	wg.Add(1)
	go func() {
		defer wg.Done()
		hammer, err := NewHammer(*prot, *addr, *port)
		hammer.rate = *rate
		if err != nil {
			os.Exit(1)
		}
		hammer.start()
	}()

	fmt.Printf("Hammer down: %s\n", *addr+":"+*port)
	wg.Wait()
}
