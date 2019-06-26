package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var (
	Host string
)

func Flags() {
	flag.StringVar(&Host, "host", "redhat.cz", "a valid host to start traceroute")
	defer flag.Parse()
	return
}

// Sample traceroute proof of concept for RedHat Interview.
func main() {

	Flags()
	ips, err := net.LookupIP(Host)
	if err != nil {
		log.Fatal(err)
	}
	var dst net.IPAddr
	for _, ip := range ips {
		dst.IP = ip
		fmt.Printf("Tracing IP packet route to %v - %s\n", dst.IP, Host)
		break
	}

	c, _ := net.ListenPacket("ip4:1", "0.0.0.0")
	defer c.Close()
	p := ipv4.NewPacketConn(c)

	if err := p.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true); err != nil {
		log.Fatal(err)
	}
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Data: []byte("entrophy"),
		},
	}
	rb := make([]byte, 1500)
	for i := 1; i <= 64; i++ { // predefined hops size
		wm.Body.(*icmp.Echo).Seq = i
		wb, err := wm.Marshal(nil)
		if err != nil {
			log.Fatal(err)
		}
		if err := p.SetTTL(i); err != nil {
			log.Fatal(err)
		}

		begin := time.Now()
		if _, err := p.WriteTo(wb, nil, &dst); err != nil {
			log.Fatal(err)
		}
		if err := p.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
			log.Fatal(err)
		}
		n, cm, peer, err := p.ReadFrom(rb)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				fmt.Printf("%v\t*\n", i)
				continue
			}
			log.Fatal(err)
		}
		rm, _ := icmp.ParseMessage(1, rb[:n])
		fmt.Println(rm)
		rtt := time.Since(begin)

		switch rm.Type {
		case ipv4.ICMPTypeTimeExceeded:
			names, _ := net.LookupAddr(peer.String())
			fmt.Printf("%d\t%v %+v %v\n\t%+v\n", i, peer, names, rtt, cm)
		case ipv4.ICMPTypeEchoReply:
			names, _ := net.LookupAddr(peer.String())
			fmt.Printf("%d\t%v %+v %v\n\t%+v\n", i, peer, names, rtt, cm)
			return
		default:
			log.Printf("unknown ICMP message: %+v\n", rm)
		}
	}
}
