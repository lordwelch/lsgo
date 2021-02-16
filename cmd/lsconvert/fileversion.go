package main

import (
	"github.com/jdferrell3/peinfo-go/peinfo"
)

func getPEFileVersion(filepath string) string {
	f, err := peinfo.Initialize(filepath, false, "", false)
	if err != nil {
		panic(err)
	}
	defer f.OSFile.Close()
	defer f.PEFile.Close()
	v, _, err := f.GetVersionInfo()
	if err != nil {
		panic(err)
	}
	return v["FileVersion"]
}
