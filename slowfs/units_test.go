package slowfs

import "testing"

func TestNumBytesMin(t *testing.T) {
	cases := []struct {
		a    NumBytes
		b    NumBytes
		want NumBytes
	}{
		{1, 1, 1},
		{100, -12, -12},
		{100, 101, 100},
		{0, 1, 0},
	}

	for _, c := range cases {
		if got, want := NumBytesMin(c.a, c.b), c.want; got != want {
			t.Errorf("NumBytesMin(%d, %d) = %d, want %d", c.a, c.b, got, want)
		}
	}
}
