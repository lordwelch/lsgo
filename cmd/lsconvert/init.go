package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/hashicorp/go-multierror"

	"git.narnian.us/lordwelch/lsgo/gog"
	ie "git.narnian.us/lordwelch/lsgo/internal/exec"
	"golang.org/x/sys/windows/registry"
)

var (
	initFS    = flag.NewFlagSet("init", flag.ExitOnError)
	initFlags struct {
		BG3Path    string
		divinePath string
		Repo       string
	}
)

const ignore = `# ignores all binary files
*.wem
*.dds
*.png
*.jpg
*.gr2
*.bin
*.patch
*.bshd
*.tga
*.ttf
*.fnt
Assets_pak/Public/Shared/Assets/Characters/_Anims/Humans/_Female/HUM_F_Rig/HUM_F_Rig_DFLT_SPL_Somatic_MimeTalk_01
Assets_pak/Public/Shared/Assets/Characters/_Models/_Creatures/Beholder/BEH_Spectator_A_GM_DDS
Assets_pak/Public/Shared/Assets/Consumables/CONS_GEN_Food_Turkey_ABC/CONS_GEN_Food_Turkey_ABC.ma
Assets_pak/Public/Shared/Assets/Decoration/Harbour/DEC_Harbour_Shell_ABC/DEC_Harbour_Shell_Venus_A
*.bnk
*.pak
*.data
*.osi
*.cur
*.gtp
*.gts
*.ffxanim
*.ffxactor
*.ffxbones
*.bk2
`

type BG3Repo struct {
	git        *git.Repository
	gitPath    string
	gog        gog.GOGalaxy
	BG3Path    string
	divinePath string
	Empty      bool
}

func openRepository(path string) (BG3Repo, error) {
	var (
		b   BG3Repo
		err error
	)
	b.gitPath = path
	b.BG3Path, b.divinePath = preRequisites()
	if b.BG3Path == "" {
		panic("Could not locate Baldur's Gate 3")
	}

	b.git, err = git.PlainOpen(b.gitPath)
	if err != nil {
		b.Empty = true
	}
	return b, nil
}

func (bgr *BG3Repo) Init() error {
	var err error
	if bgr.Empty || bgr.git == nil {
		bgr.git, err = git.PlainInit(bgr.gitPath, false)
		if err == git.ErrRepositoryAlreadyExists {
			bgr.git, err = git.PlainOpen(bgr.gitPath)
			if err != nil {
				return err
			}
		}
	}

	if citer, err := bgr.git.CommitObjects(); err == nil {
		if _, err := citer.Next(); err == nil {
			// Repo contains commits, nothing to do
			return nil
		}
	}
	return bgr.AddNewVersion()
}

func (bgr *BG3Repo) RetrieveGOGalaxy() error {
	var err error
	bgr.gog, err = gog.RetrieveGOGInfo("1456460669")
	if err != nil {
		return err
	}
	return nil
}

func (bgr *BG3Repo) AddNewVersion() error {
	var (
		version   string
		changelog gog.Change
		err       error
	)
	version = getPEFileVersion(filepath.Join(bgr.BG3Path, "bin/bg3.exe"))
	err = bgr.RetrieveGOGalaxy()
	if err != nil {
		return err
	}
	changelog, err = gog.ParseChangelog(bgr.gog.Changelog, bgr.gog.Title)
	if err != nil {
		return err
	}
	versionChanges := changelog.FindChange(version)
	if versionChanges == nil {
		return fmt.Errorf("no version information found for %q", version)
	}

	// TODO: Replace with custom errors
	t, err := bgr.git.Worktree()
	if err != nil {
		return err
	}
	var stat git.Status
	stat, err = t.Status()
	if err != nil {
		return err
	}
	if !stat.IsClean() {
		return errors.New("git working directory must be clean before adding a new version, please save your work:\n" + stat.String())
	}
	if _, err := t.Filesystem.Stat(".gitignore"); err != nil {
		gitignore, err := t.Filesystem.Create(".gitignore")
		if err != nil {
			return err
		}
		_, err = gitignore.Write([]byte(ignore))
		if err != nil {
			return err
		}
		gitignore.Close()
	}

	t.Add(".gitignore")

	wg := &sync.WaitGroup{}
	mut := &sync.Mutex{}
	errs := new(multierror.Error)
	paths := []string{}

	err = filepath.Walk(filepath.Join(bgr.BG3Path, "Data"), func(path string, info os.FileInfo, err error) error {
		rpath := strings.TrimPrefix(path, filepath.Join(bgr.BG3Path, "Data"))
		fmt.Println("bg3  ", filepath.Join(bgr.BG3Path, "Data"))
		fmt.Println("path ", path)
		fmt.Println("rpath", rpath)
		repoPath := filepath.Join(filepath.Dir(rpath), strings.TrimSuffix(filepath.Base(rpath), filepath.Ext(rpath))+strings.ReplaceAll(filepath.Ext(rpath), ".", "_"))
		fmt.Println("repopath", repoPath)
		if filepath.Ext(info.Name()) != ".pak" || info.IsDir() || strings.Contains(info.Name(), "Textures") {
			return nil
		}

		err = t.Filesystem.MkdirAll(repoPath, 0o660)
		if err != nil {
			return err
		}
		paths = append(paths, repoPath)
		divine := exec.Command(bgr.divinePath, "-g", "bg3", "-s", path, "-d", filepath.Join(bgr.gitPath, repoPath), "-a", "extract-package")
		var buf bytes.Buffer
		divine.Stderr = &buf
		divine.Stdout = &buf
		err = divine.Run()
		if err != nil {
			return fmt.Errorf("an error occurred running divine.exe: %s", buf.String())
		}
		wg.Add(1)
		go func() {
			if err := convert("-w", "-r", filepath.Join(bgr.gitPath, repoPath)); err != nil {
				mut.Lock()
				errs = multierror.Append(errs, err)
				mut.Unlock()
			}
			wg.Done()
		}()
		return nil
	})
	if err != nil {
		return err
	}
	wg.Wait()

	if errs.ErrorOrNil() != nil {
		return nil
	}
	for _, v := range paths {
		_, err = t.Add(v)
		if err != nil {
			return err
		}
	}
	_, err = t.Commit(strings.Replace(versionChanges.String(), versionChanges.Title, versionChanges.Title+"\n", 1), nil)
	if err != nil {
		return err
	}
	return nil
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
			for _, vv := range []string{`:\Program Files*\*\`, `:\app*\`, `:\`} {
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

func checkPaths() string {
	var searchPath []string
	folderName := `Baldurs Gate 3\bin\`
	if runtime.GOOS == "windows" {
		for _, v := range "cdefgh" {
			for _, vv := range []string{`:\Program Files*\*\`, `:\Program Files*\GOG Galaxy\Games\`, `:\GOG Galaxy\Games\`, `:\GOG Games\`, `:\app*\`, `:\`} {
				paths, err := filepath.Glob(filepath.Join(string(v)+vv, folderName))
				if err != nil {
					panic(err)
				}
				searchPath = append(searchPath, paths...)
			}
		}
	}
	bg3, _ := ie.LookPath("bg3.exe", strings.Join(searchPath, string(os.PathListSeparator)))
	return strings.TrimSuffix(bg3, `\bin\bg3.exe`)
}

func locateBG3() string {
	installPath := initFlags.BG3Path
	if installPath == "" {
		installPath = checkRegistry()
	}
	if installPath == "" {
		installPath = checkPaths()
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
