package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

const asciiByteOffset = 48

type ButtonActiveConfig interface {
	Run(chan uint8, chan bool, uint8, string)
}

type UserConfig struct {
	LedBoard LedBoardConfig         `json:"ledBoard"`
	Buttons  map[uint8]ButtonConfig `json:"buttons"`
}

type LedBoardConfig struct {
	Port   string `json:"port"`
	Baud   uint   `json:"baud"`
	Keys   int    `json:"keys"`
	serial serialBus
}

type serialBus struct {
	bus io.ReadWriteCloser
}

func (s *serialBus) RunRead(readCh chan uint8) uint8 {
	for {
		buf := make([]byte, 1)
		s.bus.Read(buf)
		readCh <- buf[0]
	}
}

func (s *serialBus) RunWrite(writeCh chan uint8) {
	var b uint8
	for {
		b = <-writeCh
		sendValue := b + asciiByteOffset
		n, err := s.bus.Write([]byte{sendValue})
		fmt.Printf("sent:%v, len:%v\n", sendValue, n)
		if err != nil {
			fmt.Printf("err:%v\n", err.Error())
		}
	}
}

type ButtonConfig struct {
	Cmd                  string          `json:"cmd"`
	LedType              string          `json:"type"`
	LedToggleConfig      LedToggleConfig `json:"toggle"`
	LedCmdConfig         LedCmdConfig    `json:"ledCmd"`
	LedToggleIfCmdConfig LedToggleIfCmdConfig
	ByteValue            uint8 `json:"byte"`
	callback             ButtonActiveConfig
	activeCh             chan bool
}

func (b *ButtonConfig) initCallback() {
	fmt.Printf("type:%v ", b.LedType)
	switch b.LedType {
	case "toggle":
		fmt.Printf("init:%v\n", b.LedToggleConfig.InitOn)
		b.callback = &b.LedToggleConfig
	case "cmd":
		fmt.Printf("cmd:%v every[sec]:%v\n", b.LedCmdConfig.Cmd, b.LedCmdConfig.Sec)
		b.callback = &b.LedCmdConfig
	case "toggleIfCmd":
		fmt.Printf("cmd:%v every[sec]:%v init:%v\n", b.LedCmdConfig.Cmd, b.LedCmdConfig.Sec, b.LedToggleConfig.InitOn)
		b.callback = &LedToggleIfCmdConfig{
			b.LedCmdConfig,
			b.LedToggleConfig,
		}
	default:
		fmt.Printf("default\n")
		b.LedToggleConfig = LedToggleConfig{
			InitOn: false,
		}
		b.callback = &b.LedToggleConfig
	}
}

type LedCmdConfig struct {
	Sec   int    `json:"sec"`
	Cmd   string `json:"cmd"`
	state bool
}

func (cc *LedCmdConfig) Run(writeCh chan uint8, activate chan bool, buttonName uint8, cmd string) {
	runArgs := strings.Split(cmd, " ")
	toggleArgs := strings.Split(cc.Cmd, " ")
	var okCh = make(chan bool, 1)
	waitingForCmd := true
	waitTime := time.Tick(time.Duration(cc.Sec) * time.Second)
	go checkCommand(okCh, toggleArgs[0], toggleArgs[1:])
	for {
		if !waitingForCmd {
			select {
			case <-waitTime:
				go checkCommand(okCh, toggleArgs[0], toggleArgs[1:])
			default:
			}
		}
		select {
		case <-activate:
			err := exec.Command(runArgs[0], runArgs[1:]...).Run()
			checkError(err)
		case ok := <-okCh:
			waitingForCmd = false
			fmt.Printf("Btn:%v cmd:%v\n", buttonName, ok)
			switch {
			case !ok && cc.state == true:
				writeCh <- buttonName
				cc.state = false
			case ok && cc.state == false:
				writeCh <- buttonName
				cc.state = true
			}
		}
	}
}

type LedToggleConfig struct {
	InitOn bool `json:"init"`
}

func (tg *LedToggleConfig) Run(writeCh chan uint8, activate chan bool, buttonName uint8, cmd string) {
	fmt.Printf("Running toggle go of: %v\n", buttonName)
	if tg.InitOn {
		writeCh <- buttonName
	}
	runArgs := strings.Split(cmd, " ")
	for {
		<-activate
		fmt.Printf("running cmd: %v ", cmd)
		err := exec.Command(runArgs[0], runArgs[1:]...).Run()
		checkError(err)
		writeCh <- buttonName
	}
}

type LedToggleIfCmdConfig struct {
	LedCmdConfig
	LedToggleConfig
}

func (tic *LedToggleIfCmdConfig) Run(writeCh chan uint8, activate chan bool, buttonName uint8, cmd string) {
	runArgs := strings.Split(cmd, " ")
	toggleArgs := strings.Split(tic.Cmd, " ")
	okCh := make(chan bool, 1)
	prevCmdOk := false
	waitingForCmd := true
	waitTime := time.Tick(time.Duration(tic.Sec) * time.Second)
	go checkCommand(okCh, toggleArgs[0], toggleArgs[1:])
	for {
		if !waitingForCmd {
			select {
			case <-waitTime:
				go checkCommand(okCh, toggleArgs[0], toggleArgs[1:])
			default:
			}
		}
		select {
		case <-activate:
			err := exec.Command(runArgs[0], runArgs[1:]...).Run()
			checkError(err)
			if prevCmdOk {
				writeCh <- buttonName
				tic.state = !tic.state
			}
		case ok := <-okCh:
			waitingForCmd = false
			fmt.Printf("Btn:%v cmd:%v\n", buttonName, ok)
			switch {
			case ok && prevCmdOk == false && tic.InitOn:
				writeCh <- buttonName
				tic.state = !tic.state
				fallthrough
			case ok && prevCmdOk == false:
				prevCmdOk = true
			//always turn off when cmd false
			case !ok && prevCmdOk == true && tic.state == true:
				writeCh <- buttonName
				tic.state = !tic.state
				fallthrough
			case !ok && prevCmdOk == true:
				prevCmdOk = false
			}
		}
	}
}

func checkCommand(ok chan bool, name string, args []string) {
	err := exec.Command(name, args...).Run()
	if err != nil {
		ok <- false
		return
	}
	ok <- true
}
