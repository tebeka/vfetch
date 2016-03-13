/* vfetch gets a package with depedencies to vendor directory

vfetch uses "go get" to get the package to a temporary directory and then uses
"rsync" to copy the content of "src" directory to "vendor".
*/
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var (
	Version = "0.1.0"
)

// die prints error message and aborts the program
func die(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}

// isDir return true if path exists and is a directory
func isDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		return fi.Mode().IsDir(), err
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func main() {
	var showVersion, verbose bool

	flag.BoolVar(&verbose, "verbose", false, "emit more noise")
	flag.BoolVar(&showVersion, "version", false, "show version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s PACKAGE\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("vfetch %s\n", Version)
		os.Exit(0)
	}

	info := func(format string, v ...interface{}) {}
	if verbose {
		info = log.Printf
	}

	vendor := "vendor"

	if flag.NArg() != 1 {
		die("wrong number of arguments")
	}
	pkg := flag.Arg(0)

	exists, err := isDir(vendor)
	if err != nil {
		die("can't find if 'vendor' exists - %s", err)
	}
	if !exists {
		info("creating %s", vendor)
		err = os.Mkdir(vendor, 0755)
		if err != nil {
			die("can't create %s - %s", vendor, err)
		}
	}

	gopath, err := ioutil.TempDir("", "vfetch")
	if err != nil {
		die("can't create temp dir - %s", err)
	}
	defer os.RemoveAll(gopath)
	info("GOPATH = %s", gopath)

	oldPath := os.Getenv("GOPATH")
	if err = os.Setenv("GOPATH", gopath); err != nil {
		die("can't set GOPATH - %s", err)
	}
	// No sure this is needed but play nice
	defer func() {
		os.Setenv("GOPATH", oldPath)
	}()

	info("go getting %s", pkg)
	cmd := exec.Command("go", "get", pkg)
	if err = cmd.Run(); err != nil {
		die("can't 'go get %s' - %s", pkg, err)
	}

	// The trailing / is important
	src := fmt.Sprintf("%s/src/", gopath)

	info("rsync from %s to %s", src, vendor)
	// TODO: Find pure Go rsync package
	cmd = exec.Command(
		"rsync", "-a",
		"--exclude", ".git",
		"--exclude", ".hg",
		"--exclude", ".svn",
		"--exclude", ".bzr",
		src, vendor,
	)
	if err = cmd.Run(); err != nil {
		die("can't rsync from %s to vendor - %s", src, err)
	}
}
