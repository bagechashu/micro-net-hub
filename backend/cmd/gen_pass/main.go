package main

import (
	"flag"
	"fmt"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/tools"
)

func main() {
	var encode bool
	var decode bool
	var pass string
	flag.BoolVar(&encode, "encode", false, "encode password")
	flag.BoolVar(&decode, "decode", false, "decode password")
	flag.StringVar(&pass, "pass", "", "encoded password")
	flag.Parse()

	config.InitConfig()

	if !encode && !decode && pass == "" {
		fmt.Println("[-h] to get Usage.")
		return
	}
	if encode {
		fmt.Printf("Password Encoded: %s", tools.NewGenPasswd(pass))
		return
	}
	if decode {
		fmt.Printf("Password Decoded: %s", tools.NewParsePasswd(pass))
		return
	}
}
