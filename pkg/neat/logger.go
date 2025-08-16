package neatnetwork

import (
	"errors"
	"fmt"
	"monopoly/pkg/monopoly"
	"os"
)

type TrainerLogger struct {
	outputPath string
	stateId    int64
}

func NewTrainerLogger(outputPath string) (*TrainerLogger, error) {
	if _, err := os.Stat(outputPath); err == nil {
		return nil, errors.New("output file already exists")
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return &TrainerLogger{
		outputPath: outputPath,
		stateId:    0,
	}, nil
}

func (l *TrainerLogger) Log(message string) {
	file, err := os.OpenFile(l.outputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
	}
	defer file.Close()

	if _, err := file.WriteString(message + "\n"); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}

func (l *TrainerLogger) Error(message string) {
	newMsg := "!!!!!!!!!! ERROR: " + message
	l.Log(newMsg)
}

func (l *TrainerLogger) LogState(state monopoly.GameState) {
	f, err := os.OpenFile(l.outputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		println("Error opening file:", err.Error())
		return
	}
	defer f.Close()
	log := "#" + fmt.Sprint(l.stateId) + "\n" + state.String()
	_, err = f.WriteString(log)

	if err != nil {
		println("Error writing to file:", err.Error())
		return
	}
	l.stateId++
	fmt.Printf("#%d\n", l.stateId)
}
