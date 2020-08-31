package main

import (
	"net"

	"github.com/zhanglongx/Transit/transit"
)

func main() {

	transit := transit.Transit{
		IPArray: [2]net.IP{
			net.IPv4(11, 11, 11, 104),
			net.IPv4(11, 11, 11, 106),
		},

		ThirdPartyAddr: "11.11.11.104:8001",

		IP: net.IPv4(11, 11, 11, 109),

		Port: 7001,
	}

	if err := transit.Open(); err != nil {
		panic(err)
	}

	go transit.Transit()

	for {

	}
}
