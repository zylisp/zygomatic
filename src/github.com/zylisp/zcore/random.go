package zcore

import (
	"math/rand"
	"time"
)

var defaultRand = rand.New(rand.NewSource(time.Now().Unix()))

func RandomFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"random":  RandomFunction,
	}
}

func RandomFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	return &SexpFloat{Val: defaultRand.Float64()}, nil
}

func (env *Zlisp) ImportRandom() {
	env.AddFunction("random", RandomFunction)
}
