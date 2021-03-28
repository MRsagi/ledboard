package controllers

type ButtonConfig struct {
	Cmd string `json:"cmd"`
}

func (b *ButtonConfig) Init(name uint8, triggerCh ...chan bool) chan bool {
	activeCh := make(chan bool, 1)
	if len(triggerCh) > 0 {
		go runBtnWithNotify(name, b.Cmd, activeCh, triggerCh[0])
	} else {
		go runBtn(name, b.Cmd, activeCh)
	}
	return activeCh
}

func runBtnWithNotify(name uint8, cmd string, activeCh chan bool, triggerCh chan bool) {
	globalLog.Debugf("successfully init btn:%v cmd:%v with notification", name, cmd)
	for {
		<-activeCh
		globalLog.Debugf("Btn:%v running cmd:%v", name, cmd)
		go func() {
			err := runOsCommand(cmd)
			globalLog.CheckError(err)
		}()
		triggerCh <- true
	}
}

func runBtn(name uint8, cmd string, activeCh chan bool) {
	globalLog.Debugf("successfully init btn:%v cmd:%v", name, cmd)
	for {
		<-activeCh
		globalLog.Debugf("Btn:%v running cmd:%v", name, cmd)
		go func() {
			err := runOsCommand(cmd)
			globalLog.CheckError(err)
		}()
	}
}
