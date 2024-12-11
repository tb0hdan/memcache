package memcache

import (
	"sync"
	"time"
)

// Logger - memcache logger interface. Can be stdlib log, logrus etc
type Logger interface {
	Printf(fmt string, args ...interface{})
	Debug(s ...interface{})
}

// ValueType - memcache item
type ValueType struct {
	Value    interface{}
	Expires  int64
	MetaData string
}

// CacheType - cache structure with bound methods
type CacheType struct {
	cache              map[string]*ValueType
	items              []*ValueType
	m                  sync.RWMutex
	ticker             *time.Ticker
	done               chan struct{}
	logger             Logger
	lockWithKeyTimeout time.Duration
}
