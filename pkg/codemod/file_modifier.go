package codemod

import (
	"os"

	"github.com/gocopper/copper/cerrors"
)

func newFileModifier(fp string, cm *CodeModder) *FileModifier {
	contents, err := os.ReadFile(fp)
	if err != nil {
		return &FileModifier{
			err: cerrors.New(err, "failed to read file", map[string]interface{}{
				"path": fp,
			}),
		}
	}

	return &FileModifier{
		path:     fp,
		contents: string(contents),
		err:      nil,
		cm:       cm,
	}
}

type FileModifier struct {
	path     string
	contents string
	err      error
	cm       *CodeModder
}

func (m *FileModifier) Apply(fn ...FileModifierFn) *FileModifier {
	var ret = m
	for i := range fn {
		ret = m.apply(fn[i])
	}
	return ret
}

func (m *FileModifier) Close() *CodeModder {
	if m.err != nil {
		return &CodeModder{
			err: m.err,
		}
	}

	err := os.WriteFile(m.path, []byte(m.contents), 0644)
	if err != nil {
		return &CodeModder{
			err: m.err,
		}
	}

	return m.cm
}

func (m *FileModifier) CloseAndOpen(fp string) *FileModifier {
	return m.Close().OpenFile(fp)
}

func (m *FileModifier) CloseAndDone() error {
	return m.Close().Done()
}

func (m *FileModifier) apply(fn FileModifierFn) *FileModifier {
	if m.err != nil {
		return m
	}

	d, err := fn(m.contents, m.cm.data)
	if err != nil {
		return &FileModifier{
			err: cerrors.New(err, "failed to apply modifier", nil),
		}
	}

	m.contents = d

	return m
}
