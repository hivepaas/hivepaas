package gocronqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging/mocks"
)

// lines is a logger that keeps what it is told.
type lines struct {
	mu   sync.Mutex
	logs []string
}

func (l *lines) add(level, template string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, level+" "+fmt.Sprintf(template, args...))
}

// has says whether a line holds the text.
func (l *lines) has(text string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, line := range l.logs {
		if strings.Contains(line, text) {
			return true
		}
	}
	return false
}

func (l *lines) Info(msg string, _ ...any)           { l.add("info", "%s", msg) }
func (l *lines) Error(msg string, _ ...any)          { l.add("error", "%s", msg) }
func (l *lines) Debug(msg string, _ ...any)          { l.add("debug", "%s", msg) }
func (l *lines) Warn(msg string, _ ...any)           { l.add("warn", "%s", msg) }
func (l *lines) Infof(template string, args ...any)  { l.add("info", template, args...) }
func (l *lines) Errorf(template string, args ...any) { l.add("error", template, args...) }
func (l *lines) Warnf(template string, args ...any)  { l.add("warn", template, args...) }
func (l *lines) Debugf(template string, args ...any) { l.add("debug", template, args...) }
func (l *lines) Fatal(_ ...any)                      {}
func (l *lines) Panic(_ ...any)                      {}
func (l *lines) Fatalf(string, ...any)               {}
func (l *lines) Panicf(string, ...any)               {}

func serverStartedAt(startedAt time.Time) (*Server, *lines) {
	logs := &lines{}
	return &Server{
		config:            &Config{Logger: logs},
		jobMap:            make(map[string]*jobData),
		startedAt:         startedAt,
		pauseLimit:        defaultPauseLimit,
		pauseRecheckEvery: defaultPauseRecheck,
	}, logs
}

// An update pauses the schedulers of the processes it is about to replace. A
// process started after the stop was sent is not one of them: the message, left
// in the list by a process that went before reading it, stopped beta3's app for
// good the moment it started.
func TestAStopSentBeforeTheServerStartedIsIgnored(t *testing.T) {
	started := time.Now()
	server, logs := serverStartedAt(started)

	server.handleCtrlMessage(context.Background(), &Message{StopScheduler: true, SentAt: started.Add(-5 * time.Second)})
	// One from a release that did not say when it sent it is as old.
	server.handleCtrlMessage(context.Background(), &Message{StopScheduler: true})

	assert.False(t, server.paused(time.Now()))
	assert.True(t, logs.has("ignored a stop"), logs.logs)
}

// A stop sent while the process runs is meant for it: the scheduler is paused,
// for pauseLimit at most, and says so.
func TestAStopSentAfterTheServerStartedPausesIt(t *testing.T) {
	started := time.Now()
	server, logs := serverStartedAt(started)

	server.handleCtrlMessage(context.Background(), &Message{StopScheduler: true, SentAt: started.Add(time.Second)})

	assert.True(t, server.paused(time.Now()))
	assert.WithinDuration(t, time.Now().Add(defaultPauseLimit), server.pausedUntil, time.Second)
	assert.True(t, logs.has("paused for at most"), logs.logs)
}

// A start ends the pause at once, and says so.
func TestAStartEndsThePause(t *testing.T) {
	server, logs := serverStartedAt(time.Now())
	assert.NoError(t, server.StopScheduler())

	server.handleCtrlMessage(context.Background(), &Message{StartScheduler: true})

	assert.False(t, server.paused(time.Now()))
	assert.True(t, logs.has("resumed"), logs.logs)
}

// BLPOP coming back empty is the listener's every five seconds, not a failure:
// it listens again at once. Waiting ten seconds after each had a process deaf
// two thirds of the time - long enough to miss the update's stop entirely.
func TestAnEmptyReadIsNotWaitedOut(t *testing.T) {
	assert.Equal(t, time.Duration(0), ctrlReadBackoff(hperrors.NewNotFound(taskQueueCtrlKey)))
	assert.Equal(t, ctrlReadErrorBackoff, ctrlReadBackoff(errors.New("connection refused")))
}

// What a client sends says when it was sent.
func TestTheClientSaysWhenItSentAStop(t *testing.T) {
	mr := &mockRedisClient{}
	client, _ := NewClient(mr, &mocks.Logger{})
	before := time.Now()

	assert.NoError(t, client.StopScheduler(context.Background()))

	var msg Message
	if assert.Len(t, mr.rpushValues, 1) {
		assert.NoError(t, json.Unmarshal([]byte(mr.rpushValues[0].(string)), &msg))
	}
	assert.True(t, msg.StopScheduler)
	assert.False(t, msg.SentAt.Before(before.Truncate(time.Second)), "sentAt %v", msg.SentAt)
}
