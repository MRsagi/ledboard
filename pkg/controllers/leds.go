package controllers

import (
	"time"

	"github.com/ledboard/pkg/bus"
)

type LedConfig struct {
	LedType         string          `json:"type"`
	LedToggleConfig LedToggleConfig `json:"toggle"`
	LedCmdConfig    LedCmdConfig    `json:"ledCmd"`
}

//inits led goroutine and returns the button number if such should be matched
func (c *LedConfig) init(writeCh chan bus.LedLight, ledNum uint8) (uint8, chan bool) {
	switch c.LedType {
	case "toggle":
		globalLog.Infof("type:%v init:%v", c.LedType, c.LedToggleConfig.InitOn)
		ledToggleCh := make(chan bool, 1)
		go c.LedToggleConfig.run(writeCh, ledToggleCh, ledNum)
		return c.LedToggleConfig.ButtonNum, ledToggleCh
	case "cmd":
		globalLog.Infof("type:%v cmd:%v every[sec]:%v", c.LedType, c.LedCmdConfig.Cmd, c.LedCmdConfig.Sec)
		go run(writeCh, c.LedCmdConfig.Cmd, c.LedCmdConfig.Sec, ledNum)
		return 0, nil
	}
	return 0, nil
}

type LedToggleConfig struct {
	InitOn    bool  `json:"init"`
	ButtonNum uint8 `json:"button"`
}

func (c *LedToggleConfig) run(writeCh chan bus.LedLight, ledToggleCh chan bool, ledNum uint8) {
	writeCh <- bus.LedLight{ledNum, c.InitOn}
	state := c.InitOn
	for {
		<-ledToggleCh
		writeCh <- bus.LedLight{ledNum, !state}
		state = !state
	}
}

type LedCmdConfig struct {
	Sec   int    `json:"sec"`
	Cmd   string `json:"cmd"`
	state bool
}

func run(writeCh chan bus.LedLight, cmd string, sec int, ledName uint8) {
	var cmdOk bool
	waitTime := time.Tick(time.Duration(sec) * time.Second)
	for {
		<-waitTime
		err := runOsCommand(cmd)
		if err != nil {
			cmdOk = false
		} else {
			cmdOk = true
		}
		writeCh <- bus.LedLight{ledName, cmdOk}
		globalLog.Debugf("LED:%v cmd:%v", ledName, cmdOk)
	}
}
