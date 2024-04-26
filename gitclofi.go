package gitclofi

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
	"time"

	//  "bytes"
	"encoding/json"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
	//	"github.com/LeKovr/go-kit/ver"
)

// Config holds all config vars.
type Config struct {
	Sources  string `default:".gitclofi.yml" description:"Sourcer sonfig file"                         long:"sources"`
	Root     string `default:""              description:"Templates directory (use embedded if empty)" long:"root"`
	Clone    string `default:"var"           description:"Clone destination directory"                 long:"clone"`
	Dest     string `default:"dest"          description:"Destination directory"                       long:"dest"`
	TmplPath string `default:"tmpl"          description:"Template files subdirectory"                 long:"tmpl"`
	TmplExt  string `default:".gohtml"       description:"Template extention"                          long:"tmpl_ext"`
}

// File holds filtered file atributes.
type File struct {
	Name   string
	Dest   string
	Filter string
	//    Header string
	Template string // has access to all vars
	Vars     map[string]string
}

// Repo holds repository attributes.
type Repo struct {
	Name  string
	URL   string
	Files []File
}

// Service hold service attributes.
type Service struct {
	config Config
	tmplFS fs.FS
}

// New returns Service object.
func New(cfg Config, tmplFS fs.FS) *Service {
	return &Service{cfg, tmplFS}
}

// Run does the whole job.
func (srv *Service) Run(ctx context.Context) error {

	fm := sprig.FuncMap()
	//parser.AddFuncs(fm)
	tmpl, err := template.New("").Funcs(fm).ParseFS(srv.tmplFS, "*"+srv.config.TmplExt)

	ensureDir(filepath.Join(srv.config.Clone, ".nogit"))

	repos := []Repo{}

	var data []byte
	data, err = os.ReadFile(srv.config.Sources)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &repos)
	if err != nil {
		return err
	}
	// out(repos)
	for _, repo := range repos {
		err = updateRepo(repo.URL, srv.config.Clone)
		if err != nil {
			// something went wrong
			return err
		}

		/*
					        "Name": "README.md",
			        "Dest": "ru/dcape.md",
			        "Filter": "tail -n +2",
		*/

		for _, file := range repo.Files {
			dst := filepath.Join(srv.config.Dest, file.Dest)
			err := ensureDir(dst)
			if err != nil {
				return err
			}
			dstFile, err := os.Create(dst)
			if err != nil {
				slog.Error("create file", "err", err)
				return err
			}
			if file.Template != "" {
				args := file.Vars
				args["name"] = repo.Name
				args["url"] = repo.URL
				if err = tmpl.ExecuteTemplate(dstFile, "header"+srv.config.TmplExt, args); err != nil {
					slog.Error("exec tmpl", "err", err)
					dstFile.Close()
					continue
				}
			}
			// filter file

			destDir := filepath.Join(srv.config.Clone, filepath.Base(repo.URL))
			srcFile, err := os.Open(filepath.Join(destDir, file.Name))
			defer srcFile.Close()
			if file.Filter != "" {
				slog.Debug("Use filter", "cmd", file.Filter)
				err = execCmd(context.Background(), time.Duration(time.Second), file.Filter, srcFile, dstFile)
			} else {
				_, err = io.Copy(dstFile, srcFile)
			}
			if err != nil {
				slog.Error("fiter file", "err", err)
			}
		}
	}
	return nil
}

func updateRepo(url, clonePath string) error {
	// check repo exists
	destDirName := filepath.Base(url)
	destDir := filepath.Join(clonePath, destDirName)
	execDir := clonePath
	cmdArgs := []string{}
	if _, serr := os.Stat(destDir); serr != nil {
		slog.Debug("Will clone", "repo", url, "dest", destDir)
		cmdArgs = append(cmdArgs, "clone", url+".git")
	} else {
		slog.Debug("Will pull", "repo", url)
		return nil
		//	cmdArgs = append(cmdArgs, "pull")
		//	execDir = filepath.Join(execDir, destDirName)
	}
	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = execDir
	return cmd.Run()
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

func ensureDir(fileName string) error {
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
func execCmd(ctx context.Context, wait time.Duration, cmd string, src io.Reader, dest io.Writer) error {
	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	e := exec.CommandContext(ctx, "bash", "-c", cmd)
	e.Stderr = os.Stderr
	e.Stdin = src
	e.Stdout = dest
	return e.Run()
}
