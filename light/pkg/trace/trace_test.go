package trace_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/kiselev-nikolay/ha-http-proxy/light/pkg/trace"
)

const idRegExp = `^[A-z0-9+\/]+$`

func TestGetID(t *testing.T) {
	id := trace.GetID()
	r := regexp.MustCompile(idRegExp)
	if !r.Match([]byte(id)) {
		t.Errorf("id = %s; want match %s", id, idRegExp)
	}
}

func TestGetIDWith(t *testing.T) {
	for lenght := 0; lenght < 2^16; lenght++ {
		testname := fmt.Sprintf("with lenght = %d", lenght)
		t.Run(testname, func(t *testing.T) {
			id := trace.GetIDWithLenght(lenght)
			if len(id) != lenght {
				t.Errorf("id = '%s', len(id) = %d; want string with len() = %d", id, len(id), lenght)
			}
		})
	}
}
