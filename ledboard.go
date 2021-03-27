package main

import (
	"time"

	"github.com/ledboard/pkg/bus"
	e "github.com/ledboard/pkg/errors"
	"github.com/ledboard/pkg/types"
)

const defaultConfig = "conf.json"

func main() {
	conf := types.NewConfig(defaultConfig)
	bus := bus.NewSerialBus(conf.LedBoard.Port, conf.LedBoard.Baud, conf.Log)
	bus.Connect()
	defer bus.Disconnect()
	run(conf, bus)
}

func run(conf types.UserConfig, serialBus bus.SerialBus) {
	var activeChans = make(map[uint8]chan bool)
	//setup
	stopCh := e.SetupCloseHandler()
	for i, button := range conf.Buttons {
		conf.Log.Debugf("Init button:%v", i)
		button.InitCallback()
		activeChans[i] = make(chan bool, 1)
		go button.Run(serialBus.GetWriteChannel(), activeChans[i], i, button.Cmd)
		time.Sleep(100 * time.Millisecond)
	}

	//loop
	var btn uint8
	readCh := serialBus.GetReadChannel()
	for {
		select {
		case btn = <-readCh:
			conf.Log.Debugf("Pushed:%v", btn)
			ch, ok := activeChans[uint8(btn)]
			if !ok {
				conf.Log.Debugf("Error activating %v", btn)
				continue
			}
			ch <- true
		case <-stopCh:
			conf.Log.Infof("\r- Ctrl+C pressed in Terminal")
			return
		}
	}
}
