for ($i=1; $i -le 4; $i++) {
    Start-Process powershell -ArgumentList "cd C:\Users\Hubert\Desktop\studia\inzynierka\app; go run main.go --cli"
    Start-Sleep -Seconds 0.5
}