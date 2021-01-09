package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5"

	ie "git.narnian.us/lordwelch/lsgo/internal/exec"
	"golang.org/x/sys/windows/registry"
)

var (
	initFS    = flag.NewFlagSet("init", flag.ExitOnError)
	initFlags struct {
		BG3Path    string
		divinePath string
	}
)

func initRepository(arguments []string) {
	var (
		BG3Path, divinePath string
		repo                *git.Repository
		err                 error
	)
	BG3Path, divinePath = preRequisites()
	if BG3Path == "" {
		panic("Could not locate Baldur's Gate 3")
	}
	fmt.Println(divinePath)
	panic("Baldur's Gate 3 located " + BG3Path)
	repo, err = git.PlainInit(".", false)
	if err != nil {
		panic(err)
	}
	repo.Fetch(nil)
}

func preRequisites() (BG3Path, divinePath string) {
	BG3Path = locateBG3()
	divinePath = locateLSLIB()
	return BG3Path, divinePath
}

// locateLSLIB specifically locates divine.exe
func locateLSLIB() string {
	var searchPath []string
	folderName := "ExportTool-*"
	if runtime.GOOS == "windows" {
		for _, v := range "cdefgh" {
			for _, vv := range []string{`:\Program Files*\*\`, `:\Program Files*\*\`, `:\app*\`, `:\`} {
				paths, err := filepath.Glob(filepath.Join(string(v)+vv, folderName))
				if err != nil {
					panic(err)
				}
				searchPath = append(searchPath, paths...)
			}
		}
		for _, v := range []string{os.ExpandEnv(`${USERPROFILE}\app*\`), os.ExpandEnv(`${USERPROFILE}\Downloads\`), os.ExpandEnv(`${USERPROFILE}\Documents\`)} {
			paths, err := filepath.Glob(filepath.Join(v, folderName))
			if err != nil {
				panic(err)
			}
			searchPath = append(searchPath, paths...)
		}
	}
	divine, _ := ie.LookPath("divine.exe", strings.Join(searchPath, string(os.PathListSeparator)))
	return divine
}

func locateBG3() string {
	installPath := initFlags.BG3Path
	if installPath == "" {
		installPath = checkRegistry()
	}
	return installPath
}

func checkRegistry() string {
	var (
		k           registry.Key
		err         error
		installPath string
	)
	k, err = registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\1456460669_is1`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()
	for _, v := range []string{"Inno Setup: App Path", "InstallLocation"} {
		var s string
		s, _, err = k.GetStringValue(v)
		if err == nil {
			installPath = s
			break
		}
	}

	return installPath
}
