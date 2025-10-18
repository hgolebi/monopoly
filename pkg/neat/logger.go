package neatnetwork

import (
	"fmt"
	"monopoly/pkg/monopoly"
	"os"
	"path/filepath"
)

type TrainerLogger struct {
	disabled   bool
	outputPath string
	stateId    int64
}

func NewTrainerLogger(outputPath string, disabled bool) (*TrainerLogger, error) {
	if _, err := os.Stat(outputPath); err == nil {
		if err := os.Remove(outputPath); err != nil {
			return nil, fmt.Errorf("failed to remove existing file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if !disabled {
		if err := createFile(outputPath); err != nil {
			return nil, fmt.Errorf("failed to create log file: %w", err)
		}
	}
	return &TrainerLogger{
		outputPath: outputPath,
		stateId:    0,
		disabled:   disabled,
	}, nil
}

func createFile(outputPath string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

func (l *TrainerLogger) log(message string) {
	file, err := os.OpenFile(l.outputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
	}
	defer file.Close()

	if _, err := file.WriteString(message + "\n"); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}

func (l *TrainerLogger) Log(message string) {
	if l.disabled {
		return
	}
	l.log(message)
}

func (l *TrainerLogger) Error(message string, state monopoly.GameState) {
	if l.disabled {
		err := createFile(l.outputPath)
		if err != nil {
			fmt.Printf("Error creating log file: %v\n", err)
		}
	}
	newMsg := "!!!!!!!!!! ERROR: " + message
	l.log(newMsg)
	l.LogState(state)
}

func (l *TrainerLogger) LogState(state monopoly.GameState) {
	f, err := os.OpenFile(l.outputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		println("Error opening file:", err.Error())
		return
	}
	defer f.Close()
	log := fmt.Sprintf("%s", state.String())
	_, err = f.WriteString(log)

	if err != nil {
		println("Error writing to file:", err.Error())
		return
	}
	l.stateId++
}

func (l *TrainerLogger) LogWithState(message string, state monopoly.GameState) {
	if l.disabled {
		return
	}
	l.log(message)
	l.LogState(state)
}
