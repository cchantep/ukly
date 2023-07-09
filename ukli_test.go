package main

import (
	"fmt"
	"testing"
)

func TestCheckConfigFile(t *testing.T) {
	testCases := []struct {
		file     string
		indent   string
		maxLine  uint
		expected error
	}{
		// Valid Configurations
		{
			file:     "valid1.conf",
			indent:   "\t",
			maxLine:  80,
			expected: nil,
		},
		{
			file:     "valid2.conf",
			indent:   "    ",
			maxLine:  80,
			expected: nil,
		},
		{
			file:     "valid3.conf",
			indent:   "  ",
			maxLine:  80,
			expected: nil,
		},

		// Invalid Configurations
		{
			file:     "e001.conf",
			indent:   "\t",
			maxLine:  80,
			expected: fmt.Errorf("[E001] More than one blank line successively at line 3"),
		},
		{
			file:     "e002.conf",
			indent:   "    ",
			maxLine:  80,
			expected: fmt.Errorf("[E002] Blank line is not allowed at line 4"),
		},
		{
			file:     "e003.conf",
			indent:   "  ",
			maxLine:  80,
			expected: fmt.Errorf("[E003] Expecting a blank line after nested section at line 4"),
		},
		{
			file:     "e004a.conf",
			indent:   "    ",
			maxLine:  80,
			expected: fmt.Errorf("[E004] Missing blank line before section declaration at line 2"),
		},
		{
			file:     "e004b.conf",
			indent:   "  ",
			maxLine:  80,
			expected: nil,
		},
		{
			file:     "e004c.conf",
			indent:   "  ",
			maxLine:  80,
			expected: nil,
		},
		{
			file:     "e004d.conf",
			indent:   "  ",
			maxLine:  80,
			expected: nil,
		},
		{
			file:     "e005.conf",
			indent:   "\t",
			maxLine:  80,
			expected: fmt.Errorf("[E005] Indentation mismatch at line 2 (0 != 1)"),
		},
		{
			file:     "e006.conf",
			indent:   "\t",
			maxLine:  80,
			expected: fmt.Errorf("[E006] Whitespace characters must be trimmed on blank line 2"),
		},
		{
			file:     "e007a.conf",
			indent:   "  ",
			maxLine:  20,
			expected: fmt.Errorf("[E007] Line 1 is too long: 70 > 20"),
		},
		{
			file:     "e007b.conf",
			indent:   "  ",
			maxLine:  40,
			expected: fmt.Errorf("[E007] Line 4 is too long: 90 > 40"),
		},
		{
			file:     "f001a.conf",
			indent:   "  ",
			maxLine:  80,
			expected: fmt.Errorf("[F001] Unbalanced '}' at line 7"),
		},
		{
			file:     "f001b.conf",
			indent:   "  ",
			maxLine:  80,
			expected: fmt.Errorf("[F001] Unbalanced ']' at line 7"),
		},
		{
			file:     "f002.conf",
			indent:   "  ",
			maxLine:  80,
			expected: fmt.Errorf("[F002] Invalid assignation ':' at line 3"),
		},
	}

	for _, tc := range testCases {
		err := checkConfigFile(
			"./resources/"+tc.file,
			tc.indent,
			tc.maxLine,
		)

		if err != nil && tc.expected == nil {
			t.Errorf("Unexpected error for '%s'; Expected: %v, Got: %v", tc.file, tc.expected, err)
		} else if err == nil && tc.expected != nil {
			t.Errorf("Expected error for '%s'; Expected: %v, Got: nil", tc.file, tc.expected)
		} else if err != nil && tc.expected != nil && err.Error() != tc.expected.Error() {
			t.Errorf("Expected error for '%s'; Expected: %v, Got: %v", tc.file, tc.expected, err)
		}
	}
}
