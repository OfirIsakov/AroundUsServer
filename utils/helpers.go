package helpers

import (
	"bytes"
	"encoding/gob"
)

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func IntInArray(needle int, hayStack []int) bool {
	for _, hay := range hayStack {
		if hay == needle {
			return true
		}
	}
	return false
}
