package codegen

func EscapeJavaReservedWord(s string) string {
	if IsJavaReservedWord(s) {
		return s + "_"
	}
	return s
}

func IsJavaReservedWord(s string) bool {
	switch s {
	case
		"abstract", "assert", "boolean", "break", "byte",
		"case", "catch", "char", "class", "const",
		"continue", "default", "do", "double", "else",
		"enum", "extends", "final", "finally", "float",
		"for", "if", "goto", "implements", "import",
		"instanceof", "int", "interface", "long", "native",
		"new", "package", "private", "protected", "public",
		"return", "short", "static", "strictfp", "super",
		"switch", "synchronized", "this", "throw", "throws",
		"transient", "try", "void", "volatile", "while":
		return true
	default:
		return false
	}
}
