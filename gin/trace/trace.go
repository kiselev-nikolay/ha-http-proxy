package trace

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"log"
	"math"
)

func GenerateIDWithLength(length int) string {
	if length <= 0 {
		return ""
	}

	// to minimize unnecessary work, we take only as many bytes as needed
	size := 2 + math.Ceil(float64(length)*3/4)
	data := make([]byte, int(size))
	_, err := rand.Read(data)
	if err != nil {
		// rand.Read error is allways nil
		log.Printf("trace.GenerateID unexpected error: %v", err)
		return ""
	}

	// the simplest way to convert random bytes to a string is to base64 encode them
	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.RawStdEncoding, buf)
	defer encoder.Close()
	_, err = encoder.Write(data)
	if err != nil {
		// encoder.Write error is allways nil
		log.Printf("trace.GenerateID unexpected error: %v", err)
		return ""
	}

	// since calculating the length of a input random array can be one character more
	// its better to truncate buffer
	randomString := buf.String()
	id := randomString[:length]
	return id
}

func GenerateID() string {
	return GenerateIDWithLength(32)
}
