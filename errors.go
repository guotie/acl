package acl

import (
	"fmt"
)

var (
	// ErrRedisKeyNotExist Key not exist
	ErrRedisKeyNotExist = fmt.Errorf("Redis key Not Exist")
	// ErrRedisSubkeyNotExist Subkey not exist
	ErrRedisSubkeyNotExist = fmt.Errorf("Redis Subkey Not Exist")
)
