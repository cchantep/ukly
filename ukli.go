package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	config, err := parseConfigArgs()

	printUsage := func() {
		fmt.Printf("Usage of ukli: %s [options] /path/to/dir1 [...more dir paths]\n\n", os.Args[0])

		flag.PrintDefaults()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Fails to parse arguments: %v\n", err)

		printUsage()

		os.Exit(1)
	} else if config.PrintUsage {
		printUsage()

		os.Exit(0)
	} else if len(config.DirectoryPaths) == 0 {
		fmt.Fprintf(os.Stderr, "Missing directory path(s)\n\n")

		printUsage()

		os.Exit(2)
	}

	// ---

	ext := fmt.Sprintf(".%s", config.FileExtension)
	fileErrors := uint(0)

	for _, dirPath := range config.DirectoryPaths {
		// Recursively scan the directory
		err := filepath.Walk(dirPath, func(
			path string,
			info os.FileInfo,
			err error,
		) error {
			if err != nil {
				return err
			}

			if info.IsDir() || !strings.HasSuffix(info.Name(), ext) {
				// Process only regular files with extension
				return nil
			}

			for _, re := range config.ExcludeFiles {
				if re.MatchString(path) {
					return nil
				}
			}

			// ---

			if err := checkConfigFile(path, config.Indent, config.LineMaxLen); err != nil {
				fmt.Fprintf(os.Stderr, "File '%s' is not properly formatted: %v\n", path, err)

				fileErrors++
				//os.Exit(2)
			} else {
				fmt.Printf("File '%s' is properly formatted.\n", path)
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(3)
		}
	}

	if fileErrors > 0 {
		os.Exit(2)
	}
}

func checkConfigFile(
	filePath string,
	indent string,
	lineMaxLen uint,
) error {
	file, err := os.Open(filePath)

	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := uint(0)
	indentLevel := uint(0)
	inCurly := uint(0)
	inBracket := uint(0)
	danglingAssignation := false

	prevLineComment := false
	prevLineEmpty := false
	prevLineSectionDecl := false
	expectingLineEmpty := false
	preventLineEmpty := false

	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		// Process line characters after the indent
		lineLen := len(line)

		if lineLen > int(lineMaxLen) {
			return fmt.Errorf("[E007] Line %d is too long: %d > %d",
				lineNumber, lineLen, lineMaxLen)
		}

		trimmedLine := strings.TrimLeft(line, indent)
		trimmedLen := len(trimmedLine)

		// Check for blank lines
		if trimmedLen == 0 {
			if prevLineEmpty {
				// No more than one blank line
				return fmt.Errorf("[E001] More than one blank line successively at line %d", lineNumber)
			} else if preventLineEmpty {
				// No blank line after section started with '{' or '['
				return fmt.Errorf("[E002] Blank line is not allowed at line %d", lineNumber)
			}

			expectingLineEmpty = false
			prevLineEmpty = true
			prevLineSectionDecl = false

			continue
			// Skip further check for blank lines
		}

		if len(strings.TrimSpace(line)) == 0 && lineLen > 0 {
			// Whitespace characters on blank line
			return fmt.Errorf("[E006] Whitespace characters must be trimmed on blank line %d", lineNumber)
		}

		// ---

		// First non-whitespace character
		firstNonWhite := trimmedLine[0]
		preventLineEmpty = false

		if firstNonWhite == '#' || (firstNonWhite == '/' &&
			len(trimmedLine) > 1 &&
			trimmedLine[1] == '/') {
			// Skip comment lines starting with '#' or '//'
			prevLineEmpty = false
			prevLineComment = true
		} else {
			// Handling non comment line
			if expectingLineEmpty && firstNonWhite != '}' && firstNonWhite != ']' {
				return fmt.Errorf("[E003] Expecting a blank line after nested section at line %d", lineNumber)
			}

			// ---

			if err := checkNonCommentLine(
				indent,
				&indentLevel,
				lineNumber,
				line,
				lineLen,
				trimmedLen,
				firstNonWhite,
				&prevLineComment,
				&prevLineEmpty,
				&prevLineSectionDecl,
				&preventLineEmpty,
				&expectingLineEmpty,
				&danglingAssignation,
				&inCurly,
				&inBracket,
			); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// ---

type Config struct {
	PrintUsage     bool
	Indent         string
	FileExtension  string
	DirectoryPaths []string
	ExcludeFiles   []*regexp.Regexp
	LineMaxLen     uint
}

// Create config object from CLI arguments
func parseConfigArgs() (Config, error) {
	// Define flags
	help := flag.Bool("help", false, "Print this help")

	indent := flag.String("indent", "  ", "Indentation string (default: '  ')")

	fileExtension := flag.String(
		"file-extension", "conf", "File extension")

	excludeFiles := flag.String(
		"exclude-file", "", "Exclude file pattern (comma separated for multiple patterns)")

	lineMaxLength := flag.Uint("line-max-length", 100, "Maximum line length (default: 100)")

	// Parse command line arguments
	flag.Parse()

	// Extract directory paths (non-optional arguments without "-" prefix)
	directoryPaths := make([]string, 0)

	for _, arg := range flag.Args() {
		if !strings.HasPrefix(arg, "-") {
			directoryPaths = append(directoryPaths, arg)
		}
	}

	// Extract exclude file patterns
	excludePatterns := strings.Split(*excludeFiles, ",")
	excludeRegexps := make([]*regexp.Regexp, 0)

	for _, pattern := range excludePatterns {
		if len(pattern) > 0 {
			regex, err := regexp.Compile(pattern)

			if err != nil {
				return Config{}, fmt.Errorf("Invalid exclude file pattern: %s\n", pattern)
			}

			excludeRegexps = append(excludeRegexps, regex)
		}
	}

	return Config{
		PrintUsage:     *help,
		Indent:         *indent,
		FileExtension:  *fileExtension,
		DirectoryPaths: directoryPaths,
		ExcludeFiles:   excludeRegexps,
		LineMaxLen:     *lineMaxLength,
	}, nil
}

func checkNonCommentLine(
	indent string,
	indentLevel *uint,
	lineNumber uint,
	line string,
	lineLen int,
	trimmedLen int,
	firstNonWhite byte,
	prevLineComment *bool,
	prevLineEmpty *bool,
	prevLineSectionDecl *bool,
	preventLineEmpty *bool,
	expectingLineEmpty *bool,
	danglingAssignation *bool,
	inCurly *uint,
	inBracket *uint,
) error {
	waitingAssignation := false
	nextLineChange := uint(0)
	inVariableDecl := false
	inQuote := false // TODO: Check quote balance

	leadingSpaces := lineLen - trimmedLen
	currIndentLevel := uint(leadingSpaces) / uint(len(indent))

	for i := leadingSpaces; i < lineLen; i++ {
		ch := line[i]

		if inQuote {
			if ch == '"' {
				inQuote = false
			}

			continue
		} else if ch == '"' {
			inQuote = true
		} else if ch == '$' {
			if i+1 < lineLen && line[i+1] == '{' {
				inVariableDecl = true
				i++ // Skip '{'
			}
		} else if ch == '=' || ch == ':' {
			if waitingAssignation {
				return fmt.Errorf("[F002] Invalid assignation '%s' at line %d", string(ch), lineNumber)
			}

			waitingAssignation = true
		} else if ch == '}' && inVariableDecl {
			inVariableDecl = false
		} else if ch == '{' || ch == '[' {
			nextLineChange++

			if ch == '{' {
				*inCurly++
			} else {
				*inBracket++
			}

			if i+1 == lineLen {
				// Section start at end of line

				if *inBracket == 0 && !*prevLineEmpty && !*prevLineComment && !*prevLineSectionDecl && lineNumber > 1 {
					// No blank line before a line declaring a section
					return fmt.Errorf("[E004] Missing blank line before section declaration at line %d", lineNumber)
				}

				*preventLineEmpty = true
			}
		} else if (ch == ']' && *inBracket == 0) || (ch == '}' && *inCurly == 0) {
			return fmt.Errorf(
				"[F001] Unbalanced '%s' at line %d",
				string(ch), lineNumber)

		} else if ch == ']' || ch == '}' {
			if ch == ']' {
				*inBracket--
			}

			if ch == '}' {
				*inCurly--
			}

			if ni := i + 1; i == leadingSpaces && (ni == lineLen || line[ni] != ',') {
				// First non-whitespace
				*expectingLineEmpty = true
			} else {
				*expectingLineEmpty = false
			}

			if nextLineChange > 0 {
				nextLineChange--
			} else if *indentLevel > 0 {
				*indentLevel--
			} else {
				return fmt.Errorf(
					"[F001] Unbalanced '%s' at line %d",
					string(ch), lineNumber)
			}
		} else {
			waitingAssignation = false
		}
	}

	if waitingAssignation {
		nextLineChange++
		*danglingAssignation = true
	} else if *danglingAssignation {
		nextLineChange--
		*danglingAssignation = false
	}

	*prevLineEmpty = false
	*prevLineComment = false

	// Check for consistent indentation
	if currIndentLevel != *indentLevel {
		return fmt.Errorf(
			"[E005] Indentation mismatch at line %d (%d != %d)",
			lineNumber, currIndentLevel, *indentLevel)
	}

	// Whether current line doesn't contain a section declaration
	*prevLineSectionDecl = nextLineChange > 0

	*indentLevel += nextLineChange

	return nil
}
