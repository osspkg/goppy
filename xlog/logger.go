/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xlog

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go.osspkg.com/goppy/iosync"
)

// log base model
type log struct {
	status    uint32
	writer    io.Writer
	entities  sync.Pool
	formatter Formatter
	channel   chan []byte
	mux       iosync.Lock
	wg        iosync.Group
}

// New init new logger
func New() Logger {
	object := &log{
		status:    LevelError,
		writer:    os.Stdout,
		formatter: NewFormatJSON(),
		channel:   make(chan []byte, 1024),
		wg:        iosync.NewGroup(),
		mux:       iosync.NewLock(),
	}
	object.entities = sync.Pool{
		New: func() interface{} {
			return newEntity(object)
		},
	}
	object.wg.Background(func() {
		object.queue()
	})
	return object
}

func (l *log) SendMessage(level uint32, call func(v *Message)) {
	if l.GetLevel() < level {
		return
	}

	m, ok := poolMessage.Get().(*Message)
	if !ok {
		m = &Message{}
	}

	call(m)
	lvl, ok := levels[level]
	if !ok {
		lvl = "UNK"
	}
	m.Level, m.Time = lvl, time.Now().Unix()

	l.mux.RLock(func() {
		b, err := l.formatter.Encode(m)
		if err != nil {
			b = []byte(err.Error())
		}

		select {
		case l.channel <- b:
		default:
		}
	})

	m.Reset()
	poolMessage.Put(m)
}

func (l *log) queue() {
	for {
		b, ok := <-l.channel
		if !ok {
			return
		}
		if b == nil {
			return
		}
		l.mux.RLock(func() {
			l.writer.Write(b) //nolint:errcheck
		})
	}
}

func (l *log) getEntity() *entity {
	lw, ok := l.entities.Get().(*entity)
	if !ok {
		lw = newEntity(l)
	}
	return lw
}

func (l *log) PutEntity(v *entity) {
	v.Reset()
	l.entities.Put(v)
}

// Close waiting for all messages to finish recording
func (l *log) Close() {
	l.channel <- nil
	l.wg.Wait()
}

// SetOutput change writer
func (l *log) SetOutput(out io.Writer) {
	l.mux.Lock(func() {
		l.writer = out
	})
}

func (l *log) SetFormatter(f Formatter) {
	l.mux.Lock(func() {
		l.formatter = f
	})
}

// SetLevel change log level
func (l *log) SetLevel(v uint32) {
	atomic.StoreUint32(&l.status, v)
}

// GetLevel getting log level
func (l *log) GetLevel() uint32 {
	return atomic.LoadUint32(&l.status)
}

// Infof info message
func (l *log) Infof(format string, args ...interface{}) {
	l.getEntity().Infof(format, args...)
}

// Warnf warning message
func (l *log) Warnf(format string, args ...interface{}) {
	l.getEntity().Warnf(format, args...)
}

// Errorf error message
func (l *log) Errorf(format string, args ...interface{}) {
	l.getEntity().Errorf(format, args...)
}

// Debugf debug message
func (l *log) Debugf(format string, args ...interface{}) {
	l.getEntity().Debugf(format, args...)
}

// Fatalf fatal message and exit
func (l *log) Fatalf(format string, args ...interface{}) {
	l.getEntity().Fatalf(format, args...)
}

// WithFields setter context to log message
func (l *log) WithFields(v Fields) Writer {
	return l.getEntity().WithFields(v)
}

// WithError setter context to log message
func (l *log) WithError(key string, err error) Writer {
	return l.getEntity().WithError(key, err)
}

// WithField setter context to log message
func (l *log) WithField(key string, value interface{}) Writer {
	return l.getEntity().WithField(key, value)
}
