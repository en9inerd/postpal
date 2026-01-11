package main

import (
	"fmt"
	"os"

	"github.com/en9inerd/postpal/internal/auth"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <password>\n", os.Args[0])
		os.Exit(1)
	}

	password := os.Args[1]
	hash, err := auth.HashPassword(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(hash)
}
