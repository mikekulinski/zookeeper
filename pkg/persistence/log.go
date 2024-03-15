package persistence

import (
	"fmt"
	"os"
)

const (
	LogFilePrefix      = "log"
	SnapshotFilePrefix = "snapshot"
)

type LogManager struct {
	logDir *os.File
	// TODO: Update to zxid
	lastZxid int64
}

func NewLogManager(filePath string) (*LogManager, error) {
	logDir, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return &LogManager{
		logDir:   logDir,
		lastZxid: 0,
	}, nil
}

func (l *LogManager) Close() error {
	err := l.logDir.Close()
	if err != nil {
		return fmt.Errorf("error closing log directory file: %w", err)
	}
	return nil
}
