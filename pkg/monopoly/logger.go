package monopoly

import "time"

type Logger interface {
	Log(message string)
}

type ConsoleLogger struct{}

func (c *ConsoleLogger) Log(message string) {
	time.Sleep(500 * time.Millisecond)
	println(message)
}
