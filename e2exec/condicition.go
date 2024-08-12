package e2exec

func TrueThen(b bool, trueFunc, falseFunc func()) func() {
	if b {
		return trueFunc
	}
	return falseFunc
}

func TrueThenExec(b bool, trueFunc, falseFunc func() any) any {
	if b {
		return trueFunc()
	}
	return falseFunc()
}
