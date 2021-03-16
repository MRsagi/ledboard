var sh = new ActiveXObject("WScript.Shell")

if (sh.AppActivate("Zoom Meeting")){
    sh.SendKeys("%a"); 
    WScript.Quit(0);
}