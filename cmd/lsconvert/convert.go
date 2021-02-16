package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"git.narnian.us/lordwelch/lsgo"
	_ "git.narnian.us/lordwelch/lsgo/lsb"
	_ "git.narnian.us/lordwelch/lsgo/lsf"

	"github.com/go-kit/kit/log"
	"github.com/kr/pretty"
)

type convertFlags struct {
	write         bool
	printXML      bool
	printResource bool
	recurse       bool
	logging       bool
	parts         string
}

func convert(arguments ...string) error {
	var (
		convertFS = flag.NewFlagSet("convert", flag.ExitOnError)
		cvFlags   = convertFlags{}
	)
	convertFS.BoolVar(&cvFlags.write, "w", false, "Replace the file with XML data")
	convertFS.BoolVar(&cvFlags.printXML, "x", false, "Print XML to stdout")
	convertFS.BoolVar(&cvFlags.printResource, "R", false, "Print the resource struct to stderr")
	convertFS.BoolVar(&cvFlags.recurse, "r", false, "Recurse into directories")
	convertFS.BoolVar(&cvFlags.logging, "l", false, "Enable logging to stderr")
	convertFS.StringVar(&cvFlags.parts, "p", "", "Parts to filter logging for, comma separated")
	convertFS.Parse(arguments)
	if cvFlags.logging {
		lsgo.Logger = lsgo.NewFilter(map[string][]string{
			"part": strings.Split(cvFlags.parts, ","),
		}, log.NewLogfmtLogger(os.Stderr))
	}

	for _, v := range convertFS.Args() {
		fi, err := os.Stat(v)
		if err != nil {
			return err
		}
		switch {
		case !fi.IsDir():
			err = cvFlags.openLSF(v)
			if err != nil && !errors.As(err, &lsgo.HeaderError{}) {
				return err
			}

		case cvFlags.recurse:
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
				err = cvFlags.openLSF(path)
				if err != nil && !errors.Is(err, lsgo.ErrFormat) {
					fmt.Fprintln(os.Stderr, err)
				}
				return nil
			})

		default:
			return fmt.Errorf("lsconvert: %s: Is a directory\n", v)
		}
	}
	return nil
}

func (cf *convertFlags) openLSF(filename string) error {
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
	if cf.printResource {
		pretty.Log(l)
	}
	if cf.printXML || cf.write {
		n, err = marshalXML(l)
		if err != nil {
			return fmt.Errorf("creating XML from LSF file %s failed: %w", filename, err)
		}

		if cf.write {
			f, err = os.OpenFile(filename, os.O_TRUNC|os.O_RDWR, 0o666)
			if err != nil {
				return fmt.Errorf("writing XML from LSF file %s failed: %w", filename, err)
			}
		} else if cf.printXML {
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
		l    lsgo.Resource
		r    io.ReadSeeker
		file *os.File
		fi   os.FileInfo
		err  error
	)
	switch filepath.Ext(filename) {
	case ".lsf", ".lsb":
		var b []byte
		fi, err = os.Stat(filename)
		if err != nil {
			return nil, err
		}
		// Arbitrary size, no lsf file should reach 100 MB (I haven't found one over 90 KB)
		// and if you don't have 100 MB of ram free you shouldn't be using this
		if fi.Size() <= 100*1024*1024 {
			b, err = ioutil.ReadFile(filename)
			if err != nil {
				return nil, err
			}
			r = bytes.NewReader(b)
			break
		}
		fallthrough
	default:
		b := make([]byte, 4)
		file, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		_, err = file.Read(b)
		if err != nil {
			return nil, err
		}
		if !lsgo.SupportedFormat(b) {
			return nil, lsgo.ErrFormat
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
		fi, _ = os.Stat(filename)

		// I have never seen a valid "ls*" file over 90 KB
		if fi.Size() < 1*1024*1024 {
			b, err = ioutil.ReadAll(file)
			if err != nil {
				return nil, err
			}
			r = bytes.NewReader(b)
		} else {
			r = file
		}
	}

	l, _, err = lsgo.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decoding %q failed: %w", filename, err)
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
	var err error
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
