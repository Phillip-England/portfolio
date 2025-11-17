package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/phillip-england/marki/mdfile"
	"github.com/phillip-england/wherr"
	"github.com/phillip-england/whip"
)

type Run struct {
	Src          string
	SrcIsFile    bool
	Out          string
	Theme        string
	HasWatchFlag bool
}

func NewRun(cli *whip.Cli) (whip.Cmd, error) {
	src, err := cli.ArgGetByPositionForce(2, "missing <SOURCE> in 'marki convert'")
	if err != nil {
		return Run{}, wherr.Consume(wherr.Here(), err, "")
	}
	srcIsFile := whip.IsFile(src)
	srcIsDir := whip.IsDir(src)
	if !srcIsDir && !srcIsFile {
		return Run{}, wherr.Err(wherr.Here(), "<SOURCE> in 'marki convert' must be either a file or dir on your system")
	}
	if srcIsFile {
		if !whip.FileExists(src) {
			return Run{}, wherr.Err(wherr.Here(), "file %s does not exist", src)
		}
	} else {
		if !whip.DirExists(src) {
			return Run{}, wherr.Err(wherr.Here(), "dir %s does not exist", src)
		}
	}
	out, err := cli.ArgGetByPositionForce(3, "missing <OUT> in 'marki convert'")
	if err != nil {
		return Run{}, wherr.Consume(wherr.Here(), err, "")
	}
	if srcIsFile {
		if !whip.IsFile(out) {
			return Run{}, wherr.Err(wherr.Here(), "<OUT> must be a properly formatted file path if <SOURCE> is a file")
		}
	} else {
		if !whip.IsDir(out) {
			return Run{}, wherr.Err(wherr.Here(), "<OUT> must be a properly formatted dir path if <SOURCE> is a dir")
		}
	}
	theme, err := cli.ArgGetByPositionForce(4, "<THEME> ")
	if err != nil {
		theme = "dracula"
	}
	err = isValidTheme(theme)
	if err != nil {
		return Run{}, wherr.Consume(wherr.Here(), err, "")
	}
	hasWatchFlag := cli.FlagExists("--watch")
	return Run{
		Src:          src,
		SrcIsFile:    srcIsFile,
		Out:          out,
		Theme:        theme,
		HasWatchFlag: hasWatchFlag,
	}, nil
}

func (cmd Run) Execute(app *whip.Cli) error {
	if cmd.SrcIsFile {
		err := handleFile(cmd, app)
		if err != nil {
			return wherr.Consume(wherr.Here(), err, "")
		}
		return nil
	}
	err := handleDir(cmd, app)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	return nil
}

func handleFile(cmd Run, app *whip.Cli) error {
	mdFile, err := mdfile.NewMarkdownFileFromPath(cmd.Src, cmd.Theme)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	err = mdfile.SaveMarkdownHtmlToDisk(mdFile, cmd.Out)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	if !cmd.HasWatchFlag {
		return nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	defer watcher.Close()
	err = watcher.Add(cmd.Src)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	errCh := make(chan error)
	fmt.Printf("👁️: watching %s\n", cmd.Src)
	triggerCh := make(chan bool)

	go func() {
		var timer *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(100*time.Millisecond, func() {
						triggerCh <- true
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("🚨: watcher error: %s\n", err.Error())
			case <-triggerCh:
				fmt.Printf("📝: writing to %s\n", cmd.Out)
				if err := convertAndSaveFile(cmd.Src, cmd.Theme, cmd.Out); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-make(chan struct{}):
		fmt.Println("👋 bye-bye!")
	}
	return nil
}

func handleDir(cmd Run, app *whip.Cli) error {
	err := convertAndSaveDir(cmd.Src, cmd.Out, cmd.Theme)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	if !cmd.HasWatchFlag {
		return nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	defer watcher.Close()
	err = watcher.Add(cmd.Src)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}

	err = filepath.WalkDir(cmd.Src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return wherr.Consume(wherr.Here(), err, "")
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		err = watcher.Add(path)
		if err != nil {
			return wherr.Consume(wherr.Here(), err, "")
		}
		return nil
	})
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	errCh := make(chan error)
	fmt.Printf("👁️: watching %s\n", cmd.Src)
	triggerCh := make(chan bool)
	go func() {
		var timer *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(100*time.Millisecond, func() {
						triggerCh <- true
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("🚨: watcher error: %s\n", err.Error())
			case <-triggerCh:
				fmt.Printf("📝: writing to %s\n", cmd.Out)
				if err := convertAndSaveDir(cmd.Src, cmd.Out, cmd.Theme); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-make(chan struct{}):
		fmt.Println("👋 bye-bye!")
	}
	return nil
}

func convertAndSaveDir(inDir string, outDir string, theme string) error {
	err := filepath.WalkDir(inDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return wherr.Consume(wherr.Here(), err, "")
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		relPath, err := filepath.Rel(inDir, path)
		if err != nil {
			return wherr.Consume(wherr.Here(), err, "")
		}
		outPath := filepath.Join(outDir, relPath)
		outPath = strings.TrimSuffix(outPath, ".md")
		outPath = outPath + ".html"
		err = convertAndSaveFile(path, theme, outPath)
		if err != nil {
			return wherr.Consume(wherr.Here(), err, "")
		}
		return nil
	})
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	return nil
}

func convertAndSaveFile(mdFilePath string, theme string, saveTo string) error {
	mdFile, err := mdfile.NewMarkdownFileFromPath(mdFilePath, theme)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "failed to load markdown")
	}
	err = mdfile.SaveMarkdownHtmlToDisk(mdFile, saveTo)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "failed to save html")
	}
	return nil
}
