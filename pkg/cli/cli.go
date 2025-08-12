package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type ArgMap map[string]string

type ArgList struct {
	Args ArgMap
}

func (a *ArgList) String() string {
	return fmt.Sprintf("%v", a.Args)
}

func (a *ArgList) Set(value string) error {
	return fmt.Errorf("use custom parsing")
}

// Inspired by https://github.com/tmccombs/hcl2json/blob/main/main.go
// Allow for input to come from STDIN or multiple files provided
// as program arguments. If withQuery is true, assume first arg
// is the jq query to be used
func CommandLineArgsBuffer(withQuery bool) (ArgMap, *bytes.Buffer, string, error) {

	args := os.Args[1:]
	argMap := make(ArgMap)
	rest := []string{}
	i := 0
	for i < len(args) {
		if args[i] == "--arg" && i+2 < len(args) {
			key := args[i+1]
			val := args[i+2]
			argMap[key] = val
			i += 3
		} else {
			rest = append(rest, args[i])
			i++
		}
	}
	files := []string{}
	query := ""
	if withQuery {
		if len(rest) == 0 {
			return argMap, nil, "", fmt.Errorf("missing jq query")
		}
		query = rest[0]
		files = rest[1:]
	}
	var inputName string
	switch len(files) {
	case 0:
		files = append(files, "-")
		inputName = "STDIN"
	case 1:
		inputName = files[0]
		if inputName == "-" {
			inputName = "STDIN"
		}
	default:
		inputName = "COMPOSITE"
	}
	buffer := bytes.NewBuffer([]byte{})
	for _, filename := range files {
		var stream io.Reader
		if filename == "-" {
			stream = os.Stdin
			filename = "STDIN" // for better error message
		} else {
			file, err := os.Open(filename)
			if err != nil {
				return argMap, nil, "", fmt.Errorf("Failed to open %s: %s\n", filename, err)
			}
			defer file.Close()
			stream = file
		}
		_, err := buffer.ReadFrom(stream)
		if err != nil {
			return argMap, nil, "", fmt.Errorf("Failed to read from %s: %s\n", filename, err)
		}
		buffer.WriteByte('\n') // just in case it doesn't have an ending newline
	}
	return argMap, buffer, query, nil
}
