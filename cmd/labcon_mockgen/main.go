package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func names(path string) (string, string) {
	dir := filepath.Dir(filepath.Dir(path))
	pkg := fmt.Sprintf("%s_mock", filepath.Base(filepath.Dir(path)))
	base := strings.Replace(filepath.Base(path), "_iface", "", 1)
	return filepath.Join(dir, pkg, base), pkg
}

func run() error {
	srcs := []string{}

	if err := filepath.WalkDir("cmd/labcon/app", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Delete old mock files.
		if strings.HasSuffix(path, "_mock") {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
			log.Printf("delete %s", path)
			return filepath.SkipDir
		}

		if !d.IsDir() && strings.HasSuffix(path, "_iface.go") {
			srcs = append(srcs, path)
		}

		return nil
	}); err != nil {
		return err
	}

	for _, src := range srcs {
		dst, pkg := names(src)
		log.Printf("generate %s -> %s", src, dst)
		if err := exec.Command(
			"mockgen",
			"-source", src,
			"-destination", dst,
			"-package", pkg,
		).Run(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	log.Println("Finish mockup generation")
}
