package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
)

func abort(err error) {
	if err != nil {
		panic(err)
	}
}

const IPV4_SERVER = "http://ipv4.myexternalip.com/raw"
const IPV6_SERVER = "http://ipv6.myexternalip.com/raw"

func findExternalAddress(wg *sync.WaitGroup, c chan string, version int) {
	var url string
	if version == 4 {
		url = IPV4_SERVER
	} else {
		url = IPV6_SERVER
	}
	out := "Could not find external address"

	if resp, err := http.Get(url); err == nil {
		defer resp.Body.Close()
		if host, err := ioutil.ReadAll(resp.Body); err == nil {
			out = fmt.Sprintf("External Address (IPv%d): %s", version, host)
		}
	}

	c <- out

	wg.Done()
}

// return address string, is ip4, is addressable address
func processAddr(addr net.Addr) (string, bool, bool) {
	if n, ok := addr.(*net.IPNet); ok {
		ip := n.IP
		return ip.String(), ip.DefaultMask() != nil, ip.IsGlobalUnicast()
	} else {
		return "", false, false
	}
}

func main() {
	var (
		name     string
		version  int
		external bool
	)

	flag.IntVar(&version, "t", 0, "ipv4 or ipv6")
	flag.StringVar(&name, "n", "", "named interface")
	flag.BoolVar(&external, "e", false, "find external IP")
	flag.Parse()

	c := make(chan string, 2)
	var wg sync.WaitGroup

	if external {
		if version != 6 {
			wg.Add(1)
			go findExternalAddress(&wg, c, 4)
		}
		if version != 4 {
			wg.Add(1)
			go findExternalAddress(&wg, c, 6)
		}
	}

	ifaces, e := net.Interfaces()
	abort(e)

	if external {
		fmt.Println("== Local Addresses ==")
	}

	for _, iface := range ifaces {

		if name != "" && iface.Name != name {
			continue
		}

		addrs, e := iface.Addrs()
		abort(e)

		for _, addr := range addrs {
			straddr, ip4, addressable := processAddr(addr)

			if !addressable {
				continue
			}
			if !ip4 && version == 4 {
				continue
			}
			if ip4 && version == 6 {
				continue
			}

			if name == "" {
				fmt.Printf("%s: ", iface.Name)
			}
			fmt.Println(straddr)
		}
	}

	if external {
		fmt.Println("== External Addresses ==")
		wg.Wait()
		for i := 0; i <= len(c); i++ {
			fmt.Print(<-c)
		}
	}
}
