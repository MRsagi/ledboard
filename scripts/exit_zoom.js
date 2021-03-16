var sh = new ActiveXObject("WScript.Shell")

if (sh.AppActivate("Zoom Meeting")){
    sh.SendKeys("%q");
    WScript.Sleep(50);
    sh.SendKeys("{TAB}");
    WScript.Sleep(50);
    sh.SendKeys("{ENTER}");
    WScript.Quit(0);
}