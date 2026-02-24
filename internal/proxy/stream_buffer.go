package proxy

import (
	"cmp"
	"encoding/json"
	"io"
	"slices"
	"time"

	log "github.com/sirupsen/logrus"
)

type LogsBuffer struct {
	logsBuffer []map[string]any
}

func NewLogsBuffer(size int) *LogsBuffer {
	return &LogsBuffer{
		logsBuffer: make([]map[string]any, 0, size),
	}
}

// Write sorts the buffer by time, writes all entries to w, and resets the buffer.
func (b *LogsBuffer) Write(w io.Writer) error {
	if b.IsEmpty() {
		return nil
	}

	if b.Size() > 1 {
		b.SortByTime()
	}

	encoder := json.NewEncoder(w)
	for _, item := range b.logsBuffer {
		if err := encoder.Encode(item); err != nil {
			return err
		}
	}

	b.Reset()
	return nil
}

// Reset clears the buffer while retaining its allocated capacity.
func (b *LogsBuffer) Reset() {
	b.logsBuffer = b.logsBuffer[:0]
}

func (b *LogsBuffer) SortByTime() {
	slices.SortStableFunc(b.logsBuffer, func(a, b map[string]any) int {
		tsA, okA := a["_time"].(string)
		tsB, okB := b["_time"].(string)
		if !okA || !okB {
			return 0
		}

		timeA, err := time.Parse(time.RFC3339Nano, tsA)
		if err != nil {
			log.Errorf("failed to parse timestamp: %v", err)
			return 0
		}

		timeB, err := time.Parse(time.RFC3339Nano, tsB)
		if err != nil {
			log.Errorf("failed to parse timestamp: %v", err)
			return 0
		}

		return cmp.Compare(timeB.Unix(), timeA.Unix())
	})
}

func (b *LogsBuffer) AddLog(log map[string]any) {
	b.logsBuffer = append(b.logsBuffer, log)
}

func (b *LogsBuffer) Size() int {
	return len(b.logsBuffer)
}

func (b *LogsBuffer) IsEmpty() bool {
	return b.Size() == 0
}
