package main

import (
	"fmt"
	"os"
)

func main() {
	// flag.Parse()
	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	switch arg {
	case "convert":
		err := convert(os.Args[1:]...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "init":
		p := "."
		if len(os.Args) > 2 {
			p = os.Args[2]
		}
		repo, err := openRepository(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		err = repo.Init()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		err := convert(os.Args[1:]...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
