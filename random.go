package flower

import (
	"math/rand"
	"sync"
	"time"
)

// randomKeyCharacters is a []byte of the characters to choose from when generating
// random keys.
var randomKeyCharacters = []byte("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// seedOnce is the sync.Once used to rand.Seed.
var seedOnce sync.Once

// RandomKey generates a random key at the given length.
//
// The first time this is called, the rand.Seed will be set
// to the current time.
func randomKey(length int) string {

	// randomise the seed
	seedOnce.Do(func() {
		rand.Seed(time.Now().UTC().UnixNano())
	})

	// Credit: http://stackoverflow.com/questions/12321133/golang-random-number-generator-how-to-seed-properly

	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		randInt := randInt(0, len(randomKeyCharacters))
		bytes[i] = randomKeyCharacters[randInt : randInt+1][0]
	}
	return string(bytes)

}

// randInt generates a random integer between min and max.
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
