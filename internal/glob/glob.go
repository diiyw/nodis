// Package glob matches Redis byte patterns. '*' includes '/', unlike filepath.Match.
package glob

func Match(pattern, value string) bool {
	p, v, star, retry := 0, 0, -1, 0
	for v < len(value) {
		if p < len(pattern) && pattern[p] == '*' {
			star = p
			p++
			retry = v
			continue
		}
		if p < len(pattern) {
			next, matched := token(pattern, p, value[v])
			if matched {
				p = next
				v++
				continue
			}
		}
		if star < 0 {
			return false
		}
		retry++
		v = retry
		p = star + 1
	}
	for p < len(pattern) && pattern[p] == '*' {
		p++
	}
	return p == len(pattern)
}
func token(pattern string, p int, b byte) (int, bool) {
	switch pattern[p] {
	case '?':
		return p + 1, true
	case '\\':
		if p+1 < len(pattern) {
			return p + 2, pattern[p+1] == b
		}
		return p + 1, b == '\\'
	case '[':
		i := p + 1
		negated := i < len(pattern) && pattern[i] == '^'
		if negated {
			i++
		}
		matched := false
		for i < len(pattern) && pattern[i] != ']' {
			start := pattern[i]
			if start == '\\' && i+1 < len(pattern) {
				i++
				start = pattern[i]
			}
			i++
			end := start
			if i+1 < len(pattern) && pattern[i] == '-' && pattern[i+1] != ']' {
				i++
				end = pattern[i]
				i++
			}
			if start > end {
				start, end = end, start
			}
			if b >= start && b <= end {
				matched = true
			}
		}
		if i == len(pattern) {
			return p + 1, false
		}
		if negated {
			matched = !matched
		}
		return i + 1, matched
	default:
		return p + 1, pattern[p] == b
	}
}
