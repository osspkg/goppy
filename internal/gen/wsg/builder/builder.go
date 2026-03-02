package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"go.osspkg.com/bb"
	"go.osspkg.com/gogen/golang"
	"go.osspkg.com/gogen/types"
	"go.osspkg.com/goppy/v3/console"
	wsgt "go.osspkg.com/goppy/v3/internal/gen/wsg/types"
)

type Builder struct {
	Out   string
	IFace map[string]struct{}
	Files []wsgt.File
}

func (b *Builder) filterObjects(objs []wsgt.Object) []wsgt.Object {
	if len(b.IFace) == 0 {
		return objs
	}
	result := make([]wsgt.Object, 0, len(objs))
	for _, obj := range objs {
		if _, ok := b.IFace[obj.Name]; ok {
			result = append(result, obj)
		}
	}

	return result
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

	for _, file := range b.Files {
		file.Objects = b.filterObjects(file.Objects)
	}

	for _, file := range b.Files {
		for _, obj := range file.Objects {
			tags, ok := obj.Tags[TagModule]
			if !ok {
				continue
			}

			for _, tag := range tags {
				mod, ok := wsgt.Resolve(wsgt.IFace + tag)
				if !ok {
					continue
				}

				if err := mod.Build(wsgt.Ctx{
					W:       b,
					PkgName: pkgName,
					File:    file,
					Object:  obj,
					Tags:    obj.Tags,
				}); err != nil {
					return fmt.Errorf("build module `%s`, file `%s`: %w",
						mod.Name(), file.FilePath, err)
				}
			}
		}
	}

	//return errors.Queue(
	//	func() error { return b.buildInterface(pkgName) },
	//	func() error { return b.WriteFile("transport.go", module.TypeTransport(pkgName)) },
	//)

	return nil
}
