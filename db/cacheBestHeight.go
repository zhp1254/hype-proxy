package db

import "sync"

var (
	bestHeight int64
	lock       sync.RWMutex
)

func GetBestHeight() int64 {
	lock.RLock()
	defer lock.RUnlock()
	return bestHeight
}

func SetBestHeight(height int64) {
	lock.Lock()
	defer lock.Unlock()
	if height > bestHeight {
		bestHeight = height
	}
}
