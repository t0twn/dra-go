package wildcard

// Match checks if a string matches a wildcard pattern.
// Supported wildcards:
//   - '*' matches any sequence of characters (including empty)
//   - '?' matches any single character
func Match(pattern, s string) bool {
	return match(pattern, s, 0, 0)
}

func match(pattern, s string, pi, si int) bool {
	for pi < len(pattern) {
		if si < len(s) && (pattern[pi] == '?' || pattern[pi] == s[si]) {
			pi++
			si++
			continue
		}

		if pattern[pi] == '*' {
			// Skip consecutive '*'
			for pi < len(pattern) && pattern[pi] == '*' {
				pi++
			}
			// If '*' is at end, it matches everything
			if pi == len(pattern) {
				return true
			}
			// Try matching '*' with 0, 1, 2, ... characters
			for i := si; i <= len(s); i++ {
				if match(pattern, s, pi, i) {
					return true
				}
			}
			return false
		}

		return false
	}

	return si == len(s)
}
