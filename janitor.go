package cache

import "time"

type janitor struct {
	interval time.Duration
	stop     chan struct{}
}

func (j *janitor) run(cache Cache) {
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			cache.DeleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func (j *janitor) stopJanitor(cache Cache) {
	j.stop <- struct{}{}
}
