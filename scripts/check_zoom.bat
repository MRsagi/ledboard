tasklist /v /nh /fi "WINDOWTITLE eq Zoom Meeting" |findstr /B /C:"INFO: No tasks are running">nul && (exit 1) || (exit 0)
