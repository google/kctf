package controller

import (
	"github.com/google/kctf/pkg/controller/challenge"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, challenge.Add)
}
