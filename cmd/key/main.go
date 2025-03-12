package main

import (
	"fmt"

	"github.com/hesoyamTM/apphelper-sso/internal/lib/jwt"
)

func main() {
	privKey, publicKey, err := jwt.GenerateKeys()
	if err != nil {
		panic(err)
	}

	encodedPrivKey, encodedPubKey := jwt.Encode(privKey, publicKey)
	fmt.Println(encodedPrivKey)
	fmt.Println(encodedPubKey)
}
