package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/ledboard/pkg/bus"
	e "github.com/ledboard/pkg/errors"
	"github.com/ledboard/pkg/types"
)

const defaultConfig = "conf.json"

func main() {
	conf := newConfig(defaultConfig)
	bus := bus.NewSerialBus(conf.LedBoard.Port, conf.LedBoard.Baud)
	bus.Connect()
	run(conf, bus)
}

func newConfig(filename string) types.UserConfig {
	file, err := ioutil.ReadFile(filename)
	e.CheckPanic(err)
	var conf types.UserConfig
	err = json.Unmarshal(file, &conf)
	e.CheckPanic(err)
	fmt.Printf("%v\n", conf)
	return conf
}

func run(conf types.UserConfig, serialBus bus.SerialBus) {
	var activeChans = make(map[uint8]chan bool)

	//setup
	for i, button := range conf.Buttons {
		fmt.Printf("Init button:%v ", i)
		button.InitCallback()
		activeChans[i] = make(chan bool, 1)
		go button.Run(serialBus.GetWriteChannel(), activeChans[i], i, button.Cmd)
		time.Sleep(50 * time.Millisecond)
	}

	//loop
	var btn uint8
	readCh := serialBus.GetReadChannel()
	for {
		btn = <-readCh
		fmt.Printf("Pushed:%v\n", btn)
		ch, ok := activeChans[uint8(btn)]
		if !ok {
			fmt.Printf("Error activating %v\n", btn)
			continue
		}
		ch <- true
	}
}
