package controllers

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

type UserConfig struct {
	LedBoard LedBoardConfig         `json:"ledBoard"`
	Buttons  map[uint8]ButtonConfig `json:"buttons"`
	Leds     map[uint8]LedConfig    `json:"leds"`
	Log      *e.Logger
}

func (c *UserConfig) MakeControllers(serialBus bus.SerialBus) map[uint8]chan bool {
	var relatedButtons = make(map[uint8]chan bool)
	for ledName, ledConf := range c.Leds {
		c.Log.Debugf("Init Led:%v", ledName)
		matchedBtn, triggerCh := ledConf.init(serialBus.GetWriteChannel(), ledName)
		relatedButtons[matchedBtn] = triggerCh
		time.Sleep(100 * time.Millisecond)
	}
	var activeChans = make(map[uint8]chan bool)
	for btnName, button := range c.Buttons {
		c.Log.Debugf("Init button:%v, cmd: %v", btnName, button.Cmd)
		triggerCh, ok := relatedButtons[btnName]
		if ok {
			activeChans[btnName] = button.Init(btnName, triggerCh)
		} else {
			activeChans[btnName] = button.Init(btnName)
		}
	}
	globalLog.Debugf("activeChans: %v", activeChans)
	return activeChans
}

type LedBoardConfig struct {
	Port   string `json:"port"`
	Baud   uint   `json:"baud"`
	serial bus.SerialBus
}

func NewConfig(filename string) UserConfig {
	globalLog = e.NewLogger()
	file, err := ioutil.ReadFile(filename)
	globalLog.CheckPanic(err)
	var conf UserConfig
	err = json.Unmarshal(file, &conf)
	globalLog.CheckPanic(err)
	conf.Log = globalLog
	conf.Log.Debugf("%v", conf.LedBoard)
	conf.Log.Debugf("%v", conf.Leds)
	conf.Log.Debugf("%v", conf.Buttons)
	return conf
}

func runOsCommand(cmd string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	/*
		case "windows":
			name = "cmd"
			args = []string{"\\c", cmd}
	*/
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
