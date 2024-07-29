package util

import (
	"math/rand"
	"strings"
)

func RandomName() string {
	letters := `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`
	s := make([]string, len(letters))
	for i := 0; i < len(letters); i++ {
		s[i] = string(letters[i])
	}

	result := make([]string, 5)
	for i := 0; i < len(result); i++ {
		randIndex := rand.Intn(len(s))
		result[i] = s[randIndex]
	}
	return strings.Join(result, "")
}
