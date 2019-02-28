package zcore

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
