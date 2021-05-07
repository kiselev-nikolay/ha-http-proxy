package trace_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/kiselev-nikolay/ha-http-proxy/pure/pkg/trace"
)

const idRegExp = `^[A-z0-9+\/]+$`

func TestGenerateID(t *testing.T) {
	id := trace.GenerateID()
	r := regexp.MustCompile(idRegExp)
	if !r.Match([]byte(id)) {
		t.Errorf("id = %s; want match %s", id, idRegExp)
	}
}

func TestGenerateIDWith(t *testing.T) {
	for length := 0; length < 2^16; length++ {
		testname := fmt.Sprintf("with length = %d", length)
		t.Run(testname, func(t *testing.T) {
			id := trace.GenerateIDWithLength(length)
			if len(id) != length {
				t.Errorf("id = '%s', len(id) = %d; want string with len() = %d", id, len(id), length)
			}
		})
	}
}
