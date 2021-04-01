package main

import (
	"flag"

	"github.com/ledboard/pkg/bus"
	c "github.com/ledboard/pkg/controllers"
	e "github.com/ledboard/pkg/errors"
)

const defaultConfig = "conf.json"

func main() {
	debugFlag := flag.Bool("debug", false, "debug logging")
	confFile := flag.String("conf", defaultConfig, "override default config file")
	flag.Parse()

	conf := c.NewConfig(*confFile, *debugFlag)
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
			conf.Log.Infof("\r- Sutting down")
			return
		}
	}
}
