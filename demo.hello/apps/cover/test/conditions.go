package test

func getConditions(cond1, cond2 bool) string {
	var ret string
	if cond1 {
		ret = "a"
	} else {
		ret = "b"
	}

	if cond2 {
		ret = ret + "c"
	} else {
		ret = ret + "d"
	}
	return ret
}
