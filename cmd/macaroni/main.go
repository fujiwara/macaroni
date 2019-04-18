package main

import (
	"fmt"
	"os"

	"github.com/fujiwara/macaroni"
)

func main() {
	conf := macaroni.BuildConfig()
	err := macaroni.Run(conf, os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
