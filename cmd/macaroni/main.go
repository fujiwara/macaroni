package main

import (
	"fmt"
	"os"

	"github.com/fujiwara/macaroni"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("macaroni version:", macaroni.Version)
		return
	}
	conf := macaroni.BuildConfig()
	err := macaroni.Run(conf, os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
