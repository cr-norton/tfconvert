package codegen

import "github.com/cr-norton/tfconvert/pkg/types"

type Stack interface {
	Lookup(id string) *types.Resource
}
