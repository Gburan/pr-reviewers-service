package randomizer

import (
	"math/rand"
	"time"
)

type Randomizer struct{}

func (r Randomizer) Shuffle(n int, swap func(i, j int)) {
	rand.New(rand.NewSource(time.Now().UnixNano())).Shuffle(n, swap)
}
