package base

func Contain[v comparable](aSlice []v, target v) bool {
	if aSlice == nil {
		return false
	}
	for _, v := range aSlice {
		if v == target {
			return true
		}
	}
	return false
}
