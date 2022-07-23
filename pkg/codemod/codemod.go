package codemod

import (
	"context"
)

type CodeModifier interface {
	Apply(ctx context.Context) error
}
