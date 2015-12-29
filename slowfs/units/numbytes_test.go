// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package units

import (
	"errors"
	"fmt"
	"testing"
)

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

func TestNumBytes_String(t *testing.T) {
	cases := []struct {
		numBytes NumBytes
		want     string
	}{
		{1000, "1KB (1000)"},
		{1024, "1.02KB (1024)"},
		{1000000, "1MB (1000000)"},
		{1048576, "1.05MB (1048576)"},
		{1000000000, "1GB (1000000000)"},
		{1073741824, "1.07GB (1073741824)"},
		{1000000000000, "1TB (1000000000000)"},
		{1099511627776, "1.10TB (1099511627776)"},
		{1234, "1.23KB (1234)"},
		{23672, "23.67KB (23672)"},
		{62753, "62.75KB (62753)"},
		{0, "0B (0)"},
		{123, "123B (123)"},
		{-123, "-123B (-123)"},
	}

	for _, c := range cases {
		if got, want := c.numBytes.String(), c.want; got != want {
			t.Errorf("%d.String() = %s, want %s", int64(c.numBytes), got, want)
		}
	}
}

func ExampleNumBytes_String() {
	n := Kibibyte
	fmt.Println(n)
	m := 123 * Megabyte
	fmt.Println(m)
	// Output:
	// 1.02KB (1024)
	// 123MB (123000000)
}

func TestParseNumBytesFromString(t *testing.T) {
	cases := []struct {
		strNumBytes string
		want        NumBytes
		shouldErr   bool
	}{
		{"1  KB", 1000, false},
		{"1 KiB", 1024, false},
		{"1MB", 1000000, false},
		{"1  MiB", 1048576, false},
		{"1   GB", 1000000000, false},
		{"1 GiB", 1073741824, false},
		{"1 TB", 1000000000000, false},
		{"1   TiB", 1099511627776, false},
		{"1.234KB", 1234, false},
		{"23.672KB", 23672, false},
		{"62.753  KB  ", 62753, false},
		{"62.753GiB  ", 67380520681, false},
		{"  0  B  ", 0, false},
		{"  123  B", 123, false},
		{"  -123  B  ", -123, false},
		{"42Test", 0, true},
		{"42tEst", 0, true},
		{"42te", 0, true},
		{"asdf", 0, true},
		{"", 0, true},
		{"!@#", 0, true},
		{"432", 0, true},
	}

	for _, c := range cases {
		got, err := ParseNumBytesFromString(c.strNumBytes)
		var expectedErr error
		if c.shouldErr {
			expectedErr = errors.New("expected an error")
		}

		if got != c.want {
			t.Errorf("ParseNumBytesFromString(%s) = %s, want %s", c.strNumBytes, got, c.want)
		}

		if c.shouldErr != (err != nil) {
			t.Errorf("ParseNumBytesFromString(%s) = _, %v, want _, %v", c.strNumBytes, err, expectedErr)
		}
	}
}

func ExampleParseNumBytesFromString() {
	n, _ := ParseNumBytesFromString("12.3KB")
	fmt.Println(n)
	m, _ := ParseNumBytesFromString("10B")
	fmt.Println(m)
	// Output:
	// 12.30KB (12300)
	// 10B (10)
}
