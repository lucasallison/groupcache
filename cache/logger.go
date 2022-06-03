package groupcache

import (
	"fmt"
	"log"
	"sync"
)

type Logger struct {
	mu            sync.RWMutex
	totalAccesses float64
	hits          float64
	totalBytes    float64
	cachedBytes   float64
}

func NewLogger() *Logger {
	return &Logger{
		totalAccesses: 0,
		hits:          0,
		totalBytes:    0,
		cachedBytes:   0,
	}
}

func (l *Logger) registerAccess(key string, cachehit bool, servedBytes float64) {

	var msg string

	l.mu.Lock()
	defer l.mu.Unlock()

	l.totalAccesses++
	l.totalBytes += servedBytes

	hitRatio := l.hits / l.totalAccesses

	var byteHitRatio float64
	if l.totalBytes != 0 {
		byteHitRatio = l.cachedBytes / l.totalBytes
	}

	if cachehit {
		l.hits++
		l.cachedBytes += servedBytes
		msg = fmt.Sprintf("CACHE HIT! For %s. HR: %.4f. BHR: %.4f", key, hitRatio, byteHitRatio)
	} else {
		msg = fmt.Sprintf("MISS or MODIFIED! Updating cache for %s. HR: %.4f. BHR: %.4f", key, hitRatio, byteHitRatio)
	}

	log.Println(msg)
}
