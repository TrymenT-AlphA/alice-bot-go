package util

import (
	"math/rand"
	"time"
)

// GetDice return a dice in [0, n)
func GetDice(n int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(n)
}
