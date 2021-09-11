// +build tools

package tools

import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/rakyll/hey"
	_ "golang.org/x/tools/cmd/goimports"
)

//go:generate go install github.com/golang/mock/mockgen
//go:generate go install golang.org/x/tools/cmd/goimports
