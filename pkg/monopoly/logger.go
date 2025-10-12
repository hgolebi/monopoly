package monopoly

import (
	"fmt"
	"os"
)

type Logger interface {
	Log(message string)
	LogWithState(message string, state GameState)
	LogState(state GameState)
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
	println(message)
}

func (c *ConsoleLogger) LogState(state GameState) {
	f, err := os.OpenFile("state_log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		println("Error opening file:", err.Error())
		return
	}
	defer f.Close()
	log := "#" + fmt.Sprint(c.StateID) + "\n" + state.String()
	_, err = f.WriteString(log)

	if err != nil {
		println("Error writing to file:", err.Error())
		return
	}
	c.StateID++
	fmt.Printf("#%d\n", c.StateID)
}

func (c *ConsoleLogger) LogWithState(message string, state GameState) {
	c.Log(message)
	c.LogState(state)
}
