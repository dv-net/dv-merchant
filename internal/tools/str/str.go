package str

import (
	"crypto/rand"
	"io"
)

func RandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b), nil
}

func RemoveDuplicates(strings []string) []string {
	uniqueMap := make(map[string]struct{})
	var result []string

	for _, str := range strings {
		if _, exists := uniqueMap[str]; !exists {
			uniqueMap[str] = struct{}{}
			result = append(result, str)
		}
	}
	return result
}
