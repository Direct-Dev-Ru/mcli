package mcliutils

import (
	"math/rand"
	"time"
)

func Random(min, max int) int {
	b := make([]byte, 2)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	rand.Seed(time.Now().Unix()*int64(b[0]) + int64(b[1]))
	return rand.Intn(max-min) + min
}
