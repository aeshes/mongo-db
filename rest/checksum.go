package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Sha256 calculates SHA-256 chechsum for the file defined by its path
func Sha256(pathToFile string) string {
	file, err := os.Open(pathToFile)
	if err != nil {
		log.Println(err)
		return ""
	}

	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Println(err)
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%x", hasher.Sum(nil))
	return b.String()
}
