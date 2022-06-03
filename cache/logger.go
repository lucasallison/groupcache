package groupcache

import (
	"fmt"
	"log"
	"sync"
)

type Logger struct {
	mu            sync.RWMutex
	log           bool
	totalAccesses float64
	hits          float64
}

func NewLogger(log bool) *Logger {
	return &Logger{
		log:           log,
		totalAccesses: 0,
		hits:          0,
	}
}

func (l *Logger) registerAccess(key string, cachehit bool) {

	var msg string

	l.mu.Lock()
	defer l.mu.Unlock()

	l.totalAccesses++

	hitRatio := l.hits / l.totalAccesses

	if cachehit {
		l.hits++
		msg = fmt.Sprintf("CACHE HIT! For %s. HR: %.4f", key, hitRatio)
	} else {
		msg = fmt.Sprintf("MISS or MODIFIED! Updating cache for %s. HR: %.4f", key, hitRatio)
	}

	if l.log {
		log.Println(msg)
	}
}
