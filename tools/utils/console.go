package utils

import (
	"errors"
	"os"
)

// Read the password entered by stdin
func ReadPassWd() ([]byte, error) {
	var passWd [33]byte

	n, err := os.Stdin.Read(passWd[:])
	if err != nil {
		return nil, err
	}
	if n <= 1 {
		return nil, errors.New("read failure")
	}
	return passWd[:n-1], nil
}
