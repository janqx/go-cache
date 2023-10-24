package cache

import "time"

type item struct {
	value      interface{}
	expiration time.Duration
}

func newItem(value interface{}, expiration, defaultExpiration time.Duration) *item {
	result := &item{
		value: value,
	}
	if expiration == NoExpiration {
		result.expiration = NoExpiration
	} else {
		if expiration == DefaultExpiration {
			expiration = defaultExpiration
		}
		result.expiration = time.Duration(time.Now().Add(expiration).UnixNano())
	}
	return result
}

func (i *item) expired() bool {
	return (i.expiration != NoExpiration) && time.Now().UnixNano() > int64(i.expiration)
}

func (i *item) expectedExpiration() time.Duration {
	return i.expiration - time.Duration(time.Now().UnixNano())
}
