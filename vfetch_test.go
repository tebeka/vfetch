package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func Test_isDir(t *testing.T) {
	ok, _ := isDir(".")
	if !ok {
		t.Fatalf(". is not a directory")
	}

	path := "/path/to/no/where"
	ok, _ = isDir(path)
	if ok {
		t.Fatalf("'%s' is a directory", path)
	}
}

func build(t *testing.T) {
	if err := exec.Command("go", "build").Run(); err != nil {
		t.Fatalf("error: can't build - %s", err)
	}
}

func Test_Version(t *testing.T) {
	build(t)
	out, err := exec.Command("./vfetch", "--version").Output()
	if err != nil {
		t.Fatalf("can't run with --version - %s", err)
	}

	if !strings.Contains(string(out), Version) {
		t.Fatalf("no version in output - %s", string(out))
	}
}

func copyExe(path string, t *testing.T) {
	src, err := os.Open("vfetch")
	if err != nil {
		t.Fatalf("can't open vfetch - %s", err)
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		t.Fatalf("can't create %s - %s", path, err)
	}
	defer dst.Close()
	if err = dst.Chmod(0777); err != nil {
		t.Fatalf("can't make %s executable - %s", path, err)
	}
	_, err = io.Copy(dst, src)
	if err != nil {
		t.Fatalf("can't copy vfetch to %s - %s", path, err)
	}
}

func Test_Fetch(t *testing.T) {
	build(t)

	tmpDir, err := ioutil.TempDir("", "vfetch")
	if err != nil {
		t.Fatalf("can't create temp dir - %s", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := fmt.Sprintf("%s/vfetch", tmpDir)
	copyExe(tmpFile, t)

	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("can't get $PWD - %s", err)
	}
	if err = os.Chdir(tmpDir); err != nil {
		t.Fatalf("can't chdir to %s - %s", tmpDir, err)
	}
	defer os.Chdir(curDir)

	pkg := "github.com/gorilla/mux"
	if err = exec.Command("./vfetch", pkg).Run(); err != nil {
		t.Fatalf("can't fetch %s - %s", pkg, err)
	}

	ctxDir := "vendor/github.com/gorilla/context"
	ok, err := isDir(ctxDir)
	if err != nil {
		t.Fatalf("can't verify %s - %s", ctxDir, err)
	}
	if !ok {
		t.Fatalf("%s not found", ctxDir)
	}
}
