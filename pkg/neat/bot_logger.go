package neatnetwork

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	charmlog "github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

type BotDecisionLogger struct {
	logger *charmlog.Logger
	file   *os.File
}

func NewBotDecisionLogger(outputDir string, botName string) (*BotDecisionLogger, error) {
	dir := filepath.Join(outputDir, "decision_log")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create decision log directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.ans", timestamp, botName)
	path := filepath.Join(dir, filename)

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create decision log file: %w", err)
	}

	logger := charmlog.NewWithOptions(file, charmlog.Options{
		Level:           charmlog.DebugLevel,
		ReportTimestamp: false,
	})
	logger.SetColorProfile(termenv.ANSI256)

	return &BotDecisionLogger{
		logger: logger,
		file:   file,
	}, nil
}

func (l *BotDecisionLogger) Close() {
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}

func (l *BotDecisionLogger) Info(msg string, keyvals ...interface{}) {
	l.logger.Info(msg, keyvals...)
}

func (l *BotDecisionLogger) Debug(msg string, keyvals ...interface{}) {
	l.logger.Debug(msg, keyvals...)
}

func formatFloatSlice(fs []float64) string {
	strs := make([]string, len(fs))
	for i, f := range fs {
		strs[i] = fmt.Sprintf("%.2f", f)
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

func boolToYesNo(b bool) string {
	if b {
		return "YES"
	}
	return "NO"
}
