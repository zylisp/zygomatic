package zcore

func ReflectionFunctions() map[string]ZlispUserFunction {
	return map[string]ZlispUserFunction{
		"methodls":              GoMethodListFunction,
		"_method":               CallGoMethodFunction,
		"registerDemoFunctions": ScriptFacingRegisterDemoStructs,
	}
}

func ScriptFacingRegisterDemoStructs(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	RegisterDemoStructs()
	return SexpNull, nil
}
