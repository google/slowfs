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
	"strconv"
	"strings"
)

// NumBytes is used for storing a number of bytes or offsets.
type NumBytes int64

// Some standard data sizes.
const (
	Byte     NumBytes = 1
	Kilobyte          = 1000 * Byte
	Megabyte          = 1000 * Kilobyte
	Gigabyte          = 1000 * Megabyte
	Terabyte          = 1000 * Gigabyte
	Kibibyte          = 1024 * Byte
	Mebibyte          = 1024 * Kibibyte
	Gibibyte          = 1024 * Mebibyte
	Tebibyte          = 1024 * Gibibyte
)

// NumBytesMin returns the smaller of the two passed NumBytes values.
func NumBytesMin(a, b NumBytes) NumBytes {
	if a > b {
		return b
	}
	return a
}

func (n NumBytes) String() string {
	var base NumBytes
	var suffix string
	switch {
	case n >= Terabyte:
		base, suffix = Terabyte, "TB"
	case n >= Gigabyte:
		base, suffix = Gigabyte, "GB"
	case n >= Megabyte:
		base, suffix = Megabyte, "MB"
	case n >= Kilobyte:
		base, suffix = Kilobyte, "KB"
	default:
		base, suffix = 1, "B"
	}

	var strRep string
	if n%base == 0 {
		strRep = fmt.Sprintf("%d%s", n/base, suffix)
	} else {
		strRep = fmt.Sprintf("%.2f%s", float64(n)/float64(base), suffix)
	}

	return fmt.Sprintf("%s (%d)", strRep, int64(n))
}

func parseSuffix(suffix string) (NumBytes, error) {
	switch strings.ToLower(suffix) {
	case "b":
		return Byte, nil
	case "kb":
		return Kilobyte, nil
	case "mb":
		return Megabyte, nil
	case "gb":
		return Gigabyte, nil
	case "tb":
		return Terabyte, nil
	case "kib":
		return Kibibyte, nil
	case "mib":
		return Mebibyte, nil
	case "gib":
		return Gibibyte, nil
	case "tib":
		return Tebibyte, nil
	default:
		return 0, fmt.Errorf("unrecognised size suffix %s", suffix)
	}
}

// ParseNumBytesFromString parses a string of the form "<number><suffix>" to a NumBytes type.
// For example, "12KB", "43.11KiB", "0B", "33TiB".
func ParseNumBytesFromString(s string) (NumBytes, error) {
	s = strings.ToLower(s)
	// Byte, Kilo, Mega, Giga, Tera
	splitIdx := strings.IndexAny(s, "bkmgt")
	if splitIdx == -1 {
		return 0, errors.New("missing suffix for size")
	}
	num, err := strconv.ParseFloat(strings.TrimSpace(s[:splitIdx]), 64)
	if err != nil {
		return 0, err
	}
	suffix, err := parseSuffix(strings.TrimSpace(s[splitIdx:]))
	if err != nil {
		return 0, err
	}
	return NumBytes(num * float64(suffix)), nil
}
