package main

import "flag"

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "convert":
		convert(flag.Args()[1:])
	case "init":
		initRepository(flag.Args()[1:])
	default:
		convert(flag.Args())
	}
}
