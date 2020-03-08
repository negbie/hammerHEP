package main

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
	"unicode"

	"go.uber.org/ratelimit"
)

// Hammer container
type Hammer struct {
	addr  string
	trans []Transport
	proto string
	rate  int
}

// Transport for Packet
type Transport struct {
	name   string
	conn   net.Conn
	pipe   chan Packet
	buffer *bufio.Writer
	errCnt int
}

// Packet payload and length
type Packet struct {
	payload []byte
	length  int
}

func cutSpace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

// NewHammer constructor
func NewHammer(proto, addr, port, trans string, rate int) (*Hammer, error) {
	t := strings.Split(strings.ToLower(cutSpace(trans)), ",")
	l := len(t)
	h := &Hammer{
		addr:  strings.ToLower(cutSpace(addr + ":" + port)),
		trans: make([]Transport, l),
		proto: strings.ToLower(proto),
		rate:  rate / l,
	}
	for k, v := range t {
		h.trans[k].name = v
		h.trans[k].pipe = make(chan Packet, 1e6)
		if err := h.connect(k); err != nil {
			return nil, err
		}
	}
	return h, nil
}

func (h *Hammer) connect(k int) (err error) {
	if h.trans[k].name == "udp" {
		if h.trans[k].conn, err = net.Dial("udp", h.addr); err != nil {
			return fmt.Errorf("dial transport failed: %s", err.Error())
		}
	} else if h.trans[k].name == "tcp" {
		if h.trans[k].conn, err = net.Dial("tcp", h.addr); err != nil {
			return fmt.Errorf("dial transport failed: %s", err.Error())
		}
	} else if h.trans[k].name == "tls" {
		if h.trans[k].conn, err = tls.Dial("tcp", h.addr, &tls.Config{InsecureSkipVerify: true}); err != nil {
			return fmt.Errorf("dial transport failed: %s", err.Error())
		}
	} else {
		return fmt.Errorf("unsupported transport: %s", h.trans[k].name)
	}
	h.trans[k].buffer = bufio.NewWriterSize(h.trans[k].conn, 8192)
	return nil
}

func (h *Hammer) reconnect(k int) error {
	if err := h.connect(k); err != nil {
		return err
	}
	h.trans[k].buffer.Reset(h.trans[k].conn)
	return nil
}

// Hammer time
func (h *Hammer) start() {
	var wg sync.WaitGroup
	var packets = buildPackets(h.proto)

	for k := range h.trans {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			for {
				pkt := <-h.trans[k].pipe
				h.trans[k].buffer.Write(pkt.payload[:pkt.length])
				err := h.trans[k].buffer.Flush()
				if err != nil {
					h.trans[k].errCnt++
					if h.trans[k].errCnt%64 == 0 {
						h.trans[k].errCnt = 0
						fmt.Println(err)
						if err = h.reconnect(k); err != nil {
							fmt.Println("reconnect error:", err)
						}
					}
				}
			}
		}(k)

		time.Sleep(200 * time.Millisecond)

		wg.Add(1)
		go func(t Transport) {
			defer wg.Done()
			send(h.proto, h.rate, t.pipe, packets)
		}(h.trans[k])
	}

	wg.Wait()
}

func send(proto string, rate int, ch chan Packet, packets []Packet) {
	var limit ratelimit.Limiter

	if rate > 0 {
		limit = ratelimit.New(rate)
	}

	for {
		for i := range packets {
			limit.Take()
			ch <- packets[i]
		}
	}
}

func buildPackets(proto string) []Packet {
	packets := []Packet{}
	msg := [][]byte{}

	switch proto {
	case "ipfix":
		msg = fakeIPFIX
	default:
		msg = fakeHEP
		for i := 0; i < len(msg); i++ {
			binary.BigEndian.PutUint32(msg[i][62:66], uint32(0))
			binary.BigEndian.PutUint32(msg[i][66:70], uint32(0))
		}
	}

	for i := 0; i < len(msg); i++ {
		data := msg[i]
		payload := make([]byte, 8192)
		copy(payload[:], data)
		packets = append(packets, Packet{payload: payload[:len(data)], length: len(data)})
	}
	return packets
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(len int) []byte {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return bytes
}
