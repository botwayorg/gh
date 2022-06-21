package utils

import (
	"fmt"
	"os"
)

func AuthMessage() {
	fmt.Println("You're not authenticated, to authenticate run `botway github login`")

	os.Exit(0)
}
