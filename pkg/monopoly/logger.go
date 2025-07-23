package monopoly

import (
	"os"
	"time"
)

type Logger interface {
	Init()
	Log(message string)
	LogState(state GameState)
}

type ConsoleLogger struct{}

func (c *ConsoleLogger) Init() {
	f, err := os.Create("state_log.txt")
	if err != nil {
		println("Error creating file:", err.Error())
		return
	}
	f.Close()
}

func (c *ConsoleLogger) Log(message string) {
	time.Sleep(500 * time.Millisecond)
	println(message)
}

func (c *ConsoleLogger) LogState(state GameState) {
	f, err := os.OpenFile("state_log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		println("Error opening file:", err.Error())
		return
	}
	defer f.Close()

	_, err = f.WriteString(state.String())
	// _, err = f.WriteString("test\n")

	if err != nil {
		println("Error writing to file:", err.Error())
	}
}
