package sink

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func Ping(addr string) (*net.IPAddr, time.Duration, error) {
	dst, err := net.ResolveIPAddr("ip4", addr)
	if err != nil {
		fmt.Println("[PING] resolve", err)
		return nil, 0, err
	}
	return PingIPv4(dst)
}

func PingIPv4(dst *net.IPAddr) (*net.IPAddr, time.Duration, error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		fmt.Println("[PING] listen", err)
		return nil, 0, err
	}
	defer c.Close()

	// Make a new ICMP message
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1, //<< uint(seq), // TODO
			Data: []byte(""),
		},
	}
	b, err := m.Marshal(nil)
	if err != nil {
		return dst, 0, err
	}

	// Send it
	start := time.Now()
	n, err := c.WriteTo(b, dst)
	if err != nil {
		return dst, 0, err
	} else if n != len(b) {
		fmt.Println("[PING] packet size", dst.String())
		return dst, 0, fmt.Errorf("got %v; want %v", n, len(b))
	}

	// Wait for a reply
	reply := make([]byte, 1500)
	err = c.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	if err != nil {
		fmt.Println("[PING] deadline", err)
		return dst, 0, err
	}
	n, _, err = c.ReadFrom(reply)
	if err != nil {
		fmt.Println("[PING] read", dst.String(), err)
		return dst, 0, err
	}

	duration := time.Since(start)
	fmt.Println("[PING] success", dst.String(), duration)

	rm, err := icmp.ParseMessage(1, reply[:n])
	if err != nil {
		return dst, 0, err
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		return dst, duration, nil
	default:
		return dst, 0, fmt.Errorf("[PING] wrong response type %+v", rm)
	}
}
