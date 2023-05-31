package filebuf

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/anthony-ozdemir/zfse/internal/helper"
	"go.uber.org/zap"
)

type FileOutputBufferOptions struct {
	BulkOutputLimit int64
	FilePath        string
	OnAppendCB      func()
	OnFlushCB       func(int)
}

type FileOutputBuffer struct {
	opts             FileOutputBufferOptions
	buffer           string
	mutex            sync.RWMutex
	appendedLinesQty int
}

func NewFileOutputBuffer(opts FileOutputBufferOptions) *FileOutputBuffer {
	o := FileOutputBuffer{}
	o.opts = opts
	o.appendedLinesQty = 0
	return &o
}

func (o *FileOutputBuffer) AppendToFile(line string) {
	o.mutex.Lock()
	o.appendedLinesQty++
	o.buffer += line + "\n"

	if int64(len(o.buffer)) >= o.opts.BulkOutputLimit {
		o.flushToFile(o.opts.FilePath)
	}

	if o.opts.OnAppendCB != nil {
		o.opts.OnAppendCB()
	}

	o.mutex.Unlock()
}

func (o *FileOutputBuffer) Flush() {
	o.mutex.Lock()
	o.flushToFile(o.opts.FilePath)
	if o.opts.OnFlushCB != nil {
		o.opts.OnFlushCB(o.appendedLinesQty)
	}
	o.mutex.Unlock()
}

func (o *FileOutputBuffer) flushToFile(filePath string) {
	if len(o.buffer) == 0 {
		return
	}

	// Ensure parent directory exists
	err := helper.CreateFolder(filepath.Dir(filePath))
	if err != nil {
		zap.L().Fatal("Unable to create folder.", zap.String("err", err.Error()))
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		zap.L().Fatal("Unable to open file.", zap.String("err", err.Error()))
	}
	defer file.Close()

	if _, err := file.WriteString(o.buffer); err != nil {
		zap.L().Fatal("Unable to write to file.", zap.String("err", err.Error()))
	}
	o.buffer = ""
}
