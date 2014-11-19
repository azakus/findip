package main

import (
	"flag"
	"fmt"
	"net"
)

func processAddr(addr *net.Addr) (string, bool) {
	net, ok := (*addr).(*net.IPNet)
	if !ok {
		return "", false
	}
	ip := net.IP
	return ip.String(), ip.DefaultMask() != nil
}

func main() {
	var (
		name string
		typ  int
	)

	flag.IntVar(&typ, "t", 0, "ipv4 or ipv6")
	flag.StringVar(&name, "n", "", "named interface")
	flag.Parse()

	ifaces, _ := net.Interfaces()

	for _, iface := range ifaces {

		if name != "" && iface.Name != name {
			continue
		}

		addrs, _ := iface.Addrs()

		for _, addr := range addrs {
			straddr, ip4 := processAddr(&addr)

			if !ip4 && typ == 4 {
				continue
			}
			if ip4 && typ == 6 {
				continue
			}

			if name == "" {
				fmt.Printf("%s: ", iface.Name)
			}
			fmt.Println(straddr)
		}
	}
}
