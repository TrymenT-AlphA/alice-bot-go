@rm -r ./build>nul 2>nul
@cd src
@go build -o "../build/bot-go.exe"
@cd ../build
@bot-go.exe