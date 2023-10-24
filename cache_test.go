package cache_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/janqx/go-cache"
)

const keyChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateKey() string {
	result := ""
	for i := 0; i < 8; i++ {
		result += string(keyChars[rand.Intn(len(keyChars))])
	}
	return result
}

func Test_HashCache(t *testing.T) {
	start := time.Now()
	c := cache.NewHashCache(1*time.Minute, 1*time.Second)

	addCount := 0
	removeCount := 0
	expiredCount := 0

	c.AddPutListener(func(key string, value interface{}) {
		addCount++
	})

	c.AddRemoveListener(func(key string, value interface{}) {
		removeCount++
	})

	c.AddExpirationListener(func(key string, value interface{}) {
		expiredCount++
	})

	for i := 0; i < 10000; i++ {
		key := GenerateKey()
		c.PutIfAbsent(key, 1, time.Duration(rand.Intn(100))*time.Millisecond)
		c.Get(key)
		c.Exists(key)
	}

	c.Put("name", "jack", cache.DefaultExpiration)

	c.Put("val", 10, cache.NoExpiration)

	elapse := time.Since(start)

	c.ForEach(func(key string, value interface{}) bool {
		fmt.Printf("foreach key: %s\n", key)
		return true
	})

	fmt.Printf("elapse = %dms, count = %d, addCount = %d, removeCount = %d, expiredCount = %d\n", elapse.Milliseconds(), c.Count(), addCount, removeCount, expiredCount)
}
