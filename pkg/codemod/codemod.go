package codemod

import (
	"context"
)

type CodeModifier interface {
	Name() string
	Apply(ctx context.Context) error
}
