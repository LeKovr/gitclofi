package main

// mdclofil - markdown cloner & filter

import (
	//        "fmt"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	//  "bytes"
	"encoding/json"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"github.com/LeKovr/go-kit/config"
	"github.com/LeKovr/go-kit/slogger"

	//	"github.com/LeKovr/go-kit/ver"

	"github.com/LeKovr/gitclofi/static"
)

// Config holds all config vars.
type Config struct {
	Sources  string `default:"sources.yml" description:"Sourcer sonfig file"         long:"sources"`
	Root     string `default:""            description:"Templates directory"         long:"root"`
	Clone    string `default:"var"         description:"Clone destination directory" long:"clone"`
	Dest     string `default:"dest"        description:"Destination directory"       long:"dest"`
	TmplPath string `long:"tmpl" default:"tmpl" description:"Template files subdirectory"`
	TmplExt  string `long:"tmpl_ext" default:".gohtml" description:"Template extention"`

	Logger slogger.Config `env-namespace:"LOG" group:"Logging Options"      namespace:"log"`
}

type File struct {
	Name   string
	Dest   string
	Filter string
	//    Header string
	Template string // has access to all vars
	Vars     map[string]string
}

type Repo struct {
	Name  string
	URL   string
	Files []File
}

const (
	application = "gitclofi"
)

var (
	// App version, actual value will be set at build time.
	version = "0.0-dev"

	// Repository address, actual value will be set at build time.
	repo = "repo.git"
)

// Run app and exit via given exitFunc.
func Run(ctx context.Context, exitFunc func(code int)) {
	// Load config
	var cfg Config
	err := config.Open(&cfg)
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered panic", "err", r)
		}
		config.Close(err, exitFunc)
	}()
	if err != nil {
		return
	}
	err = slogger.Setup(cfg.Logger, nil)
	if err != nil {
		return
	}
	slog.Info(application, "version", version)
	//	go ver.Check(repo, version)

	var root, tfs fs.FS
	if root, err = static.New(cfg.Root); err != nil {
		return
	}
	if tfs, err = fs.Sub(root, cfg.TmplPath); err != nil {
		return
	}

	fm := sprig.FuncMap()
	//parser.AddFuncs(fm)
	var tmpl *template.Template
	tmpl, err = template.New("").Funcs(fm).ParseFS(tfs, "*"+cfg.TmplExt)

	EnsureDir(filepath.Join(cfg.Clone, ".nogit"))

	repos := []Repo{}

	var data []byte
	data, err = os.ReadFile(cfg.Sources)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &repos)
	if err != nil {
		return
	}
	// out(repos)
	for _, repo := range repos {
		// check repo exists
		destDirName := filepath.Base(repo.URL)
		destDir := filepath.Join(cfg.Clone, destDirName)
		execDir := cfg.Clone
		cmdArgs := []string{}
		if _, serr := os.Stat(destDir); serr != nil {
			slog.Debug("Will clone", "repo", repo.URL, "dest", destDir)
			cmdArgs = append(cmdArgs, "clone", repo.URL+".git")
		} else {
			slog.Debug("Will pull", "repo", repo.URL)
			cmdArgs = append(cmdArgs, "pull")
			execDir = filepath.Join(execDir, destDirName)
		}
		cmd := exec.Command("git", cmdArgs...)
		cmd.Dir = execDir
		err = cmd.Run()
		if err != nil {
			// something went wrong
			return
		}

		/*
					        "Name": "README.md",
			        "Dest": "ru/dcape.md",
			        "Filter": "tail -n +2",
		*/

		for _, file := range repo.Files {
			dst := filepath.Join(cfg.Dest, file.Dest)
			err := EnsureDir(dst)
			if err != nil {
				return
			}
			dstFile, err := os.Create(dst)
			if err != nil {
				slog.Error("create file", "err", err)
				return
			}
			if file.Template != "" {
				args := file.Vars
				args["name"] = repo.Name
				args["url"] = repo.URL
				if err = tmpl.ExecuteTemplate(dstFile, "header"+cfg.TmplExt, args); err != nil {
					slog.Error("exec tmpl", "err", err)
					dstFile.Close()
					continue
				}
			}
			// filter file
			srcFile, err := os.Open(filepath.Join(destDir, file.Name))
			defer srcFile.Close()
			if file.Filter != "" {
				slog.Debug("Use filter", "cmd", file.Filter)
				a := strings.Split(file.Filter, " ")
				execCmd(context.Background(), time.Duration(time.Second), a, srcFile, dstFile)
			} else {
				io.Copy(dstFile, srcFile)
			}
			//			cmdArgs = append(cmdArgs, ">>", dst)
			/*
				cmd := exec.Command("cat", cmdArgs...)
				err = cmd.Run()
				if err != nil {
					slog.Debug("Content write", "err", err)
					// something went wrong
					return
				}
			*/
		}
	}
}

func out(data interface{}) {
	//var prettyJSON bytes.Buffer
	//error := json.Indent(&prettyJSON, data, "", "\t")
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println("JSON parse error: ", err)
		return
	}
	fmt.Println(string(b))
}

func EnsureDir(fileName string) error {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			return merr
		}
	}
	return nil
}

// execCmd выполняет в ОС команду cmd.
func execCmd(ctx context.Context, wait time.Duration, cmd []string, src io.Reader, dest io.Writer) {
	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	e := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	e.Stderr = os.Stderr
	e.Stdin = src
	e.Stdout = dest
	err := e.Run()
	//	slog.InfoContext(ctx, "execCmd results", "cmd", cmd, "out", out, "err", err)
	slog.InfoContext(ctx, "execCmd results", "cmd", cmd, "err", err)
}
