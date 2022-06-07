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
	logAll        bool
}

func NewLogger(logAll bool) *Logger {
	return &Logger{
		totalAccesses: 0,
		hits:          0,
		totalBytes:    0,
		cachedBytes:   0,
		logAll:        logAll,
	}
}

func (l *Logger) registerAccess(key string, cachehit bool, servedBytes float64) {

	var msg string

	l.mu.Lock()
	defer l.mu.Unlock()

	l.totalAccesses++
	l.totalBytes += servedBytes

	hitRatio, byteHitRatio := l.calcRatios()

	if cachehit {
		l.hits++
		l.cachedBytes += servedBytes
		msg = fmt.Sprintf("CACHE HIT! For %s. HR: %.4f. BHR: %.4f", key, hitRatio, byteHitRatio)
	} else {
		msg = fmt.Sprintf("MISS or MODIFIED! Updating cache for %s. HR: %.4f. BHR: %.4f", key, hitRatio, byteHitRatio)
	}

	if l.logAll {
		log.Println(msg)
	}
}

func (l *Logger) log() {
	l.mu.Lock()
	defer l.mu.Unlock()

	hitRatio, byteHitRatio := l.calcRatios()

	log.Printf("HIT RATIO: %.4f", hitRatio)
	log.Printf("BYTE HIT RATIO: %.4f", byteHitRatio)
}

func (l *Logger) calcRatios() (hitRatio float64, byteHitRatio float64) {
	if l.totalAccesses != 0 {
		hitRatio = l.hits / l.totalAccesses
	}

	if l.totalBytes != 0 {
		byteHitRatio = l.cachedBytes / l.totalBytes
	}

	return
}
