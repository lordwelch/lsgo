package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/kr/pretty"
	lslib "github.com/lordwelch/golslib"
)

var (
	write         = flag.Bool("w", false, "replace the file with xml data")
	printXML      = flag.Bool("x", false, "print xml to stdout")
	printResource = flag.Bool("R", false, "print the resource struct to stderr")
	recurse       = flag.Bool("r", false, "recurse into directories")
	logging       = flag.Bool("l", false, "enable logging to stderr")
	parts         = flag.String("p", "", "parts to filter logging for, comma separated")
)

func init() {
	flag.Parse()
	if *logging {
		lslib.Logger = lslib.NewFilter(map[string][]string{
			"part": strings.Split(*parts, ","),
		}, log.NewLogfmtLogger(os.Stderr))
	}
}

func main() {

	for _, v := range flag.Args() {
		fi, err := os.Stat(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !fi.IsDir() {
			err = openLSF(v)
			if err != nil && !errors.As(err, &lslib.HeaderError{}) {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else if *recurse {
			filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					if info.Name() == ".git" {
						return filepath.SkipDir
					}
					return nil
				}
				err = openLSF(path)
				if err != nil && !errors.As(err, &lslib.HeaderError{}) {
					fmt.Fprintln(os.Stderr, err)
				}
				return nil
			})
		} else {
			fmt.Fprintf(os.Stderr, "lsconvert: %s: Is a directory\n", v)
			os.Exit(1)
		}
	}
}
func openLSF(filename string) error {
	var (
		l   *lslib.Resource
		err error
		n   string
		f   strwr
	)
	l, err = readLSF(filename)
	if err != nil {
		return fmt.Errorf("Reading LSF file %s failed: %w\n", filename, err)
	}
	if *printResource {
		pretty.Log(l)
	}
	if *printXML || *write {
		n, err = marshalXML(l)
		if err != nil {
			return fmt.Errorf("Creating XML from LSF file %s failed: %w\n", filename, err)
		}

		if *write {
			f, err = os.OpenFile(filename, os.O_TRUNC|os.O_RDWR, 0o666)
			if err != nil {
				return fmt.Errorf("Writing XML from LSF file %s failed: %w\n", filename, err)
			}
		} else if *printXML {
			f = os.Stdout
		}

		err = writeXML(f, n)
		fmt.Fprint(f, "\n")
		if err != nil {
			return fmt.Errorf("Writing XML from LSF file %s failed: %w\n", filename, err)
		}
	}
	return nil
}

func readLSF(filename string) (*lslib.Resource, error) {
	var (
		l   lslib.Resource
		f   *os.File
		err error
	)
	f, err = os.Open(filename)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	l, err = lslib.ReadLSF(f)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func marshalXML(l *lslib.Resource) (string, error) {
	var (
		v   []byte
		err error
	)
	v, err = xml.MarshalIndent(struct {
		*lslib.Resource
		XMLName string `xml:"save"`
	}{l, ""}, "", "\t")
	if err != nil {
		return string(v), err
	}
	n := string(v)
	n = strings.ReplaceAll(n, "></version>", " />")
	n = strings.ReplaceAll(n, "></attribute>", " />")
	n = strings.ReplaceAll(n, "></node>", " />")
	n = strings.ReplaceAll(n, "false", "False")
	n = strings.ReplaceAll(n, "true", "True")
	n = strings.ReplaceAll(n, "&#39;", "'")
	return n, nil
}

type strwr interface {
	io.Writer
	io.StringWriter
}

func writeXML(f strwr, n string) error {
	var (
		err error
	)
	_, err = f.WriteString(strings.ToLower(xml.Header))
	if err != nil {
		return err
	}
	_, err = f.WriteString(n)
	if err != nil {
		return err
	}
	return nil
}
