var sh = new ActiveXObject("WScript.Shell")

if (sh.AppActivate("Zoom Meeting")){
    sh.SendKeys("%v"); 
    WScript.Quit(0);
}