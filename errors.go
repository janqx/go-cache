package cache

import "errors"

var ErrKeyNotFound = errors.New("key not found")

var ErrKeyExpired = errors.New("key expired")
