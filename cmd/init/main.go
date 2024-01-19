// init initializes an existing or empty directory as a gobl project
//
// Usage:
//
//	init <directory> <main cmd> [watchdirs...]
//
// If directory does not exist, it will be created.
// If any .go file in the root of the directory has a main func, init will create
// the gobl.go file with a gobl() func. This should be called by the main func to
// enable gobl functionality.
// If the main cmd directory + file does not exist, they will be created.
// Any watchdirs will be created if they do not exist.
//
// Example:
//
//	init myproject cmd/mycmd assets/ templates/
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()
	var dir string
	var cmdPath string
	var cmd string
	var watchdirs []string
	funcSig := "main"

	if len(os.Args) < 3 {
		fmt.Println("usage: init dir cmd watchdirs...")
		os.Exit(1)
	}
	dir = os.Args[1]
	cmdPath = filepath.Dir(os.Args[2])
	cmd = filepath.Base(os.Args[2])
	if len(os.Args) > 3 {
		watchdirs = os.Args[3:]
	}

	// Check if dir exists and if not, create it.
	if _, err := os.Stat(dir); err == nil {
		fmt.Printf(" - %s already exists, not creating\n", dir)
	} else if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		} else {
			fmt.Printf(" - no go.mod found, initializing with \"%s\"\n", filepath.Base(dir))
			c := exec.Command("go", "mod", "init", filepath.Base(dir))
			c.Dir = dir
			if err := c.Run(); err != nil {
				panic(err)
			}
		}
	} else {
		fmt.Println(" - go.mod already exists, skipping creation")
	}

	// Check if any root .go files have a func main entry. TODO: We could make gobl not require a main func, but rather prompt the enduser to call gobl() in their main func.
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, d := range files {
		path := filepath.Join(dir, d.Name())
		if strings.HasSuffix(path, ".go") {
			b, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			for _, line := range strings.Split(string(b), "\n") {
				if strings.HasPrefix(line, "func main") {
					fmt.Printf(" - %s already has a main function, creating non-main gobl func. Please call gobl() in your program to have gobl functionality!\n", path)
				} else {
					funcSig = "gobl"
				}
			}
		}
	}

	if _, err := os.Stat(filepath.Join(dir, "gobl.go")); err == nil {
		fmt.Println(" ! gobl.go already exists, bailing")
		os.Exit(1)
	}

	var realdirs []string

	realdirs = append(realdirs, fmt.Sprintf("\"%s/*\"", filepath.Join(cmdPath, cmd)))

	for _, d := range watchdirs {
		realdirs = append(realdirs, fmt.Sprintf("\"%s/*\"", d))
	}

	if err := os.WriteFile(filepath.Join(dir, "gobl.go"), []byte(fmt.Sprintf(template, funcSig, cmd, filepath.Join(cmdPath, cmd), strings.Join(realdirs, ", "))), 0644); err != nil {
		panic(err)
	}

	// Create cmd directory.
	if err := os.MkdirAll(filepath.Join(dir, filepath.Join(cmdPath, cmd)), 0755); err != nil {
		panic(err)
	}
	// Check if cmd file exists and if not, create it.
	if _, err := os.Stat(filepath.Join(dir, cmdPath, cmd, "main.go")); err == nil {
		fmt.Printf(" - %s already exists, not creating\n", filepath.Join(dir, cmdPath, cmd, "main.go"))
	} else if err := os.WriteFile(filepath.Join(dir, cmdPath, cmd, "main.go"), []byte("package main\n\nfunc main() {\n\t// TODO: Implement\n}\n"), 0644); err != nil {
		panic(err)
	}

	// Create stub watchdirs.
	for _, d := range watchdirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			panic(err)
		}
	}

	// Install gobl
	c := exec.Command("go", "get", "github.com/kettek/gobl")
	c.Dir = dir
	if err := c.Run(); err != nil {
		panic(err)
	}
	fmt.Println("gobl initialized in " + dir)
}

const template = `package main

import (
	"runtime"
	. "github.com/kettek/gobl"
)

func %s() {
	var exe string
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}

	runArgs := append([]interface{}{}, "./%s"+exe)

	Task("build").
		Exec("go", "build", "./%s")
	Task("run").
		Exec(runArgs...)
	Task("watch").
		Watch(%s).
		Signaler(SigQuit).
		Run("build").
		Run("run")
	Go()
}
`
