# Ukli

Ukli is [small god of a Howondalandish tribe](https://discworld.fandom.com/wiki/Small_gods#Ukli), or a utility that verifies the proper formatting and structure of HOCON files.

It ensures your [HOCON configuration](https://github.com/lightbend/config#using-hocon-the-json-superset) files are not u~~k~~gly but correctly formatted.
It checks indentation, detects missing or extra blank lines, and provides error messages.

## Usage

Locally:

```
./ukli [options] /path/to/dir1 [...more dir paths]

  -exclude-file string
    	Exclude file pattern (comma separated for multiple patterns)
  -file-extension string
    	File extension (default "conf")
  -help
    	Print this help
  -indent string
    	Indentation string (default "  ")
```

With Docker:

```bash
docker run --rm cchantep/ukli:latest [options] /path/inside/container
```

### Diagnostics

**E001**

More than one blank line successively at line

*Example:*

Raised for:

```
option1 = value1


option2 = value2
```

Fix:

```
option1 = value1

option2 = value2
```

**E002**

Blank line is not allowed at line

*Example:*

Raised for:

```
key1 = value1

section1 {

    key2 = value2
    key3 = value3
}
```

Fix:

```
key1 = value1

section1 {
    key2 = value2
    key3 = value3
}
```

**E003**

Expecting a blank line after nested section

*Example:*

Raised for:

```
section1 {
  option1 = value1
}
section2 {
  option2 = value2
}
```

Fix:

```
section1 {
  option1 = value1
}

section2 {
  option2 = value2
}
```

**E004**

Missing blank line before section declaration

*Example:*

Raised for:

```
key1 = value1
section1 {
    key2 = value2
    key3 = value3
}
```

Fix:

```
key1 = value1

section1 {
    key2 = value2
    key3 = value3
}
```

**E005**

Indentation mismatch

*Example:*

Raised for:

```
section1 {
  option1 = value1
    option2 = value2
}
```

Fix:

```
section1 {
  option1 = value1
  option2 = value2
}
```

**E006**

Whitespace characters must be trimmed.

**E007**

Line is too long

*Example:*

Raised for:

```
another-very-long-line: "this-one-is-not-ignore-as-there-is-not-special-comment-before-it"
```

Fix:

Either split the line or reduce its content.

**E008**

Unexpected `{` or `[`;
Must follow either a space or a `$`, and if not a `$` a letter or a number must be found before.

*Example:*

Raised for:

```
lorem{
  ipsum: 1
}
```

Fix:

```
lorem {
  ipsum: 1
}
```

### Additional messages

Even if ukli doesn't intend on semantic validation of the configuration files, while checking the format it assume some rules, and will raise errors for a file that doesn't comply with.

**F001**

`Unbalanced '}'` or `Unbalanced ']'`

*Example:*

Raised at line 7:

```
section1 {
  option1 = value1
  option2 = value2
  option3 = value3
}

]
foo = 1
```

Fix:

```
section1 {
  option1 = value1
  option2 = value2
  option3 = value3
}

foo = 1
```

**F002**

Invalid assignation (multiple assignation operators)

*Example:*

Raised at line 3:

```
section1 {
  option1 = value1
  option2 =: value2
  option3 = value3
  option4 = value4
}
```

Fix:

```
section1 {
  option1 = value1
  option2 = value2
  option3 = value3
  option4 = value4
}
```

## Build

The project is built using [Go](https://golang.org/) 1.20+.

Then to execute the incremental build:

    go build

Run the tests:

    go test
