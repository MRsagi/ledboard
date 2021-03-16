package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

const defaultConfig = "conf.json"

func main() {
	conf := newConfig(defaultConfig)
	conf.LedBoard.connect()
	conf.run()
}

func checkPanic(err error) {
	if err != nil {
		panic(err.Error())
	}
}
func checkError(err error) {
	if err != nil {
		fmt.Printf("ERROR: %v\n", err.Error())
	}
}
func newConfig(filename string) UserConfig {
	file, err := ioutil.ReadFile(filename)
	checkPanic(err)
	var conf UserConfig
	err = json.Unmarshal(file, &conf)
	checkPanic(err)
	fmt.Printf("%v\n", conf)
	return conf
}

func (lbConf *LedBoardConfig) connect() {
	options := serial.OpenOptions{
		PortName: lbConf.Port,
		BaudRate: lbConf.Baud,
		DataBits: 8,
	}
	port, err := serial.Open(options)
	checkPanic(err)
	buf := make([]byte, 1)
	_, err = port.Read(buf)
	checkPanic(err)
	if string(buf[0]) != "s" {
		panic("error reading from ledboard")
	}
	_, err = port.Write([]byte{0})
	checkPanic(err)
	time.Sleep(time.Second)
	fmt.Println("connection successful")
	lbConf.serial = serialBus{bus: port}
}
func (conf *UserConfig) run() {
	var activeChans = make(map[uint8]chan bool)
	var readCh = make(chan uint8, 256)
	var writeCh = make(chan uint8, 256)
	//setup
	for i, button := range conf.Buttons {
		fmt.Printf("Init button:%v ", i)
		button.initCallback()
		activeChans[i] = make(chan bool, 1)
		go button.callback.Run(writeCh, activeChans[i], i, button.Cmd)
		time.Sleep(250 * time.Millisecond)
	}
	go conf.LedBoard.serial.RunRead(readCh)
	go conf.LedBoard.serial.RunWrite(writeCh)
	//loop
	var btn uint8
	for {
		btn = <-readCh
		fmt.Printf("Pushed: %v ", btn)
		ch, ok := activeChans[uint8(btn)]
		if !ok {
			fmt.Printf("Error activating %v\n", btn)
			continue
		}
		ch <- true
	}
}
