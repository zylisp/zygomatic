package zcore

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
)

// CoreFunctions returns all of the core logic
func CoreFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"!=":          CompareFunction,
		"&":           AddressOfFunction,
		"*":           PointerOrNumericFunction,
		"**":          NumericFunction,
		"+":           NumericFunction,
		"-":           NumericFunction,
		"->":          ThreadMapFunction,
		".":           DotFunction,
		"/":           NumericFunction,
		//":":          ColonAccessFunction,
		":=":          AssignmentFunction,
		"<":           CompareFunction,
		"<=":          CompareFunction,
		"=":           AssignmentFunction,
		"==":          CompareFunction,
		">":           CompareFunction,
		">=":          CompareFunction,
		"aget":        ArrayAccessFunction,
		"append":      AppendFunction,
		"appendslice": AppendFunction,
		"apply":       ApplyFunction,
		"array":       ConstructorFunction,
		"array?":      TypeQueryFunction,
		"arrayidx":    ArrayIndexFunction,
		"aset":        ArrayAccessFunction,
		"asUint64":    AsUint64Function,
		"bitAnd":      BitwiseFunction,
		"bitNot":      ComplementFunction,
		"bitOr":       BitwiseFunction,
		"bitXor":      BitwiseFunction,
		"car":         FirstFunction,
		"cdr":         RestFunction,
		"char?":       TypeQueryFunction,
		"concat":      ConcatFunction,
		"cons":        ConsFunction,
		"defined?":    DefinedFunction,
		"deref":       DerefFunction,
		"derefSet":    DerefFunction,
		"empty?":      TypeQueryFunction,
		"field":       ConstructorFunction,
		"fieldls":     GoFieldListFunction,
		"first":       FirstFunction,
		"flatten":     FlattenToWordsFunction,
		"float?":      TypeQueryFunction,
		"func?":       TypeQueryFunction,
		"GOOS":        GOOSFunction,
		"hash":        ConstructorFunction,
		"hash?":       TypeQueryFunction,
		"hashidx":     HashIndexFunction,
		"hdel":        HashAccessFunction,
		"hget":        GenericAccessFunction, // handles arrays or hashes
		"hpair":       GenericHpairFunction,
		"hset":        HashAccessFunction,
		"int?":        TypeQueryFunction,
		"isnan":       IsNaNFunction,
		"isNaN":       IsNaNFunction,
		"joinsym":     JoinSymFunction,
		"keys":        HashAccessFunction,
		"len":         LenFunction,
		"list":        ConstructorFunction,
		"list?":       TypeQueryFunction,
		"makeArray":   MakeArrayFunction,
		"map":         MapFunction,
		"mod":         BinaryIntFunction,
		"not":         NotFunction,
		"null?":       TypeQueryFunction,
		"number?":     TypeQueryFunction,
		"pretty":      SetPrettyPrintFlag,
		"quotelist":   QuoteListFunction,
		"raw":         ConstructorFunction,
		"read":        ReadFunction,
		"rest":        RestFunction,
		"second":      SecondFunction,
		"sget":        SgetFunction,
		"slice":       SliceFunction,
		"sll":         BinaryIntFunction,
		"sra":         BinaryIntFunction,
		"srl":         BinaryIntFunction,
		"stop":        StopFunction,
		"str":         StringifyFunction,
		"string?":     TypeQueryFunction,
		"struct":      ConstructorFunction,
		"symbol?":     TypeQueryFunction,
		"type?":       TypeQueryFunction,
		"zero?":       TypeQueryFunction,
	}
}

func CompareFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	res, err := env.Compare(args[0], args[1])
	if err != nil {
		return SexpNull, err
	}

	if res > 1 {
		//fmt.Printf("CompareFunction, res = %v\n", res)
		// 2 => one NaN found
		// 3 => two NaN found
		// NaN != NaN needs to return true.
		// NaN != 3.0 needs to return true.
		if name == "!=" {
			return &SexpBool{Val: true}, nil
		}
		return &SexpBool{Val: false}, nil
	}

	cond := false
	switch name {
	case "<":
		cond = res < 0
	case ">":
		cond = res > 0
	case "<=":
		cond = res <= 0
	case ">=":
		cond = res >= 0
	case "==":
		cond = res == 0
	case "!=":
		cond = res != 0
	}

	return &SexpBool{Val: cond}, nil
}

func BinaryIntFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	var op IntegerOp
	switch name {
	case "sll":
		op = ShiftLeft
	case "sra":
		op = ShiftRightArith
	case "srl":
		op = ShiftRightLog
	case "mod":
		op = Modulo
	}

	return IntegerDo(op, args[0], args[1])
}

func PointerOrNumericFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	n := len(args)
	if n == 0 {
		return SexpNull, WrongNargs
	}
	if n >= 2 {
		return NumericFunction(env, name, args)
	}
	return PointerToFunction(env, name, args)
}

// "." dot operator
func DotFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	P("in DotFunction(), name='%v', args[0] = '%v', args[1]= '%v'",
		name,
		args[0].SexpString(nil),
		args[1].SexpString(nil))
	return SexpNull, nil
	/*
		var ret Sexp = SexpNull
		var err error
		lenpath := len(path)

		if lenpath == 1 && setVal != nil {
			// single path element set, bind it now.
			a := path[0][1:] // strip off the dot
			asym := env.MakeSymbol(a)

			// check conflict
			//Q("asym = %#v\n", asym)
			builtin, typ := env.IsBuiltinSym(asym)
			if builtin {
				return SexpNull, fmt.Errorf("'%s' is a %s, cannot assign to it with dot-symbol", asym.name, typ)
			}

			err := env.LexicalBindSymbol(asym, *setVal)
			if err != nil {
				return SexpNull, err
			}
			return *setVal, nil
		}

		// handle multiple paths that index into hashes after the
		// the first

		key := path[0][1:] // strip off the dot
		//Q("\n in DotGetSetHelper(), looking up '%s'\n", key)
		ret, err, _ = env.LexicalLookupSymbol(env.MakeSymbol(key), false)
		if err != nil {
			Q("\n in DotGetSetHelper(), '%s' not found\n", key)
			return SexpNull, err
		}
		return ret, err
	*/
}

func NumericFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}
	var err error
	args, err = env.SubstituteRHS(args)
	if err != nil {
		return SexpNull, err
	}

	accum := args[0]
	var op NumericOp
	switch name {
	case "+":
		op = Add
	case "-":
		op = Sub
	case "*":
		op = Mult
	case "/":
		op = Div
	case "**":
		op = Pow
	}

	for _, expr := range args[1:] {
		accum, err = NumericDo(op, accum, expr)
		if err != nil {
			return SexpNull, err
		}
	}
	return accum, nil
}

func BitwiseFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	var op IntegerOp
	switch name {
	case "bitAnd":
		op = BitAnd
	case "bitOr":
		op = BitOr
	case "bitXor":
		op = BitXor
	}

	accum := args[0]
	var err error

	for _, expr := range args[1:] {
		accum, err = IntegerDo(op, accum, expr)
		if err != nil {
			return SexpNull, err
		}
	}
	return accum, nil
}

func ComplementFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpInt:
		return &SexpInt{Val: ^t.Val}, nil
	case *SexpChar:
		return &SexpChar{Val: ^t.Val}, nil
	}

	return SexpNull, fmt.Errorf("Argument to bitNot should be integer")
}

func ConsFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	return Cons(args[0], args[1]), nil
}

func FirstFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch expr := args[0].(type) {
	case *SexpPair:
		return expr.Head, nil
	case *SexpArray:
		if len(expr.Val) > 0 {
			return expr.Val[0], nil
		}
		return SexpNull, fmt.Errorf("first called on empty array")
	}
	return SexpNull, WrongType
}

func RestFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch expr := args[0].(type) {
	case *SexpPair:
		return expr.Tail, nil
	case *SexpArray:
		if len(expr.Val) == 0 {
			return expr, nil
		}
		return &SexpArray{Val: expr.Val[1:], Env: env, Typ: expr.Typ}, nil
	case *SexpSentinel:
		if expr == SexpNull {
			return SexpNull, nil
		}
	}

	return SexpNull, WrongType
}

func SecondFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch expr := args[0].(type) {
	case *SexpPair:
		tail := expr.Tail
		switch p := tail.(type) {
		case *SexpPair:
			return p.Head, nil
		}
		return SexpNull, fmt.Errorf("list too small for second")
	case *SexpArray:
		if len(expr.Val) >= 2 {
			return expr.Val[1], nil
		}
		return SexpNull, fmt.Errorf("array too small for second")
	}

	return SexpNull, WrongType
}

func ArrayAccessFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg < 2 || narg > 3 {
		return SexpNull, WrongNargs
	}

	var arr *SexpArray
	switch t := args[0].(type) {
	case *SexpArray:
		arr = t
	default:
		return SexpNull, fmt.Errorf("First argument of aget must be array")
	}

	var i int
	switch t := args[1].(type) {
	case *SexpInt:
		i = int(t.Val)
	case *SexpChar:
		i = int(t.Val)
	default:
		// can we evaluate it?
		res, err := EvalFunction(env, "eval-aget-index", []Sexp{args[1]})
		if err != nil {
			return SexpNull, fmt.Errorf("error during eval of "+
				"array-access position argument: %s", err)
		}
		switch j := res.(type) {
		case *SexpInt:
			i = int(j.Val)
		default:
			return SexpNull, fmt.Errorf("Second argument of aget could not be evaluated to integer; got j = '%#v'/type = %T", j, j)
		}
	}

	switch name {
	case "hget":
		fallthrough
	case "aget":
		if i < 0 || i >= len(arr.Val) {
			// out of bounds -- do we have a default?
			if narg == 3 {
				return args[2], nil
			}
			return SexpNull, fmt.Errorf("Array index out of bounds")
		}
		return arr.Val[i], nil
	case "aset":
		if len(args) != 3 {
			return SexpNull, WrongNargs
		}
		if i < 0 || i >= len(arr.Val) {
			return SexpNull, fmt.Errorf("Array index out of bounds")
		}
		arr.Val[i] = args[2]
	}
	return SexpNull, nil
}

func SgetFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	var str *SexpStr
	switch t := args[0].(type) {
	case *SexpStr:
		str = t
	default:
		return SexpNull, fmt.Errorf("First argument of sget must be string")
	}

	var i int
	switch t := args[1].(type) {
	case *SexpInt:
		i = int(t.Val)
	case *SexpChar:
		i = int(t.Val)
	default:
		return SexpNull, fmt.Errorf("Second argument of sget must be integer")
	}

	return &SexpChar{Val: rune(str.S[i])}, nil
}

func HashAccessFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 || len(args) > 3 {
		return SexpNull, WrongNargs
	}

	// handle *SexpSelector
	container := args[0]
	var err error
	if ptr, isPtrLike := container.(Selector); isPtrLike {
		container, err = ptr.RHS(env)
		if err != nil {
			return SexpNull, err
		}
	}

	var hash *SexpHash
	switch e := container.(type) {
	case *SexpHash:
		hash = e
	default:
		return SexpNull, fmt.Errorf("first argument to h* function must be hash")
	}

	switch name {
	case "hget":
		if len(args) == 3 {
			return hash.HashGetDefault(env, args[1], args[2])
		}
		return hash.HashGet(env, args[1])
	case "hset":
		if len(args) != 3 {
			return SexpNull, WrongNargs
		}
		err := hash.HashSet(args[1], args[2])
		return SexpNull, err
	case "hdel":
		if len(args) != 2 {
			return SexpNull, WrongNargs
		}
		err := hash.HashDelete(args[1])
		return SexpNull, err
	case "keys":
		if len(args) != 1 {
			return SexpNull, WrongNargs
		}
		keys := make([]Sexp, 0)
		n := len(hash.KeyOrder)
		arr := &SexpArray{Env: env}
		for i := 0; i < n; i++ {
			keys = append(keys, (hash.KeyOrder)[i])

			// try to get a .Typ value going too... from the first available.
			if arr.Typ == nil {
				arr.Typ = (hash.KeyOrder)[i].Type()
			}
		}
		arr.Val = keys
		return arr, nil
	case "hpair":
		if len(args) != 2 {
			return SexpNull, WrongNargs
		}
		switch posreq := args[1].(type) {
		case *SexpInt:
			pos := int(posreq.Val)
			if pos < 0 || pos >= len(hash.KeyOrder) {
				return SexpNull, fmt.Errorf("hpair position request %d out of bounds", pos)
			}
			return hash.HashPairi(pos)
		default:
			return SexpNull, fmt.Errorf("hpair position request must be an integer")
		}
	}

	return SexpNull, nil
}

func SliceFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 3 {
		return SexpNull, WrongNargs
	}

	var start int
	var end int
	switch t := args[1].(type) {
	case *SexpInt:
		start = int(t.Val)
	case *SexpChar:
		start = int(t.Val)
	default:
		return SexpNull, fmt.Errorf("Second argument of slice must be integer")
	}

	switch t := args[2].(type) {
	case *SexpInt:
		end = int(t.Val)
	case *SexpChar:
		end = int(t.Val)
	default:
		return SexpNull, fmt.Errorf("Third argument of slice must be integer")
	}

	switch t := args[0].(type) {
	case *SexpArray:
		return &SexpArray{Val: t.Val[start:end], Env: env, Typ: t.Typ}, nil
	case *SexpStr:
		return &SexpStr{S: t.S[start:end]}, nil
	}

	return SexpNull, fmt.Errorf("First argument of slice must be array or string")
}

func ConcatFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}
	var err error
	args, err = env.ResolveDotSym(args)
	if err != nil {
		return SexpNull, err
	}

	switch t := args[0].(type) {
	case *SexpArray:
		return ConcatArray(t, args[1:])
	case *SexpStr:
		return ConcatStr(t, args[1:])
	case *SexpPair:
		n := len(args)
		switch {
		case n == 1:
			return t, nil
		default:
			return ConcatLists(t, args[1:])
		}
	}

	return SexpNull, fmt.Errorf("expected strings, lists or arrays")
}

func ReadFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	str := ""
	switch t := args[0].(type) {
	case *SexpStr:
		str = t.S
	default:
		return SexpNull, WrongType
	}
	env.Parser.ResetAddNewInput(bytes.NewBuffer([]byte(str)))
	exp, err := env.Parser.ParseExpression(0)
	return exp, err
}

func TypeQueryFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var result bool

	switch name {
	case "type?":
		return TypeOf(args[0]), nil
	case "list?":
		result = IsList(args[0])
	case "null?":
		result = (args[0] == SexpNull)
	case "array?":
		result = IsArray(args[0])
	case "number?":
		result = IsNumber(args[0])
	case "float?":
		result = IsFloat(args[0])
	case "int?":
		result = IsInt(args[0])
	case "char?":
		result = IsChar(args[0])
	case "symbol?":
		result = IsSymbol(args[0])
	case "string?":
		result = IsString(args[0])
	case "hash?":
		result = IsHash(args[0])
	case "zero?":
		result = IsZero(args[0])
	case "empty?":
		result = IsEmpty(args[0])
	case "func?":
		result = IsFunc(args[0])
	}

	return &SexpBool{Val: result}, nil
}

func NotFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	result := &SexpBool{Val: !IsTruthy(args[0])}
	return result, nil
}

func ApplyFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	var fun *SexpFunction
	var funargs []Sexp

	switch e := args[0].(type) {
	case *SexpFunction:
		fun = e
	default:
		return SexpNull, fmt.Errorf("first argument must be function")
	}

	switch e := args[1].(type) {
	case *SexpArray:
		funargs = e.Val
	case *SexpPair:
		var err error
		funargs, err = ListToArray(e)
		if err != nil {
			return SexpNull, err
		}
	default:
		return SexpNull, fmt.Errorf("second argument must be array or list")
	}

	return env.Apply(fun, funargs)
}

func MapFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	var fun *SexpFunction

	VPrintf("\n debug Map: args = '%#v'\n", args)

	switch e := args[0].(type) {
	case *SexpFunction:
		fun = e
	default:
		return SexpNull, fmt.Errorf("first argument must be function, but we had %T / val = '%#v'", e, e)
	}

	switch e := args[1].(type) {
	case *SexpArray:
		return MapArray(env, fun, e)
	case *SexpPair:
		x, err := MapList(env, fun, e)
		return x, err
	default:
		return SexpNull, fmt.Errorf("second argument must be array or list; we saw %T / val = %#v", e, e)
	}
}

func MakeArrayFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	var size int
	switch e := args[0].(type) {
	case *SexpInt:
		size = int(e.Val)
	default:
		return SexpNull, fmt.Errorf("first argument must be integer")
	}

	var fill Sexp
	if len(args) == 2 {
		fill = args[1]
	} else {
		fill = SexpNull
	}

	arr := make([]Sexp, size)
	for i := range arr {
		arr[i] = fill
	}

	return env.NewSexpArray(arr), nil
}

func ConstructorFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	switch name {
	case "array":
		return env.NewSexpArray(args), nil
	case "list":
		return MakeList(args), nil
	case "hash":
		return MakeHash(args, "hash", env)
	case "raw":
		return MakeRaw(args)
	case "field":
		Q("making hash for field")
		h, err := MakeHash(args, "field", env)
		if err != nil {
			return SexpNull, err
		}
		fld := (*SexpField)(h)
		Q("hash for field is: '%v'", fld.SexpString(nil))
		return fld, nil
	case "struct":
		return MakeHash(args, "struct", env)
	case "msgmap":
		switch len(args) {
		case 0:
			return MakeHash(args, name, env)
		default:
			var arr []Sexp
			var err error
			if len(args) > 1 {
				arr, err = ListToArray(args[1])
				if err != nil {
					return SexpNull, fmt.Errorf("error converting "+
						"'%s' arguments to an array: '%v'", name, err)
				}
			} else {
				arr = args[1:]
			}
			switch nm := args[0].(type) {
			case *SexpStr:
				return MakeHash(arr, nm.S, env)
			case *SexpSymbol:
				return MakeHash(arr, nm.name, env)
			default:
				return MakeHash(arr, name, env)
			}
		}
	}
	return SexpNull, fmt.Errorf("invalid constructor")
}

func ThreadMapFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 2 {
		return SexpNull, WrongNargs
	}

	h, isHash := args[0].(*SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("-> error: first argument must be a hash or defmap")
	}

	field, err := threadingHelper(env, h, args[1:])
	if err != nil {
		return SexpNull, err
	}

	return field, nil
}

func threadingHelper(env *Zlisp, hash *SexpHash, args []Sexp) (Sexp, error) {
	if len(args) == 0 {
		panic("should not recur without arguments")
	}
	field, err := hash.HashGet(env, args[0])
	if err != nil {
		return SexpNull, fmt.Errorf("-> error: field '%s' not found",
			args[0].SexpString(nil))
	}
	if len(args) > 1 {
		h, isHash := field.(*SexpHash)
		if !isHash {
			return SexpNull, fmt.Errorf("request for field '%s' was "+
				"not on a hash or defmap; instead type %T with value '%#v'",
				args[1].SexpString(nil), field, field)
		}
		return threadingHelper(env, h, args[1:])
	}
	return field, nil
}

func StringifyFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	return &SexpStr{S: args[0].SexpString(nil)}, nil
}

// GenericAccessFunction handles arrays or hashes
func GenericAccessFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 || len(args) > 3 {
		return SexpNull, WrongNargs
	}

	// handle *SexpSelector
	container := args[0]
	var err error
	if ptr, isPtrLike := container.(Selector); isPtrLike {
		container, err = ptr.RHS(env)
		if err != nil {
			return SexpNull, err
		}
	}

	switch container.(type) {
	case *SexpHash:
		return HashAccessFunction(env, name, args)
	case *SexpArray:
		return ArrayAccessFunction(env, name, args)
	}
	return SexpNull, fmt.Errorf("first argument to hget function must be hash or array")
}

func GOOSFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 0 {
		return SexpNull, WrongNargs
	}
	return &SexpStr{S: runtime.GOOS}, nil
}

func DerefFunction(env *Zlisp, name string, args []Sexp) (result Sexp, err error) {
	result = SexpNull

	defer func() {
		e := recover()
		if e != nil {
			//Q("in recover() of DerefFunction, e = '%#v'", e)
			switch ve := e.(type) {
			case *reflect.ValueError:
				err = ve
			default:
				err = fmt.Errorf("unknown typecheck error during %s: %v", name, ve)
			}
		}
	}()

	narg := len(args)
	if narg != 1 && narg != 2 {
		return SexpNull, WrongNargs
	}
	var ptr *SexpPointer
	switch e := args[0].(type) {
	case *SexpPointer:
		ptr = e
	case *SexpReflect:
		ptr = NewSexpPointer(e)
	default:
		return SexpNull, fmt.Errorf("%s only operates on pointers (*SexpPointer); we saw %T instead", name, e)
	}

	switch name {
	case "deref":
		if narg != 1 {
			return SexpNull, WrongNargs
		}
		return ptr.Target, nil

	case "derefSet":
		if narg != 2 {
			return SexpNull, WrongNargs
		}

		// delegate as much as we can to the Go type system
		// and reflection
		rhs := reflect.ValueOf(args[1])
		rhstype := rhs.Type()
		lhstype := ptr.ReflectTarget.Type()
		//P("rhstype = %#v, lhstype = %#v", rhstype, lhstype)
		if lhstype == rhstype {
			// have to exclude *SexpHash and *SexpReflect from this
			switch args[1].(type) {
			case *SexpHash:
				// handle below
			//case *SexpReflect:
			// handle here or below?
			default:
				//P("we have a reflection capable type match!")
				ptr.ReflectTarget.Elem().Set(rhs.Elem())
				return
			}
		}

		//P("derefSet: arg0 is %T and arg1 is %T,   ptr.Target = %#v", args[0], args[1], ptr.Target)
		//P("args[0] has ptr.ReflectTarget = '%#v'", ptr.ReflectTarget)
		switch payload := args[1].(type) {
		case *SexpInt:
			Q("ptr = '%#v'", ptr)
			Q("ptr.ReflectTarget = '%#v'", ptr.ReflectTarget)
			Q("ptr.ReflectTarget.CanAddr() = '%#v'", ptr.ReflectTarget.Elem().CanAddr())
			Q("ptr.ReflectTarget.CanSet() = '%#v'", ptr.ReflectTarget.Elem().CanSet())
			Q("*SexpInt case: payload = '%#v'", payload)
			vo := reflect.ValueOf(payload.Val)
			vot := vo.Type()
			if !vot.AssignableTo(ptr.ReflectTarget.Elem().Type()) {
				return SexpNull, fmt.Errorf("type mismatch: value of type '%s' is not assignable to type '%v'",
					vot, ptr.ReflectTarget.Elem().Type())
			}
			ptr.ReflectTarget.Elem().Set(vo)
			return
		case *SexpStr:
			vo := reflect.ValueOf(payload.S)
			vot := vo.Type()
			//P("payload is *SexpStr")
			//tele := ptr.ReflectTarget.Elem()
			//P("ptr = %#v", ptr)
			tele := ptr.ReflectTarget
			//P("got past tele : %#v", tele)
			if !reflect.PtrTo(vot).AssignableTo(tele.Type()) {
				return SexpNull, fmt.Errorf("type mismatch: value of type '%v' is not assignable to '%v'",
					vot, ptr.PointedToType.RegisteredName) // tele.Type())
			}
			//P("payload is *SexpStr, got past type check")
			ptr.ReflectTarget.Elem().Set(vo)
			return
		case *SexpHash:
			//P("ptr.PointedToType = '%#v'", ptr.PointedToType)
			pt := payload.Type()
			tt := ptr.PointedToType
			if tt == pt && tt.RegisteredName == pt.RegisteredName {
				//P("have matching type!: %v", tt.RegisteredName)
				ptr.Target.(*SexpHash).CloneFrom(payload)
				return
			} else {
				return SexpNull, fmt.Errorf("cannot assign type '%v' to type '%v'",
					payload.Type().RegisteredName,
					ptr.PointedToType.RegisteredName)
			}

		case *SexpReflect:
			Q("good, e2 is SexpReflect with Val='%#v'", payload.Val)

			Q("ptr.Target = '%#v'.  ... trying SexpToGoStructs()", ptr.Target)
			iface, err := SexpToGoStructs(payload, ptr.Target, env, nil)
			if err != nil {
				return SexpNull, err
			}
			Q("got back iface = '%#v'", iface)
			panic("not done yet with this implementation of args[1] of type *SexpReflect")
		}
		return SexpNull, fmt.Errorf("derefSet doesn't handle assignment of type %T at present", args[1])

	default:
		return SexpNull, fmt.Errorf("unimplemented operation '%s' in DerefFunction", name)
	}
}

