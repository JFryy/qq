package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/JFryy/qq/codec"
	qqjson "github.com/JFryy/qq/codec/json"
	"github.com/JFryy/qq/codec/keyorder"
	"github.com/JFryy/qq/internal/tui"
	"github.com/goccy/go-json"
	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
)

func CreateRootCmd() *cobra.Command {
	var inputType, outputType, queryFile string
	var rawOutput bool
	var interactive bool
	var version bool
	var help bool
	var monochrome bool
	var stream bool
	var slurp bool
	var exitStatus bool
	var preserveKeyOrder bool
	encodings := strings.Join(codec.GetSupportedExtensions(), ", ")
	desc := fmt.Sprintf("qq is a interoperable configuration format transcoder with jq querying ability powered by gojq. qq is multi modal, and can be used as a replacement for jq or be interacted with via a repl with autocomplete and realtime rendering preview for building queries. Supported formats include %s", encodings)
	cmd := &cobra.Command{
		Use:   "qq [expression] [file] [flags] \n  cat [file] | qq [expression] [flags] \n  qq -I file \n  qq --from-file query.jq [file]",
		Short: "qq - JQ processing with conversions for popular configuration formats.",

		Long: desc,
		Run: func(cmd *cobra.Command, args []string) {
			if version {
				fmt.Println("qq version", resolveVersion())
				os.Exit(0)
			}
			if len(args) == 0 && !cmd.Flags().Changed("from-file") && !cmd.Flags().Changed("input") && !cmd.Flags().Changed("output") && !cmd.Flags().Changed("raw-input") && isTerminal(os.Stdin) {
				err := cmd.Help()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				os.Exit(0)
			}
			handleCommand(cmd, args, queryFile, inputType, outputType, rawOutput, help, interactive, monochrome, stream, slurp, exitStatus, preserveKeyOrder)
		},
	}
	cmd.Flags().StringVarP(&queryFile, "from-file", "f", "", "read the jq expression from a file")
	cmd.Flags().StringVarP(&inputType, "input", "i", "json", "specify input file type, only required on parsing stdin.")
	cmd.Flags().StringVarP(&outputType, "output", "o", "json", "specify output file type by extension name. This is inferred from extension if passing file position argument.")
	cmd.Flags().BoolVarP(&rawOutput, "raw-output", "r", false, "output strings without escapes and quotes.")
	cmd.Flags().BoolVarP(&help, "help", "h", false, "help for qq")
	cmd.Flags().BoolVarP(&version, "version", "v", false, "version for qq")
	cmd.Flags().BoolVarP(&interactive, "interactive", "I", false, "interactive mode for qq")
	cmd.Flags().BoolVarP(&monochrome, "monochrome-output", "M", false, "disable colored output")
	cmd.Flags().BoolVar(&stream, "stream", false, "parse input in streaming fashion, emitting path-value pairs (supports: json, jsonl, yaml, csv, tsv, line)")
	cmd.Flags().BoolVarP(&slurp, "slurp", "s", false, "read all inputs into an array and use it as the single input value")
	cmd.Flags().BoolVarP(&exitStatus, "exit-status", "e", false, "set exit status code based on the output")
	cmd.Flags().BoolVarP(&preserveKeyOrder, "preserve-key-order", "k", false, "keep input key order for objects the query passes through unchanged; objects it creates or modifies have sorted keys (input: json, jsonc, jsonl, yaml, xml, csv, tsv; output: json, jsonl, yaml, xml, csv, tsv)")

	cmd.MarkFlagsMutuallyExclusive("from-file", "interactive")

	return cmd
}

func handleCommand(cmd *cobra.Command, args []string, queryFile string, inputtype string, outputtype string, rawInput bool, help bool, interactive bool, monochrome bool, stream bool, slurp bool, exitStatus bool, preserveKeyOrder bool) {
	var input []byte
	var err error
	var filename string
	var inputReader io.Reader

	if help {
		val := CreateRootCmd().Help()
		fmt.Println(val)
		os.Exit(0)
	}

	// Validate: streaming and interactive modes are mutually exclusive
	if stream && interactive {
		fmt.Println("Error: --stream and --interactive flags cannot be used together")
		os.Exit(1)
	}

	// Validate: slurp and streaming are mutually exclusive
	if slurp && stream {
		fmt.Println("Error: --slurp and --stream flags cannot be used together")
		os.Exit(1)
	}

	// Validate: slurp and interactive are mutually exclusive
	if slurp && interactive {
		fmt.Println("Error: --slurp and --interactive flags cannot be used together")
		os.Exit(1)
	}

	// Validate: streaming already emits paths in document order, and
	// interactive mode re-encodes query results without recorded key order.
	if preserveKeyOrder && stream {
		fmt.Println("Error: --preserve-key-order and --stream flags cannot be used together")
		os.Exit(1)
	}
	if preserveKeyOrder && interactive {
		fmt.Println("Error: --preserve-key-order and --interactive flags cannot be used together")
		os.Exit(1)
	}

	expression := "."
	if cmd.Flags().Changed("from-file") {
		query, err := os.ReadFile(queryFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading jq expression: %v\n", err)
			os.Exit(1)
		}
		expression = string(query)
	} else if len(args) > 0 && (len(args) != 1 || !isFile(args[0])) {
		expression, args = args[0], args[1:]
	}

	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "Error: expected at most one input file")
		os.Exit(1)
	}
	if len(args) == 1 {
		filename = args[0]
		if stream {
			inputReader, err = os.Open(filename)
		} else {
			input, err = os.ReadFile(filename)
		}
	} else if stream {
		inputReader = os.Stdin
	} else {
		input, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var inputCodec codec.EncodingType
	// Check if -i flag was explicitly set by user
	inputFlagSet := cmd.Flags().Changed("input")

	if inputFlagSet {
		// -i flag takes precedence over file extension
		inputCodec, err = codec.GetEncodingType(inputtype)
	} else if filename != "" {
		// Infer from file extension when no -i flag is set
		inputCodec = inferFileType(filename)
	} else {
		inputCodec, err = codec.GetEncodingType(inputtype)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	outputCodec, err := codec.GetEncodingType(outputtype)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Handle streaming mode
	if stream {
		// For streaming, we need to read the input and parse it in streaming mode
		if inputReader == nil {
			// If we already read the input into memory, create a reader from it
			inputReader = bytes.NewReader(input)
		}

		// Execute streaming query
		query, err := gojq.Parse(expression)
		if err != nil {
			fmt.Printf("Error parsing jq expression: %v\n", err)
			os.Exit(1)
		}

		executeStreamingQuery(query, inputReader, inputCodec, outputCodec, rawInput, monochrome)

		// Close file if it was opened
		if file, ok := inputReader.(*os.File); ok && file != os.Stdin {
			file.Close()
		}
		os.Exit(0)
	}

	// Standard (non-streaming) mode
	var data any

	// Key order is recorded only while the input is decoded; output encoders
	// then read it.
	if preserveKeyOrder {
		keyorder.Enable()
	}
	if slurp {
		// Slurp mode: read multiple JSON values and combine into array
		data, err = slurpInputs(input, inputCodec)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		err = codec.Unmarshal(input, inputCodec, &data)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	keyorder.StopRecording()

	if !interactive {
		query, err := gojq.Parse(expression)
		if err != nil {
			fmt.Printf("Error parsing jq expression: %v\n", err)
			os.Exit(1)
		}

		exitCode := executeQuery(query, data, outputCodec, rawInput, monochrome, exitStatus)
		os.Exit(exitCode)
	}

	jsonBytes, err := codec.Marshal(data, codec.JSON)
	if err != nil {
		fmt.Println("Error preparing TUI buffer:", err)
		os.Exit(1)
	}

	queryExpr, ok, err := tui.Interact(string(jsonBytes))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Status messages go to stderr so stdout holds only the query result.
	if !ok {
		fmt.Fprintln(os.Stderr, "\033[33mExited without executing query\033[0m")
		os.Exit(0)
	}

	query, err := gojq.Parse(queryExpr)
	if err != nil {
		fmt.Printf("Error parsing selected jq expression: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\033[36m# Query: %s\033[0m\n", queryExpr)
	exitCode := executeQuery(query, data, outputCodec, rawInput, monochrome, exitStatus)
	os.Exit(exitCode)
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

var extensionMap = codec.GetExtensionMap()

func inferFileType(fName string) codec.EncodingType {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fName)), ".")

	if encType, ok := extensionMap[ext]; ok {
		return encType
	}
	return codec.JSON
}

func executeQuery(query *gojq.Query, data any, fileType codec.EncodingType, rawOut bool, monochrome bool, exitStatus bool) int {
	iter := query.Run(data)
	var lastValue any
	hasOutput := false

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			fmt.Printf("Error executing jq expression: %v\n", err)
			return 1
		}

		hasOutput = true
		lastValue = v

		b, err := codec.Marshal(v, fileType)
		if err != nil {
			fmt.Printf("Error formatting result: %v\n", err)
			return 1
		}

		if codec.IsBinaryFormat(fileType) {
			// For binary formats, write directly to stdout as raw bytes
			os.Stdout.Write(b)
		} else {
			s := string(b)
			r, _ := codec.PrettyFormat(s, fileType, rawOut, monochrome)
			fmt.Println(r)
		}
	}

	// Handle exit status flag
	if exitStatus {
		if !hasOutput {
			return 4 // No output
		}
		// Check if last value is false or null
		if lastValue == false || lastValue == nil {
			return 1
		}
	}

	return 0
}

func slurpInputs(input []byte, inputCodec codec.EncodingType) (any, error) {
	var values []any

	switch inputCodec {
	case codec.JSON:
		// JSON can have multiple whitespace-separated values
		parsed, err := qqjson.UnmarshalAll(input)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON: %v", err)
		}
		values = append(values, parsed...)

	case codec.JSONL:
		// JSONL already parses to array of values
		var data any
		if err := codec.Unmarshal(input, inputCodec, &data); err != nil {
			return nil, err
		}
		// JSONL codec returns []any, use it directly
		if arr, ok := data.([]any); ok {
			return arr, nil
		}
		values = append(values, data)

	case codec.YAML:
		// YAML can have multiple documents separated by ---
		// The yaml codec already handles this and returns an array for multi-doc
		var data any
		if err := codec.Unmarshal(input, inputCodec, &data); err != nil {
			return nil, err
		}
		// If it's already an array (multi-doc YAML), use it
		if arr, ok := data.([]any); ok {
			return arr, nil
		}
		// Single document, wrap it
		values = append(values, data)

	case codec.LINE, codec.TXT:
		// Line-based formats already return array of lines
		var data any
		if err := codec.Unmarshal(input, inputCodec, &data); err != nil {
			return nil, err
		}
		if arr, ok := data.([]any); ok {
			return arr, nil
		}
		values = append(values, data)

	default:
		// For other formats, parse as single value and wrap
		var data any
		if err := codec.Unmarshal(input, inputCodec, &data); err != nil {
			return nil, err
		}
		values = append(values, data)
	}

	return values, nil
}

func executeStreamingQuery(query *gojq.Query, reader io.Reader, inputType codec.EncodingType, outputType codec.EncodingType, rawOut bool, monochrome bool) {
	// Parse input in streaming mode (emits path-value pairs via channels)
	dataChan, errChan := codec.StreamParser(reader, inputType)

	// Process stream elements as they arrive
	for {
		select {
		case streamElement, ok := <-dataChan:
			if !ok {
				// Channel closed, check for errors
				select {
				case err := <-errChan:
					if err != nil {
						fmt.Printf("Error parsing stream: %v\n", err)
						os.Exit(1)
					}
				default:
				}
				return
			}

			// Execute query on this stream element and output immediately
			iter := query.Run(streamElement)
			for {
				v, ok := iter.Next()
				if !ok {
					break
				}
				if err, ok := v.(error); ok {
					fmt.Printf("Error executing jq expression: %v\n", err)
					os.Exit(1)
				}

				// Stream elements are always emitted as compact JSON (one line
				// per element), matching jq --stream behaviour regardless of -o.
				b, err := json.Marshal(v)
				if err != nil {
					fmt.Printf("Error formatting result: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(string(b))
			}

		case err := <-errChan:
			if err != nil {
				fmt.Printf("Error parsing stream: %v\n", err)
				os.Exit(1)
			}
		}
	}
}
