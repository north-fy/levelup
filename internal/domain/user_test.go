package domain

import "testing"

func TestLevelFor(t *testing.T) {
	cases := []struct {
		xp   int
		want int
	}{
		{xp: 0, want: 1},
		{xp: 50, want: 2},
		{xp: 200, want: 3},
		{xp: 500, want: 4},
		{xp: 5000, want: 11},
	}

	for _, tc := range cases {
		if got := LevelFor(tc.xp); got != tc.want {
			t.Errorf("LevelFor(%d) = %d, want %d", tc.xp, got, tc.want)
		}
	}
}
