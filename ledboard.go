package main

import (
	"github.com/ledboard/pkg/bus"
	c "github.com/ledboard/pkg/controllers"
	e "github.com/ledboard/pkg/errors"
)

const defaultConfig = "conf.json"

func main() {
	conf := c.NewConfig(defaultConfig)
	bus := bus.NewSerialBus(conf.LedBoard.Port, conf.LedBoard.Baud, conf.Log)
	bus.Connect()
	defer bus.Disconnect()
	run(conf, bus)
}

func run(conf c.UserConfig, serialBus bus.SerialBus) {
	//setup
	stopCh := e.SetupCloseHandler()
	activeChans := conf.MakeControllers(serialBus)

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
