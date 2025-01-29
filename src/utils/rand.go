package utils

import (
	"math/rand"
	"time"
)

var _rand *rand.Rand = nil

func init() {
	_rand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func Rand() *rand.Rand {
	if _rand == nil {
		panic("nil Rand")
	}

	return _rand
}
