package lang

import "testing"

func TestParseRange(t *testing.T) {
	cases := []struct {
		input  string
		max    int
		ranges [][2]int
		whole  bool
	}{
		{"1", 5, [][2]int{{1, 1}}, false},
		{"1,2", 5, [][2]int{{1, 1}, {2, 2}}, false},
		{"1,2,3", 5, [][2]int{{1, 1}, {2, 2}, {3, 3}}, false},
		{"-", 5, [][2]int{}, true},
		{"", 5, [][2]int{}, true},
		{"-1", 5, [][2]int{{5, 5}}, false},
		{"1-", 5, [][2]int{{1, 5}}, false},
		{"1-2", 5, [][2]int{{1, 2}}, false},
		{"1-2,3", 5, [][2]int{{1, 2}, {3, 3}}, false},
		{"1-2,3,4-", 5, [][2]int{{1, 2}, {3, 3}, {4, 5}}, false},
		{"1-2,3-,5", 5, [][2]int{{1, 2}, {3, 5}, {5, 5}}, false},
		{"1-2,3,4-5", 5, [][2]int{{1, 2}, {3, 3}, {4, 5}}, false},
		{"1--1", 5, [][2]int{{1, 5}}, false},
	}
	for _, c := range cases {
		r, err := ParseRange(c.input, c.max)
		if err != nil {
			t.Errorf("parse range %s: %v", c.input, err)
		}
		if r.whole != c.whole {
			t.Errorf("parse range %s: whole = %v, want %v", c.input, r.whole, c.whole)
		}
		if len(r.ranges) != len(c.ranges) {
			t.Errorf("parse range %s: ranges = %v, want %v", c.input, r.ranges, c.ranges)
		}
		for i, rg := range r.ranges {
			if rg[0] != c.ranges[i][0] || rg[1] != c.ranges[i][1] {
				t.Errorf("parse range %s: ranges = %v, want %v", c.input, r.ranges, c.ranges)
			}
		}
	}

	invalidCases := []string{
		"0",
		"100",
		"1,100",
		"-100",
		"1-100",
		"1-2,100",
		"1-2,100-",
		"1-2,100-200",
		"a-b",
		"1-a",
		"100-1",
		"1-2-3",
		"error",
		"----",
		"1,,,2",
	}
	for _, c := range invalidCases {
		_, err := ParseRange(c, 5)
		if err == nil {
			t.Errorf("parse range %s: want error", c)
		}
	}
}
