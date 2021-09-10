package helpers

import "fmt"

func GetBytes(key interface{}) ([]byte, error) {
	buf, ok := key.([]byte)
	if !ok {
		return nil, fmt.Errorf("bruh, not bytes here")
	}

	return buf, nil
}
