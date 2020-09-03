package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zhanglongx/Transit/transit"
)

func main() {

	confFile := flag.String("f", "/usr/local/etc/transit.json", "configure file name")

	flag.Parse()

	if _, err := os.Stat(*confFile); os.IsNotExist(err) {
		fmt.Printf("%s not exists\n", *confFile)
		os.Exit(1)
	}

	buf, err := ioutil.ReadFile(*confFile)
	if err != nil {
		fmt.Printf("read %s failed\n", *confFile)
	}

	var transit transit.Transit
	if err := json.Unmarshal(buf, &transit); err != nil {
		fmt.Printf("parse config file:%s failed\n", *confFile)
	}

	if err := transit.Open(); err != nil {
		panic(err)
	}

	transit.Transit()
}
