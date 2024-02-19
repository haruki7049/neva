package pkgmanager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nevalang/neva/internal/compiler"
)

func (p Manager) BuildModule(ctx context.Context, workdir string) (compiler.RawModule, error) {
	manifest, err := p.retrieveManifest(workdir)
	if err != nil {
		return compiler.RawModule{}, fmt.Errorf("retrieve manifest: %w", err)
	}

	pkgs := map[string]compiler.RawPackage{}
	if err := collectNevaFiles(workdir, pkgs); err != nil {
		return compiler.RawModule{}, fmt.Errorf("walk: %w", err)
	}

	return compiler.RawModule{
		Manifest: manifest,
		Packages: pkgs,
	}, nil
}

// collectNevaFiles recursively walks the given tree and fills given pkgs with neva files
func collectNevaFiles(rootPath string, pkgs map[string]compiler.RawPackage) error {
	if err := filepath.Walk(rootPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("filepath walk: %s: %w", filePath, err)
		}

		ext := filepath.Ext(info.Name())
		if ext != ".neva" {
			return nil
		}

		bb, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		pkgName := getPkgName(rootPath, filePath)
		if _, ok := pkgs[pkgName]; !ok {
			pkgs[pkgName] = compiler.RawPackage{}
		}

		fileName := strings.TrimSuffix(info.Name(), ext)
		pkgs[pkgName][fileName] = bb

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func getPkgName(rootPath, filePath string) string {
	dirPath := filepath.Dir(filePath)
	if dirPath == rootPath { // current directory is root directory
		ss := strings.Split(dirPath, "/")
		return ss[len(ss)-1] // return last element
	}
	return strings.TrimPrefix(dirPath, rootPath+"/")
}
