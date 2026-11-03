package cli

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JFryy/qq/codec"
	"github.com/itchyny/gojq"
)

type splitFile struct {
	name string
	data []byte
}

// writeSplitResults runs query on data and writes each result to its own file
// in dir, named by evaluating nameQuery on the result and adding the output
// format's extension. All names are validated and all results encoded before
// any file is written. Existing files are kept unless overwrite is set.
func writeSplitResults(query, nameQuery *gojq.Query, data any, fileType codec.EncodingType, dir string, overwrite bool) error {
	ext := codec.Codecs[fileType].Extensions[0]

	var files []splitFile
	seen := make(map[string]bool)
	iter := query.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return fmt.Errorf("error executing jq expression: %v", err)
		}

		name, err := splitFileName(nameQuery, v)
		if err != nil {
			return err
		}
		name += "." + ext
		if seen[name] {
			return fmt.Errorf("split expression produced duplicate file name %q", name)
		}
		seen[name] = true

		b, err := codec.Marshal(v, fileType)
		if err != nil {
			return fmt.Errorf("error formatting result: %v", err)
		}
		if !codec.IsBinaryFormat(fileType) && !strings.HasSuffix(string(b), "\n") {
			b = append(b, '\n')
		}
		files = append(files, splitFile{name: name, data: b})
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if overwrite {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	for _, f := range files {
		path := filepath.Join(dir, f.name)
		file, err := os.OpenFile(path, flags, 0o644)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				return fmt.Errorf("%s already exists, use --overwrite to replace it", path)
			}
			return err
		}
		_, err = file.Write(f.data)
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// splitFileName evaluates nameQuery on v and returns the resulting base name.
// The expression must produce exactly one string or number, and the name must
// stay inside the output directory.
func splitFileName(nameQuery *gojq.Query, v any) (string, error) {
	iter := nameQuery.Run(v)
	result, ok := iter.Next()
	if !ok {
		return "", fmt.Errorf("split expression produced no value")
	}
	if err, ok := result.(error); ok {
		return "", fmt.Errorf("error executing split expression: %v", err)
	}
	if _, ok := iter.Next(); ok {
		return "", fmt.Errorf("split expression produced more than one value")
	}

	var name string
	switch n := result.(type) {
	case string:
		name = n
	case int:
		name = strconv.Itoa(n)
	case float64:
		name = strconv.FormatFloat(n, 'f', -1, 64)
	case *big.Int:
		name = n.String()
	default:
		return "", fmt.Errorf("split expression must produce a string or number, got %s", jqTypeName(result))
	}

	switch {
	case name == "":
		return "", fmt.Errorf("split expression produced an empty file name")
	case name == "." || name == ".." || strings.ContainsAny(name, `/\`+"\x00") ||
		filepath.IsAbs(name) || filepath.Clean(name) != name || !filepath.IsLocal(name):
		return "", fmt.Errorf("split expression produced invalid file name %q", name)
	}
	return name, nil
}

func jqTypeName(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("%T", v)
	}
}
