package monopoly

import (
	"os"
	"time"
)

type Logger interface {
	Log(message string)
	LogWithState(message string, state GameState)
	LogState(state GameState)
	Error(message string, state GameState)
}

type ConsoleLogger struct {
	StateID int64
}

func (c *ConsoleLogger) Init() {
	c.StateID = 1
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

func (c *ConsoleLogger) LogState(state GameState) {}

func (c *ConsoleLogger) LogWithState(message string, state GameState) {
	c.Log(message)
	c.LogState(state)
}

func (c *ConsoleLogger) Error(message string, state GameState) {
	newMsg := "!!!!!!!!!! ERROR: " + message
	c.LogWithState(newMsg, state)
}
