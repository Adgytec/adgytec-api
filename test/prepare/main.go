package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateSecureToken() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func main() {
	fmt.Println(GenerateSecureToken())

}
