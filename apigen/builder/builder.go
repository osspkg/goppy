package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.osspkg.com/bb"
	"go.osspkg.com/gogen/golang"
	"go.osspkg.com/gogen/types"
	at "go.osspkg.com/goppy/v3/apigen/types"
	"go.osspkg.com/goppy/v3/console"
)

type Builder struct {
	Out   string
	Mods  []string
	Pool  []string
	IFace map[string]struct{}
	Files []at.File
}

func (b *Builder) filterObjects(objs []at.Face) []at.Face {
	if len(b.IFace) == 0 {
		return objs
	}

	result := make([]at.Face, 0, len(objs))
	for _, obj := range objs {
		if _, ok := b.IFace[strings.ToLower(obj.Name)]; ok {
			result = append(result, obj)
		}
	}

	return result
}

func (b *Builder) getWorkFiles() []at.File {
	workFiles := make([]at.File, 0, len(b.Files))

	for _, file := range b.Files {
		file.Faces = b.filterObjects(file.Faces)
		if len(file.Faces) > 0 {
			workFiles = append(workFiles, file)
		}
	}

	slices.SortFunc(workFiles, func(a, b at.File) int {
		return strings.Compare(a.FilePath, b.FilePath)
	})

	return workFiles
}

func (b *Builder) WriteFile(fileName string, tok types.Token) error {
	buf := bb.New(1024)
	if err := golang.Render(buf, tok); err != nil {
		return err
	}

	fullPath := b.Out + "/" + fileName
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0766); err != nil {
		return fmt.Errorf("mkdir %q: %v", dir, err)
	}

	console.Debugf("Writing file %s", fullPath)
	return os.WriteFile(fullPath, buf.Bytes(), 0666)
}

func (b *Builder) Build() error {
	//golang.SetRawMode()

	pkgName := filepath.Base(b.Out)
	files := b.getWorkFiles()

	for _, name := range b.Mods {
		mod, ok := at.Resolve[at.GlobalModule](name)
		if !ok {
			continue
		}

		err := mod.Build(b, at.GlobalMeta{PkgName: pkgName, Pool: b.Pool}, files)
		if err != nil {
			return fmt.Errorf("build module %q: %v", name, err)
		}
	}

	return nil
}
