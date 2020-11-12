package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/kr/pretty"
	lslib "github.com/lordwelch/golslib"
)

func main() {
	f, err := os.Open(os.Args[1])
	defer f.Close()

	l, err := lslib.ReadLSF(f)
	pretty.Log(err, l)
	v, err := xml.MarshalIndent(struct {
		lslib.Resource
		XMLName string `xml:"save"`
	}{l, ""}, "", "\t")
	fmt.Fprintln(os.Stderr, err)
	n := string(v)
	n = strings.ReplaceAll(n, "></version>", " />")
	n = strings.ReplaceAll(n, "></attribute>", " />")
	n = strings.ReplaceAll(n, "></node>", " />")
	n = strings.ReplaceAll(n, "false", "False")
	n = strings.ReplaceAll(n, "true", "True")
	n = strings.ReplaceAll(n, "&#39;", "'")
	fmt.Printf("%s%s", strings.ToLower(xml.Header), n)
}
