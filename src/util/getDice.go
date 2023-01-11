package util

import (
	"math/rand"
)

// GetDice return a dice in [0, n)
func GetDice(n int) int {
	return rand.Intn(n)
}
