package cache

import (
	"time"
)

const (
	Expired           time.Duration = 0
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = -2
)

type Consumer func(key string, value interface{}) bool
type PutListener func(key string, value interface{})
type RemoveListener func(key string, value interface{})
type ExpirationListener func(key string, value interface{})

type Cache interface {
	Get(key string) (value interface{}, ok bool)
	GetExpiration(key string) (expiration time.Duration, ok bool)
	GetExpectedExpiration(key string) (expiration time.Duration, ok bool)
	Put(key string, value interface{}, expiration time.Duration)
	PutIfAbsent(key string, value interface{}, expiration time.Duration) bool
	PutIfExists(key string, value interface{}, expiration time.Duration) bool
	Exists(key string) bool
	Remove(key string) bool
	Count() int
	Keys() []string
	Values() []interface{}
	ForEach(consumer Consumer)
	Clear()
	DeleteExpired()
	AddPutListener(listener PutListener)
	AddRemoveListener(listener RemoveListener)
	AddExpirationListener(listener ExpirationListener)
}
