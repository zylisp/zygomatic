package zcore

import (
	"fmt"
	"os"
	"reflect"
	"unicode"
)

type ZlispFunction []Instruction
type ZlispUserFunction func(*Zlisp, string, []Sexp) (Sexp, error)

// AllBuiltinFunctions returns all built in functions
func AllBuiltinFunctions() map[string]ZlispUserFunction {
	return MergeFuncMap(
		CoreFunctions(),       // core.go
		StringFunctions(),     // string.go
		EncodingFunctions(),   // encoding.go
		SystemFunctions(),     // system.go
		RandomFunctions(),     // random.go
		ReflectionFunctions(), // reflection.go
	)
}

// SandboxSafeFunctions returns all functions that are safe to run in a sandbox
func SandboxSafeFunctions() map[string]ZlispUserFunction {
	return MergeFuncMap(
		CoreFunctions(),       // core.go
		StringFunctions(),     // string.go
		EncodingFunctions(),   // encoding.go
	)
}

func RandomFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"random":  RandomFunction,
	}
}

func EncodingFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"json":      JsonFunction,
		"unjson":    JsonFunction,
		"msgpack":   JsonFunction,
		"unmsgpack": JsonFunction,
		"gob":       GobEncodeFunction,
		"msgmap":    ConstructorFunction,
	}
}

func ReflectionFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"methodls":              GoMethodListFunction,
		"_method":               CallGoMethodFunction,
		"registerDemoFunctions": ScriptFacingRegisterDemoStructs,
	}
}

func SystemFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"source":    SourceFileFunction,
		"togo":      ToGoFunction,
		"fromgo":    FromGoFunction,
		"dump":      GoonDumpFunction,
		"slurpf":    SlurpfileFunction,
		"writef":    WriteToFileFunction,
		"save":      WriteToFileFunction,
		"bload":     ReadGreenpackFromFileFunction,
		"bsave":     WriteShadowGreenpackToFileFunction,
		"greenpack": WriteShadowGreenpackToFileFunction,
		"owritef":   WriteToFileFunction,
		"system":    SystemFunction,
		"exit":      ExitFunction,
		"_closdump": DumpClosureEnvFunction,
		"rmsym":     RemoveSymFunction,
		"typelist":  TypeListFunction,
		"setenv":    GetEnvFunction,
		"getenv":    GetEnvFunction,
		// not done "_call":     CallZMethodOnRecordFunction,
	}
}

func HashColonFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 2 || len(args) > 3 {
		return SexpNull, WrongNargs
	}

	var hash *SexpHash
	switch e := args[1].(type) {
	case *SexpHash:
		hash = e
	default:
		return SexpNull, fmt.Errorf("second argument of (:field hash) must be a hash")
	}

	if len(args) == 3 {
		return hash.HashGetDefault(env, args[0], args[2])
	}
	return hash.HashGet(env, args[0])
}

func LenFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var err error
	args, err = env.ResolveDotSym(args)
	if err != nil {
		return SexpNull, err
	}

	switch t := args[0].(type) {
	case *SexpSentinel:
		if t == SexpNull {
			return &SexpInt{}, nil
		}
		break
	case *SexpArray:
		return &SexpInt{Val: int64(len(t.Val))}, nil
	case *SexpStr:
		return &SexpInt{Val: int64(len(t.S))}, nil
	case *SexpHash:
		return &SexpInt{Val: int64(HashCountKeys(t))}, nil
	case *SexpPair:
		n, err := ListLen(t)
		return &SexpInt{Val: int64(n)}, err
	default:
		P("in LenFunction with args[0] of type %T", t)
	}
	return &SexpInt{}, fmt.Errorf("argument must be string, list, hash, or array")
}

func AppendFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpArray:
		switch name {
		case "append":
			return &SexpArray{Val: append(t.Val, args[1]), Env: env, Typ: t.Typ}, nil
		case "appendslice":
			switch sl := args[1].(type) {
			case *SexpArray:
				return &SexpArray{Val: append(t.Val, sl.Val...), Env: env, Typ: t.Typ}, nil
			default:
				return SexpNull, fmt.Errorf("Second argument of appendslice must be slice")
			}
		default:
			return SexpNull, fmt.Errorf("unrecognized append variant: '%s'", name)
		}
	case *SexpStr:
		return AppendStr(t, args[1])
	}

	return SexpNull, fmt.Errorf("First argument of append must be array or string")
}

func OldEvalFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	P("EvalFunction() called, name = '%s'; args = %#v", name, args)
	newenv := env.Duplicate()
	err := newenv.LoadExpressions(args)
	if err != nil {
		return SexpNull, fmt.Errorf("failed to compile expression")
	}
	newenv.pc = 0
	return newenv.Run()
}

// EvalFunction: new version doesn't use a duplicated environment,
// allowing eval to create closures under the lexical scope and
// to allow proper scoping in a package.
func EvalFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}
	//P("EvalFunction() called, name = '%s'; args = %#v", name, (&SexpArray{Val: args}).SexpString(0))

	// Instead of LoadExpressions:
	args = env.FilterArray(args, RemoveCommentsFilter)
	args = env.FilterArray(args, RemoveEndsFilter)

	startingDataStackSize := env.datastack.Size()

	gen := NewGenerator(env)
	err := gen.GenerateBegin(args)
	if err != nil {
		return SexpNull, err
	}

	newfunc := ZlispFunction(gen.instructions)
	orig := &SexpArray{Val: args}
	sfun := env.MakeFunction("evalGeneratedFunction", 0, false, newfunc, orig)

	err = env.CallFunction(sfun, 0)
	if err != nil {
		return SexpNull, err
	}

	var resultSexp Sexp
	resultSexp, err = env.Run()
	if err != nil {
		return SexpNull, err
	}

	err = env.ReturnFromFunction()

	// some sanity checks
	if env.datastack.Size() > startingDataStackSize {
		/*
			xtra := env.datastack.Size() - startingDataStackSize
			panic(fmt.Sprintf("we've left some extra stuff (xtra = %v) on the datastack "+
				"during eval, don't be sloppy, fix it now! env.datastack.Size()=%v, startingDataStackSize = %v",
				xtra, env.datastack.Size(), startingDataStackSize))
			P("warning: truncating datastack back to startingDataStackSize %v", startingDataStackSize)
		*/
		env.datastack.TruncateToSize(startingDataStackSize)
	}
	if env.datastack.Size() < startingDataStackSize {
		P("about panic, since env.datastack.Size() < startingDataStackSize, here is env dump:")
		env.DumpEnvironment()
		panic(fmt.Sprintf("we've shrunk the datastack during eval, don't be sloppy, fix it now! env.datastack.Size()=%v. startingDataStackSize=%v", env.datastack.Size(), startingDataStackSize))
	}

	return resultSexp, err
}

var MissingFunction = &SexpFunction{name: "__missing", user: true}

func (env *Zlisp) MakeFunction(name string, nargs int, varargs bool,
	fun ZlispFunction, orig Sexp) *SexpFunction {
	var sfun SexpFunction
	sfun.name = name
	sfun.user = false
	sfun.nargs = nargs
	sfun.varargs = varargs
	sfun.fun = fun
	sfun.orig = orig
	sfun.SetClosing(NewClosing(name, env)) // snapshot the create env as of now.
	return &sfun
}

func MakeUserFunction(name string, ufun ZlispUserFunction) *SexpFunction {
	var sfun SexpFunction
	sfun.name = name
	sfun.user = true
	sfun.userfun = ufun
	return &sfun
}

func MakeBuilderFunction(name string, ufun ZlispUserFunction) *SexpFunction {
	sfun := MakeUserFunction(name, ufun)
	sfun.isBuilder = true
	return sfun
}

// MergeFuncMap returns the union of the two given maps
func MergeFuncMap(funcs ...map[string]ZlispUserFunction) map[string]ZlispUserFunction {
	n := make(map[string]ZlispUserFunction)

	for _, f := range funcs {
		for k, v := range f {
			// disallow dups, avoiding possible security implications and confusion generally.
			if _, dup := n[k]; dup {
				panic(fmt.Sprintf(" duplicate function '%s' not allowed", k))
			}
			n[k] = v
		}
	}
	return n
}

func Sym2StrFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpSymbol:
		r := &SexpStr{S: t.name}
		return r, nil
	}
	return SexpNull, fmt.Errorf("argument must be symbol")
}

func ExitFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch e := args[0].(type) {
	case *SexpInt:
		os.Exit(int(e.Val))
	}
	return SexpNull, fmt.Errorf("argument must be int (the exit code)")
}

func StopFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg > 1 {
		return SexpNull, WrongNargs
	}

	if narg == 0 {
		return SexpNull, StopErr
	}

	switch s := args[0].(type) {
	case *SexpStr:
		return SexpNull, fmt.Errorf(s.S)
	}
	return SexpNull, StopErr
}

// AssignmentFunction: The assignment function, =
func AssignmentFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	Q("\n AssignmentFunction called with name ='%s'. args='%s'\n", name,
		env.NewSexpArray(args).SexpString(nil))

	narg := len(args)
	if narg != 2 {
		return SexpNull, fmt.Errorf("assignment requires two arguments: a left-hand-side and a right-hand-side argument")
	}

	var sym *SexpSymbol
	switch s := args[0].(type) {
	case *SexpSymbol:
		sym = s
	case Selector:
		err := s.AssignToSelection(env, args[1])
		return args[1], err

	default:
		return SexpNull, fmt.Errorf("assignment needs left-hand-side"+
			" argument to be a symbol; we got %T", s)
	}

	if !sym.isDot {
		Q("assignment sees LHS symbol but is not dot, binding '%s' to '%s'\n",
			sym.name, args[1].SexpString(nil))
		err := env.LexicalBindSymbol(sym, args[1])
		if err != nil {
			return SexpNull, err
		}
		return args[1], nil
	}

	Q("assignment calling DotGetSetHelper()\n")
	return DotGetSetHelper(env, sym.name, &args[1])
}

func JoinSymFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg == 0 {
		return SexpNull, nil
	}

	j := ""

	for k := range args {
		switch a := args[k].(type) {
		case *SexpPair:
			arr, err := ListToArray(args[k])
			if err != nil {
				return SexpNull, fmt.Errorf("error converting "+
					"joinsym arguments to an array: '%v'", err)
			}
			s, err := joinSymHelper(arr)
			if err != nil {
				return SexpNull, err
			}
			j += s

		case *SexpSymbol:
			j = j + a.name
		case *SexpArray:
			s, err := joinSymHelper(a.Val)
			if err != nil {
				return SexpNull, err
			}
			j += s
		default:
			return SexpNull, fmt.Errorf("error cannot joinsym type '%T' / val = '%s'", a, a.SexpString(nil))
		}
	}

	return env.MakeSymbol(j), nil
}

func joinSymHelper(arr []Sexp) (string, error) {
	j := ""
	for i := 0; i < len(arr); i++ {
		switch s := arr[i].(type) {
		case *SexpSymbol:
			j = j + s.name

		default:
			return "", fmt.Errorf("not a symbol: '%s'",
				arr[i].SexpString(nil))
		}
	}
	return j, nil
}

// '(a b c) -> ('a 'b 'c)
func QuoteListFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 1 {
		return SexpNull, WrongNargs
	}

	pair, ok := args[0].(*SexpPair)
	if !ok {
		return SexpNull, fmt.Errorf("list required")
	}

	arr, err := ListToArray(pair)
	if err != nil {
		return SexpNull, fmt.Errorf("error converting "+
			"quotelist arguments to an array: '%v'", err)
	}

	arr2 := []Sexp{}
	for _, v := range arr {
		arr2 = append(arr2, MakeList([]Sexp{env.MakeSymbol("quote"), v}))
	}

	return MakeList(arr2), nil
}

// helper used by DotGetSetHelper and sub-calls to check for private
func errIfPrivate(pathPart string, pkg *Stack) error {
	noDot := stripAnyDotPrefix(pathPart)

	// references through a package must be Public
	if !unicode.IsUpper([]rune(noDot)[0]) {
		return fmt.Errorf("Cannot access private member '%s' of package '%s'",
			noDot, pkg.PackageName)
	}
	return nil
}

// if setVal is nil, only get and return the lookup.
// Otherwise set and return the value we set.
func DotGetSetHelper(env *Zlisp, name string, setVal *Sexp) (Sexp, error) {
	path := DotPartsRegex.FindAllString(name, -1)
	//P("\n in DotGetSetHelper(), name = '%s', path = '%#v', setVal = '%#v'\n", name, path, setVal)
	if len(path) == 0 {
		return SexpNull, fmt.Errorf("internal error: DotFunction" +
			" path had zero length")
	}

	var ret Sexp = SexpNull
	var err error
	lenpath := len(path)

	if lenpath == 1 && setVal != nil {
		// single path element set, bind it now.
		a := stripAnyDotPrefix(path[0])
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

	key := stripAnyDotPrefix(path[0])
	//Q("\n in DotGetSetHelper(), looking up '%s'\n", key)
	keySym := env.MakeSymbol(key)
	ret, err, _ = env.LexicalLookupSymbol(keySym, nil)
	if err != nil {
		//Q("\n in DotGetSetHelper(), '%s' not found\n", key)
		return SexpNull, err
	}
	if lenpath == 1 {
		// single path element get, return it.
		return ret, err
	}

	// INVAR: lenpath > 1

	// package or hash? check for package
	pkg, isStack := ret.(*Stack)
	if isStack && pkg.IsPackage {
		//P("found a package: '%s'", pkg.SexpString(nil))

		exp, err := pkg.nestedPathGetSet(env, path[1:], setVal)
		if err != nil {
			return SexpNull, err
		}
		return exp, nil
	}

	// at least .a.b if not a.b.c. etc: multiple elements,
	// where .b and after
	// will index into hashes (.a must refer to a hash);
	// proceed deeper into the hashes.

	var h *SexpHash
	switch x := ret.(type) {
	case *SexpHash:
		h = x
	case *SexpReflect:
		// at least allow reading, if we can.
		if setVal != nil {
			return SexpNull, fmt.Errorf("can't set on an SexpReflect: on request for "+
				"field '%s' in non-record (instead of type %T)",
				stripAnyDotPrefix(path[1]), ret)
		}
		//P("functions.go DEBUG! SexpReflect value h is type: '%v', '%T', kind: '%v'", x.Val.Type(), x.Val.Interface(), x.Val.Type().Kind())
		if x.Val.Type().Kind() == reflect.Struct {
			//P("We have a struct! path[1]='%v', path='%#v'", path[1], path)
			if len(path) >= 2 && len(path[1]) > 0 {
				fieldName := stripAnyDotPrefix(path[1])
				//P("We have a struct! with dot request for member '%s'", fieldName)
				fld := x.Val.FieldByName(fieldName)
				if reflect.DeepEqual(fld, reflect.Value{}) {
					return SexpNull, fmt.Errorf("no such field '%s'", fieldName)
				}
				// ex:  We got back fld='20' of type int, kind=int
				//P("We got back fld='%v' of type %v, kind=%v", fld, fld.Type(), fld.Type().Kind())
				return GoToSexp(fld.Interface(), env)
			}
		}
		return SexpNull, fmt.Errorf("SexpReflect is not a struct: cannot get "+
			"field '%s' in non-struct (instead of type %T)",
			stripAnyDotPrefix(path[1]), ret)
	default:
		return SexpNull, fmt.Errorf("not a record: cannot get "+
			"field '%s' in non-record (instead of type %T)",
			stripAnyDotPrefix(path[1]), ret)
	}
	// have hash: rest of path handled in hashutils.go in nestedPathGet()
	//Q("\n in DotGetSetHelper(), about to call nestedPathGetSet() with"+
	//	"dotpaths = path[i+1:]='%#v\n", path[1:])
	exp, err := h.nestedPathGetSet(env, path[1:], setVal)
	if err != nil {
		return SexpNull, err
	}
	return exp, nil
}

func RemoveSymFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 1 {
		return SexpNull, WrongNargs
	}

	sym, ok := args[0].(*SexpSymbol)
	if !ok {
		return SexpNull, fmt.Errorf("symbol required, but saw %T/%v", args[0], args[0].SexpString(nil))
	}

	err := env.linearstack.DeleteSymbolFromTopOfStackScope(sym)
	return SexpNull, err
}

// DefinedFunction checks is a symbol/string/value is defined
func DefinedFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	//P("in DefinedFunction, args = '%#v'", args)
	narg := len(args)
	if narg != 1 {
		return SexpNull, WrongNargs
	}

	var checkme string
	switch nm := args[0].(type) {
	case *SexpStr:
		checkme = nm.S
	case *SexpSymbol:
		checkme = nm.name
	case *SexpFunction:
		return &SexpBool{Val: true}, nil
	default:
		return &SexpBool{Val: false}, nil
	}

	_, err, _ := env.LexicalLookupSymbol(env.MakeSymbol(checkme), nil)
	if err != nil {
		return &SexpBool{Val: false}, nil
	}
	return &SexpBool{Val: true}, nil
}

func AddressOfFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 1 {
		return SexpNull, WrongNargs
	}

	return NewSexpPointer(args[0]), nil
}

func stripAnyDotPrefix(s string) string {
	if len(s) > 0 && s[0] == '.' {
		return s[1:]
	}
	return s
}

// SubstituteRHS locates any SexpSelector(s) (Selector implementers, really)
// and substitutes
// the value of x.RHS() for each x in args.
func (env *Zlisp) SubstituteRHS(args []Sexp) ([]Sexp, error) {
	for i := range args {
		obj, hasRhs := args[i].(Selector)
		if hasRhs {
			sx, err := obj.RHS(env)
			if err != nil {
				return args, err
			}
			args[i] = sx
		}
	}
	return args, nil
}

func ScriptFacingRegisterDemoStructs(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	RegisterDemoStructs()
	return SexpNull, nil
}

func GetEnvFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	//fmt.Printf("GetEnv name='%s' called with narg = %v\n", name, narg)
	if name == "getenv" {
		if narg != 1 {
			return SexpNull, WrongNargs
		}
	} else {
		if name != "setenv" {
			panic("only getenv or setenv allowed here")
		}
		if narg != 2 {
			return SexpNull, WrongNargs
		}
	}
	nm := make([]string, narg)
	for i := 0; i < narg; i++ {
		switch x := args[i].(type) {
		case *SexpSymbol:
			nm[i] = x.name
		case *SexpStr:
			nm[i] = x.S
		default:
			return SexpNull, fmt.Errorf("symbol or string required, but saw %T/%v for i=%v arg", args[i], args[i].SexpString(nil), i)
		}
	}

	if name == "getenv" {
		return &SexpStr{S: os.Getenv(nm[0])}, nil
	}

	//fmt.Printf("calling setenv with nm[0]='%s', nm[1]='%s'\n", nm[0], nm[1])
	return SexpNull, os.Setenv(nm[0], nm[1])
}

// AsUint64Function: coerce numbers to uint64
func AsUint64Function(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var val uint64
	switch x := args[0].(type) {
	case *SexpInt:
		val = uint64(x.Val)
	case *SexpFloat:
		val = uint64(x.Val)
	default:
		return SexpNull, fmt.Errorf("Cannot convert %s to uint64", TypeOf(args[0]).SexpString(nil))

	}
	return &SexpUint64{Val: val}, nil
}
