package flower

import (
	"testing"

	"github.com/cheekybits/is"
)

func TestRandom(t *testing.T) {
	is := is.New(t)
	r := randomKey(10)
	is.OK(r)
	is.Equal(10, len(r))
}
