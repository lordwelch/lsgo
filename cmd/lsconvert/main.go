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

	"git.narnian.us/lordwelch/lsgo"

	"github.com/go-kit/kit/log"
	"github.com/kr/pretty"
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
		lsgo.Logger = lsgo.NewFilter(map[string][]string{
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
		switch {

		case !fi.IsDir():
			err = openLSF(v)
			if err != nil && !errors.As(err, &lsgo.HeaderError{}) {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

		case *recurse:
			_ = filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
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
				if err != nil && !errors.As(err, &lsgo.HeaderError{}) {
					fmt.Fprintln(os.Stderr, err)
				}
				return nil
			})

		default:
			fmt.Fprintf(os.Stderr, "lsconvert: %s: Is a directory\n", v)
			os.Exit(1)
		}
	}
}
func openLSF(filename string) error {
	var (
		l   *lsgo.Resource
		err error
		n   string
		f   interface {
			io.Writer
			io.StringWriter
		}
	)
	l, err = readLSF(filename)
	if err != nil {
		return fmt.Errorf("reading LSF file %s failed: %w", filename, err)
	}
	if *printResource {
		pretty.Log(l)
	}
	if *printXML || *write {
		n, err = marshalXML(l)
		if err != nil {
			return fmt.Errorf("creating XML from LSF file %s failed: %w", filename, err)
		}

		if *write {
			f, err = os.OpenFile(filename, os.O_TRUNC|os.O_RDWR, 0o666)
			if err != nil {
				return fmt.Errorf("writing XML from LSF file %s failed: %w", filename, err)
			}
		} else if *printXML {
			f = os.Stdout
		}

		err = writeXML(f, n)
		fmt.Fprint(f, "\n")
		if err != nil {
			return fmt.Errorf("writing XML from LSF file %s failed: %w", filename, err)
		}
	}
	return nil
}

func readLSF(filename string) (*lsgo.Resource, error) {
	var (
		l   lsgo.Resource
		f   *os.File
		err error
	)
	f, err = os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	l, err = lsgo.ReadLSF(f)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func marshalXML(l *lsgo.Resource) (string, error) {
	var (
		v   []byte
		err error
	)
	v, err = xml.MarshalIndent(struct {
		*lsgo.Resource
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
	n = strings.ReplaceAll(n, "&#34;", "&quot;")
	return n, nil
}

func writeXML(f io.StringWriter, n string) error {
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
