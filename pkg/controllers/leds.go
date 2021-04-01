package controllers

import (
	"time"

	"github.com/ledboard/pkg/bus"
)

type LedActiveType string

const (
	LedToggleType LedActiveType = "toggle"
	LedCmdType    LedActiveType = "cmd"
	LedNoneType   LedActiveType = "none"
)

type LedConfig struct {
	LedType      LedActiveType   `json:"type"`
	ToggleConfig LedToggleConfig `json:"toggle"`
	CmdConfig    LedCmdConfig    `json:"ledCmd"`
}

//inits led goroutine and returns the button number if such should be matched
func (c *LedConfig) init(writeCh chan bus.LedLight, ledNum uint8) (uint8, chan bool) {
	switch c.LedType {
	case "toggle":
		globalLog.Infof("led:%v type:%v init:%v", ledNum, c.LedType, c.ToggleConfig.InitOn)
		ledToggleCh := make(chan bool, 1)
		go c.ToggleConfig.run(writeCh, ledToggleCh, ledNum)
		return c.ToggleConfig.ButtonNum, ledToggleCh
	case "cmd":
		globalLog.Infof("led:%v type:%v cmd:%v every[sec]:%v", ledNum, c.LedType, c.CmdConfig.Cmd, c.CmdConfig.Sec)
		//I don't know why but running this func as method causes errors
		go runCmd(writeCh, c.CmdConfig.Cmd, c.CmdConfig.Sec, c.CmdConfig.Blink, ledNum)
		return 0, nil
	}
	return 0, nil
}

type LedToggleConfig struct {
	InitOn    bool  `json:"init"`
	ButtonNum uint8 `json:"button"`
}

func (c *LedToggleConfig) run(writeCh chan bus.LedLight, ledToggleCh chan bool, ledNum uint8) {
	writeCh <- bus.LedLight{ledNum, c.InitOn, false}
	state := c.InitOn
	for {
		<-ledToggleCh
		writeCh <- bus.LedLight{ledNum, !state, false}
		state = !state
	}
}

type LedCmdConfig struct {
	Sec   int    `json:"sec"`
	Cmd   string `json:"cmd"`
	Blink bool   `json:"blink"`
	state bool
}

func runCmd(writeCh chan bus.LedLight, cmd string, sec int, blink bool, ledName uint8) {
	var cmdOk bool
	state := false
	waitTime := time.Tick(time.Duration(sec) * time.Second)
	for {
		<-waitTime
		err := runOsCommand(cmd)
		if err != nil {
			cmdOk = false
		} else {
			cmdOk = true
		}
		if cmdOk != state {
			writeCh <- bus.LedLight{ledName, cmdOk, blink}
			state = cmdOk
		}
		globalLog.Debugf("LED:%v cmd:%v", ledName, cmdOk)
	}
}
