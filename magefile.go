// +build mage

package main

import (
	"fmt"
	"runtime"

	"github.com/magefile/mage/sh"
)

// Build will build the basic magefile
func Build() error {
	ldf, err := flags()
	if err != nil {
		return err
	}
	fmt.Println("go build -ldflags '" + ldf + "'")
	err = sh.RunV("go", "build", "-ldflags", ldf)
	if err != nil {
		fmt.Println(`You may need to install libpcap.

OS X: brew install libpcap

Linux: apt-get install libpcap-dev

`)
	}
	return err
}

// Release will build and then tar a release
func Release() error {
	err := sh.RunV("tar", "-czvf", "find3-cli-scanner-"+runtime.GOOS+".tar.gz", "find3-cli-scanner", "README.md")
	if err == nil {
		fmt.Println("built", "find3-cli-scanner-"+runtime.GOOS+".tar.gz")
	}
	return err
}

var Default = Build

func flags() (string, error) {
	tag := tag()
	if tag == "" {
		tag = "dev"
	}
	return fmt.Sprintf(`-s -w -X main.version=%s`, tag), nil
}

// tag returns the git tag for the current branch or "" if none.
func tag() string {
	s, _ := sh.Output("git", "describe", "--tags")
	return s
}

// hash returns the git hash for the current repo or "" if none.
func hash() string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return hash
}
