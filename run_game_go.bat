deploy.py
if ERRORLEVEL 1 EXIT /B 1
REM halite.exe -s 1437533101 -t -d "360 240" "go run MyBot.go"  "go run versions\57\MyBot.go" "go run versions\57\MyBot.go"  "go run versions\57\MyBot.go"
halite.exe "MyBot.exe" "MyBotAlly2.exe" "MyBotNoAlly.exe" "MyBotNoAlly.exe"
REM halite.exe -d "120 120" "MyBot.exe" "go run versions\81\MyBot.go"
