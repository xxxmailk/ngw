package flows

import "testing"

func TestSliceSearchString(t *testing.T) {
	s := []string{
		"a",
		"aa",
		"aa1",
		"aa2",
		"aa3",
		"aa4",
	}
	t.Logf("test string \"a\" %d", SliceSearchString(s, "a"))
	t.Logf("test string \"a1\", a1 is not exsited %d", SliceSearchString(s, "a1"))
	t.Logf("test string \"a\" %d", SliceSearchString(s, "aa2"))
}
