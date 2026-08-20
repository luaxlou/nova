package data

import "sync"

var (
	storeMu      sync.RWMutex
	usersByEmail = map[string]UserModel{}
)
