package glob

import "testing"

func TestMatch(t *testing.T) {
	for _, tc := range []struct {
		p, s string
		want bool
	}{{"*", "a/b", true}, {"a?c", "abc", true}, {"a?c", "ac", false}, {"a[b-d]", "ac", true}, {"a[^b]", "ac", true}, {"a[^b]", "ab", false}, {`a\*`, "a*", true}, {"a*b*c", "axbyc", true}, {"a*b*c", "axby", false}, {"*a*a", "aaa", true}, {"[", "[", false}} {
		if got := Match(tc.p, tc.s); got != tc.want {
			t.Errorf("%q %q = %v", tc.p, tc.s, got)
		}
	}
}
