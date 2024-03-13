// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"strconv"
	"strings"
)

func getIndex(input string) int {

	idx1 := strings.Index(input, ",")
	idx2 := strings.Index(input, "/")
	idx3 := strings.Index(input, "\n")

	// If there is no `,` separator in the input, return either the index of the next explicit terminating character (`/`)
	// or the index of the next newline character, if no terminating character is present.
	if idx1 == -1 {
		if idx2 != -1 {
			return idx2
		}
		return idx3
	}

	// If a line is terminated with a `/` character and the terminator is BEFORE the next `,` character, return
	// the index of the `/` character.
	if idx2 > -1 && idx2 < idx1 {
		return idx2
	}

	// If a line is terminated with a `\n` character (and is NOT terminated with a / character) and the terminator is
	// BEFORE the next `,` character, return the index of the `\n` character.
	if idx2 < 0 && idx3 > -1 && idx3 < idx1 {
		return idx3
	}

	// Otherwise, return the index of the next `,` character. Value will not be `-1` due to earlier function logic.
	return idx1
}

func ReadField(input string, start int) (string, int, error) {

	data := ""

	if start < len(input) {
		data = input[start:]
	}

	if data == "" {
		return "", 0, fmt.Errorf("doesn't enough input string")
	}

	idx := getIndex(data)
	if idx == -1 {
		return "", 0, fmt.Errorf("doesn't have valid delimiter")
	}

	return data[:idx], idx + 1, nil
}

func ReadFieldAsInt(input string, start int) (int64, int, error) {

	data := ""

	if start < len(input) {
		data = input[start:]
	}

	if data == "" {
		return 0, 0, fmt.Errorf("doesn't enough input string")
	}

	idx := getIndex(data)
	if idx == -1 {
		return 0, 0, fmt.Errorf("doesn't have valid delimiter")
	}

	if data[:idx] == "" {
		return 0, 1, nil
	}

	value, err := strconv.ParseInt(data[:idx], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("doesn't have valid value")
	}

	return value, idx + 1, nil
}

func GetSize(line string) int64 {
	size := strings.Index(line, "/")
	if size >= 0 {
		return int64(size + 1)
	}

	nsize := strings.Index(line, "\n")
	if nsize >= 0 {
		return int64(nsize + 1)
	}

	return int64(size)
}
