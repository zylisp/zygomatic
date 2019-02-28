package zcore

import (
	"errors"
	"fmt"
	"strings"
)

func StringFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"nsplit":  SplitStringOnNewlinesFunction,
		"split":   SplitStringFunction,
		"chomp":   StringUtilFunction,
		"trim":    StringUtilFunction,
		"println": PrintFunction,
		"print":   PrintFunction,
		"printf":  PrintFunction,
		"sprintf": PrintFunction,
		"raw2str": RawToStringFunction,
		"str2sym": Str2SymFunction,
		"sym2str": Sym2StrFunction,
		"gensym":  GensymFunction,
		"symnum":  SymnumFunction,
	}

}

func PrintFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	var str string

	switch expr := args[0].(type) {
	case *SexpStr:
		str = expr.S
	default:
		str = expr.SexpString(nil)
	}

	switch name {
	case "println":
		fmt.Println(str)
	case "print":
		fmt.Print(str)
	case "printf", "sprintf":
		if len(args) == 1 && name == "printf" {
			fmt.Printf(str)
		} else {
			ar := make([]interface{}, len(args)-1)
			for i := 0; i < len(ar); i++ {
				switch x := args[i+1].(type) {
				case *SexpInt:
					ar[i] = x.Val
				case *SexpBool:
					ar[i] = x.Val
				case *SexpFloat:
					ar[i] = x.Val
				case *SexpChar:
					ar[i] = x.Val
				case *SexpStr:
					ar[i] = x.S
				case *SexpTime:
					ar[i] = x.Tm.In(NYC)
				default:
					ar[i] = args[i+1]
				}
			}
			if name == "printf" {
				fmt.Printf(str, ar...)
			} else {
				// sprintf
				return &SexpStr{S: fmt.Sprintf(str, ar...)}, nil
			}
		}
	}

	return SexpNull, nil
}

func Str2SymFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpStr:
		return env.MakeSymbol(t.S), nil
	}
	return SexpNull, fmt.Errorf("argument must be string")
}

func GensymFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	n := len(args)
	switch {
	case n == 0:
		return env.GenSymbol("__gensym"), nil
	case n == 1:
		switch t := args[0].(type) {
		case *SexpStr:
			return env.GenSymbol(t.S), nil
		}
		return SexpNull, fmt.Errorf("argument must be string")
	default:
		return SexpNull, WrongNargs
	}
}

func SymnumFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpSymbol:
		return &SexpInt{Val: int64(t.number)}, nil
	}
	return SexpNull, fmt.Errorf("argument must be symbol")
}

func ConcatStr(str *SexpStr, rest []Sexp) (*SexpStr, error) {
	res := &SexpStr{S: str.S}
	for i, x := range rest {
		switch t := x.(type) {
		case *SexpStr:
			res.S += t.S
		case *SexpChar:
			res.S += string(t.Val)
		default:
			return &SexpStr{}, fmt.Errorf("ConcatStr error: %d-th argument (0-based) is "+
				"not a string (was %T)", i, t)
		}
	}

	return res, nil
}

func AppendStr(str *SexpStr, expr Sexp) (*SexpStr, error) {
	var chr *SexpChar
	switch t := expr.(type) {
	case *SexpChar:
		chr = t
	case *SexpStr:
		return &SexpStr{S: str.S + t.S}, nil
	default:
		return &SexpStr{}, errors.New("second argument is not a char")
	}

	return &SexpStr{S: str.S + string(chr.Val)}, nil
}

func StringUtilFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	var s string
	switch str := args[0].(type) {
	case *SexpStr:
		s = str.S
	default:
		return SexpNull, fmt.Errorf("string required, got %T", s)
	}

	switch name {
	case "chomp":
		n := len(s)
		if n > 0 && s[n-1] == '\n' {
			return &SexpStr{S: s[:n-1]}, nil
		}
		return &SexpStr{S: s}, nil
	case "trim":
		return &SexpStr{S: strings.TrimSpace(s)}, nil
	}
	return SexpNull, fmt.Errorf("unrecognized command '%s'", name)
}
