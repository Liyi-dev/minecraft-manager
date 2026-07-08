package websocket

import (
	"bufio"
	"io"
	"os"
	"time"

	"minecraft-manager/pkg/logger"
)

// LogWatcher tails a log file and broadcasts lines to the hub.
type LogWatcher struct {
	hub     *Hub
	logPath string
	stopCh  chan struct{}
}

func NewLogWatcher(hub *Hub, logPath string) *LogWatcher {
	return &LogWatcher{
		hub:     hub,
		logPath: logPath,
		stopCh:  make(chan struct{}),
	}
}

// Start begins tailing the log file.
func (w *LogWatcher) Start() {
	go w.watch()
}

// Stop stops the log watcher.
func (w *LogWatcher) Stop() {
	close(w.stopCh)
}

func (w *LogWatcher) watch() {
	for {
		select {
		case <-w.stopCh:
			return
		default:
		}

		file, err := os.Open(w.logPath)
		if err != nil {
			logger.Warn.Printf("LogWatcher: cannot open %s: %v (retrying in 5s)", w.logPath, err)
			w.sleep(5 * time.Second)
			continue
		}

		// Seek to end of file
		file.Seek(0, io.SeekEnd)

		reader := bufio.NewReader(file)
		for {
			select {
			case <-w.stopCh:
				file.Close()
				return
			default:
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// No new data, wait a bit
					w.sleep(500 * time.Millisecond)
					continue
				}
				// File probably rotated or deleted
				logger.Warn.Printf("LogWatcher: read error: %v", err)
				break
			}

			if line != "" {
				w.hub.BroadcastLog(line)
			}
		}

		file.Close()
		// Retry after a short delay (file might have been rotated)
		w.sleep(2 * time.Second)
	}
}

func (w *LogWatcher) sleep(d time.Duration) {
	select {
	case <-w.stopCh:
	case <-time.After(d):
	}
}
