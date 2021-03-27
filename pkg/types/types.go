package types

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ledboard/pkg/bus"
	e "github.com/ledboard/pkg/errors"
)

var globalLog *e.Logger

type ButtonActiveConfig interface {
	Run(chan bus.LedLight, chan bool, uint8, string)
}

type UserConfig struct {
	LedBoard LedBoardConfig         `json:"ledBoard"`
	Buttons  map[uint8]ButtonConfig `json:"buttons"`
	Log      *e.Logger
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

func NewConfig(filename string) UserConfig {
	globalLog = e.NewLogger()
	file, err := ioutil.ReadFile(filename)
	globalLog.CheckPanic(err)
	var conf UserConfig
	err = json.Unmarshal(file, &conf)
	globalLog.CheckPanic(err)
	conf.Log = globalLog
	conf.Log.Debugf("%v", conf)
	return conf
}

func (b *ButtonConfig) Run(writeCh chan bus.LedLight, activate chan bool, buttonName uint8, cmd string) {
	b.callback.Run(writeCh, activate, buttonName, cmd)
}

func (b *ButtonConfig) InitCallback() {
	switch b.LedType {
	case "toggle":
		globalLog.Infof("type:%v init:%v", b.LedType, b.LedToggleConfig.InitOn)
		b.callback = &b.LedToggleConfig
	case "cmd":
		globalLog.Infof("type:%v cmd:%v every[sec]:%v", b.LedType, b.LedCmdConfig.Cmd, b.LedCmdConfig.Sec)
		b.callback = &b.LedCmdConfig
	case "toggleIfCmd":
		globalLog.Infof("type:%v cmd:%v every[sec]:%v init:%v", b.LedType, b.LedCmdConfig.Cmd, b.LedCmdConfig.Sec, b.LedToggleConfig.InitOn)
		b.callback = &LedToggleIfCmdConfig{
			b.LedCmdConfig,
			b.LedToggleConfig,
		}
	default:
		globalLog.Infof("type:%v, default", b.LedType)
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
	var okCh = make(chan bool, 1)
	waitingForCmd := true
	waitTime := time.Tick(time.Duration(cc.Sec) * time.Second)
	go func() {
		err := runOsCommand(cc.Cmd)
		if err != nil {
			okCh <- false
		}
		okCh <- true
	}()
	for {
		if !waitingForCmd {
			select {
			case <-waitTime:
				go func() {
					err := runOsCommand(cc.Cmd)
					if err != nil {
						okCh <- false
					}
					okCh <- true
				}()
			default:
			}
		}
		select {
		case <-activate:
			err := runOsCommand(cmd)
			globalLog.CheckError(err)
		case ok := <-okCh:
			waitingForCmd = false
			globalLog.Debugf("Btn:%v cmd:%v", buttonName, ok)
			switch {
			case !ok && cc.state == true:
				writeCh <- bus.LedLight{buttonName, false}
				cc.state = false
			case ok && cc.state == false:
				writeCh <- bus.LedLight{buttonName, true}
				cc.state = true
			}
		default:
		}
	}
}

type LedToggleConfig struct {
	InitOn bool `json:"init"`
}

func (tg *LedToggleConfig) Run(writeCh chan bus.LedLight, activate chan bool, buttonName uint8, cmd string) {
	state := tg.InitOn
	globalLog.Infof("Running toggle go of:%v to:%v", buttonName, tg.InitOn)
	writeCh <- bus.LedLight{buttonName, state}
	for {
		state = !state
		<-activate
		err := runOsCommand(cmd)
		globalLog.CheckError(err)
		writeCh <- bus.LedLight{buttonName, state}
	}
}

type LedToggleIfCmdConfig struct {
	LedCmdConfig
	LedToggleConfig
}

func (tic *LedToggleIfCmdConfig) Run(writeCh chan bus.LedLight, activate chan bool, buttonName uint8, cmd string) {
	okCh := make(chan bool, 1)
	prevCmdOk := false
	waitingForCmd := true
	waitTime := time.Tick(time.Duration(tic.Sec) * time.Second)
	go func() {
		err := runOsCommand(tic.Cmd)
		if err != nil {
			okCh <- false
		}
		okCh <- true
	}()
	for {
		if !waitingForCmd {
			select {
			case <-waitTime:
				go func() {
					err := runOsCommand(tic.Cmd)
					if err != nil {
						okCh <- false
					}
					okCh <- true
				}()
			default:
			}
		}
		select {
		case <-activate:
			err := runOsCommand(cmd)
			globalLog.CheckError(err)
			if prevCmdOk {
				writeCh <- bus.LedLight{buttonName, !tic.state}
				tic.state = !tic.state
			}
		case ok := <-okCh:
			waitingForCmd = false
			globalLog.Debugf("Btn:%v cmd:%v", buttonName, ok)
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
		default:
		}
	}
}

func runOsCommand(cmd string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "windows":
		name = "cmd"
		args = []string{"\\c", cmd}
	case "linux":
		name = "/bin/sh"
		args = []string{"-c", cmd}
	default:
		splitCmd := strings.Split(cmd, " ")
		name = splitCmd[0]
		args = splitCmd[1:]
	}
	globalLog.Debugf("running cmd: %v %v", name, args)
	err := exec.Command(name, args...).Run()
	return err
}
