package text

//
// Runes functions
//

func RunesMatch(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func RunesNoMatch(a, b []rune) bool {
	return !RunesMatch(a, b)
}
