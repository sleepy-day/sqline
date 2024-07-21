package shared

func Assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

func Scale(factor float32, value int) int {
	return int(factor * float32(value))
}
