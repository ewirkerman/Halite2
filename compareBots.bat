go build MyBot.go
if ERRORLEVEL 1 EXIT /B 1
"..\..\git-projects\halite2_reload\halite2 reload.py" replay.json 1 "MyBot.exe"