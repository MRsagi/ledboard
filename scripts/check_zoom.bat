tasklist /v /nh /fi "WINDOWTITLE eq Zoom Meeting" |findstr /B /C:"INFO: No tasks are running">nul && (del .\scripts\zoom && exit 1) || ( echo ok > .\scripts\zoom && exit 0)
