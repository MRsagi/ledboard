package types

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ledboard/pkg/bus"
	e "github.com/ledboard/pkg/errors"
)

type ButtonActiveConfig interface {
	Run(chan bus.LedLight, chan bool, uint8, string)
}

type UserConfig struct {
	LedBoard LedBoardConfig         `json:"ledBoard"`
	Buttons  map[uint8]ButtonConfig `json:"buttons"`
}

type LedBoardConfig struct {
	Port   string `json:"port"`
	Baud   uint   `json:"baud"`
	Keys   int    `json:"keys"`
	serial bus.SerialBus
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

func (b *ButtonConfig) Run(writeCh chan bus.LedLight, activate chan bool, buttonName uint8, cmd string) {
	b.callback.Run(writeCh, activate, buttonName, cmd)
}

func (b *ButtonConfig) InitCallback() {
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

func (cc *LedCmdConfig) Run(writeCh chan bus.LedLight, activate chan bool, buttonName uint8, cmd string) {
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
			e.CheckError(err)
		case ok := <-okCh:
			waitingForCmd = false
			fmt.Printf("Btn:%v cmd:%v\n", buttonName, ok)
			switch {
			case !ok && cc.state == true:
				writeCh <- bus.LedLight{buttonName, false}
				cc.state = false
			case ok && cc.state == false:
				writeCh <- bus.LedLight{buttonName, true}
				cc.state = true
			}
		}
	}
}

type LedToggleConfig struct {
	InitOn bool `json:"init"`
}

func (tg *LedToggleConfig) Run(writeCh chan bus.LedLight, activate chan bool, buttonName uint8, cmd string) {
	fmt.Printf("Running toggle go of: %v\n", buttonName)
	state := tg.InitOn
	writeCh <- bus.LedLight{buttonName, state}
	runArgs := strings.Split(cmd, " ")
	for {
		state = !state
		<-activate
		fmt.Printf("running cmd: %v ", cmd)
		err := exec.Command(runArgs[0], runArgs[1:]...).Run()
		e.CheckError(err)
		writeCh <- bus.LedLight{buttonName, state}
	}
}

type LedToggleIfCmdConfig struct {
	LedCmdConfig
	LedToggleConfig
}

func (tic *LedToggleIfCmdConfig) Run(writeCh chan bus.LedLight, activate chan bool, buttonName uint8, cmd string) {
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
			e.CheckError(err)
			if prevCmdOk {
				writeCh <- bus.LedLight{buttonName, !tic.state}
				tic.state = !tic.state
			}
		case ok := <-okCh:
			waitingForCmd = false
			fmt.Printf("Btn:%v cmd:%v\n", buttonName, ok)
			switch {
			case ok && prevCmdOk == false:
				writeCh <- bus.LedLight{buttonName, tic.InitOn}
				tic.state = tic.InitOn
				prevCmdOk = true
			//always turn off when cmd false
			case !ok && prevCmdOk == true:
				writeCh <- bus.LedLight{buttonName, false}
				tic.state = false
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
