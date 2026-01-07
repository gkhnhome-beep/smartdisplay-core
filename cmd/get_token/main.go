package main

import (
	"fmt"
	"smartdisplay-core/internal/settings"
)

func main() {
	token, err := settings.DecryptToken()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(token)
}
