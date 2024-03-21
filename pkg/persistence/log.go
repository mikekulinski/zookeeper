package persistence

import (
	"fmt"
	"os"
	"strings"
	"sync"

	pbzk "github.com/mikekulinski/zookeeper/proto"
	"google.golang.org/protobuf/proto"
)

const (
	LogFilePrefix      = "log"
	SnapshotFilePrefix = "snapshot"
)

// LogManager is a Write-Ahead Log (WAL) for our in memory database. We model this as a new
// file for each transaction being written to our log. Each file is stored in the directory
// provided, and follows the following naming convention.
// "{log_directory}/log_{zxid}"
// TODO: Consider packing multiple logs into the same file to save on resources.
type LogManager struct {
	// mu is a mutex that protects all the fields in the LogManager. In order
	// to keep LogManager thread-safe, we should hold the lock before reading/writing to any
	// of the fields in LogManager.
	mu       *sync.Mutex
	logPath  string
	LastZxid int64
}

func NewLogManager(logPath string) (*LogManager, error) {
	// Make sure to trim any trailing slashes if the provided path contains one.
	logPath = strings.TrimSuffix(logPath, "/")

	fileInfo, err := os.Stat(logPath)
	if err != nil {
		return nil, err
	}

	// Check if the file is a directory.
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("file path does not point to a directory")
	}
	return &LogManager{
		mu:       &sync.Mutex{},
		logPath:  logPath,
		LastZxid: 0,
	}, nil
}

// Append will append the given transaction to the log. We do this by writing to a new file on
// the filesystem.
func (l *LogManager) Append(txn *pbzk.Transaction) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if txn.GetZxid() <= l.LastZxid {
		return fmt.Errorf("transaction has already been added to the log")
	}

	// Create a new log file for this transaction id.
	fileName := fmt.Sprintf("%s/%s_%d", l.logPath, LogFilePrefix, txn.GetZxid())
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating new file: %w", err)
	}
	defer file.Close()

	bytes, err := proto.Marshal(txn)
	if err != nil {
		return fmt.Errorf("error marshalling txn")
	}

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("error writing transaction to file")
	}

	// Update the last seen ZXID to be equal to the transaction we just wrote.
	// Do this after successfully writing the transaction to a file.
	l.LastZxid = txn.GetZxid()
	return nil
}
