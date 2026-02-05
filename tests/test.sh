#!/bin/bash

set -eo pipefail

# Global counters
total_tests=0
passed_tests=0
skipped_tests=0
failed_tests=0

# Cleanup function
cleanup() {
    if [[ $failed_tests -gt 0 ]]; then
        print "red" "Tests failed! $failed_tests failures out of $total_tests tests"
        exit 1
    fi
}
trap cleanup EXIT

# Validation function
validate_prerequisites() {
    local missing_deps=()
    
    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
    fi
    
    if [[ ! -f "bin/qq" ]]; then
        print "red" "Error: bin/qq not found. Run 'make build' first."
        exit 1
    fi
    
    if [[ ! -d "tests" ]]; then
        print "red" "Error: tests directory not found"
        exit 1
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print "red" "Error: Missing dependencies: ${missing_deps[*]}"
        print "red" "Please install the missing dependencies and try again."
        exit 1
    fi
}

# Enhanced print function
print() {
    local color="$1"
    local message="$2"
    case $color in
        red)
            echo -e "\033[0;31m$message\033[0m" >&2
            ;;
        green)
            echo -e "\033[0;32m$message\033[0m"
            ;;
        yellow)
            echo -e "\033[0;33m$message\033[0m"
            ;;
        blue)
            echo -e "\033[0;34m$message\033[0m"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

# Skip logic function
should_skip_conversion() {
    local input="$1"
    local output="$2"
    local reason=""
    
    # CSV compatibility rules
    if [[ "$input" == "csv" && "$output" != "csv" ]]; then
        reason="CSV to non-CSV conversion not supported"
    elif [[ "$input" != "csv" && "$output" == "csv" ]]; then
        reason="Non-CSV to CSV conversion not supported"
    # Parquet compatibility rules
    elif [[ "$output" == "parquet" && "$input" != "parquet" ]]; then
        reason="Non-parquet to parquet conversion not supported"
    elif [[ "$input" == "parquet" && "$output" != "parquet" ]]; then
        reason="Parquet to non-parquet conversion not supported"
    # Nested structure rules  
    elif [[ "$input" == "proto" && "$output" == "env" ]]; then
        reason="Proto to env conversion not supported (nested structures)"
    elif [[ "$output" == "env" && "$input" != "env" ]]; then
        reason="Complex structures to env conversion not supported"
    fi
    
    if [[ -n "$reason" ]]; then
        echo "$reason"
        return 0
    fi
    return 1
}

# Test execution wrapper
run_test() {
    local test_name="$1"
    local command="$2"
    
    total_tests=$((total_tests + 1))
    print "blue" "[$total_tests] Testing: $test_name"
    
    local exit_code=0
    eval "$command" &>/dev/null || exit_code=$?
    
    if [[ $exit_code -eq 0 ]]; then
        passed_tests=$((passed_tests + 1))
        print "green" "  ✓ PASS"
    else
        failed_tests=$((failed_tests + 1))
        print "red" "  ✗ FAIL: $command"
        print "red" "    Command failed with exit code $exit_code"
    fi
}

# Test gojq functionality
test_gojq_functionality() {
    print "yellow" "Testing gojq functionality..."

    # Basic queries
    run_test "identity query" \
        "echo '{\"a\":1}' | bin/qq '.'"

    run_test "key access" \
        "echo '{\"name\":\"test\"}' | bin/qq '.name' | grep -q 'test'"

    run_test "nested key access" \
        "echo '{\"a\":{\"b\":\"value\"}}' | bin/qq '.a.b' | grep -q 'value'"

    run_test "array index" \
        "echo '[1,2,3]' | bin/qq '.[1]' | grep -q '^2$'"

    run_test "array slice" \
        "echo '[1,2,3,4,5]' | bin/qq '.[1:3]' | jq -e '. == [2,3]'"

    # Array operations
    run_test "array length" \
        "echo '[1,2,3]' | bin/qq 'length' | grep -q '^3$'"

    run_test "map operation" \
        "echo '[1,2,3]' | bin/qq 'map(. * 2)' | jq -e '. == [2,4,6]'"

    run_test "select filter" \
        "echo '[{\"a\":1},{\"a\":2},{\"a\":3}]' | bin/qq '[.[] | select(.a > 1)]' | jq -e 'length == 2'"

    run_test "array iteration" \
        "echo '[1,2,3]' | bin/qq '.[]' | wc -l | grep -q '3'"

    # Object operations
    run_test "keys function" \
        "echo '{\"b\":2,\"a\":1}' | bin/qq 'keys' | jq -e '. == [\"a\",\"b\"]'"

    run_test "values function" \
        "echo '{\"a\":1,\"b\":2}' | bin/qq '[.[] ]' | jq -e 'length == 2'"

    run_test "has function" \
        "echo '{\"a\":1}' | bin/qq 'has(\"a\")' | grep -q 'true'"

    # Pipes and composition
    run_test "pipe operations" \
        "echo '[1,2,3,4,5]' | bin/qq 'map(. * 2) | map(. + 1)' | jq -e '.[0] == 3'"

    run_test "multiple filters" \
        "echo '[{\"a\":1},{\"a\":2},{\"a\":3}]' | bin/qq '[.[] | select(.a > 1) | .a]' | jq -e '. == [2,3]'"

    # String operations
    run_test "string concatenation" \
        "echo '{\"a\":\"hello\",\"b\":\"world\"}' | bin/qq '.a + \" \" + .b' | grep -q 'hello world'"

    run_test "string split" \
        "echo '\"a,b,c\"' | bin/qq 'split(\",\")' | jq -e 'length == 3'"

    # Math operations
    run_test "addition" \
        "echo '{\"a\":5,\"b\":3}' | bin/qq '.a + .b' | grep -q '^8$'"

    run_test "arithmetic expression" \
        "echo '[1,2,3,4,5]' | bin/qq 'map(. * 2) | add' | grep -q '^30$'"

    # Type operations
    run_test "type function" \
        "echo '42' | bin/qq 'type' | grep -q 'number'"

    run_test "type check string" \
        "echo '\"test\"' | bin/qq 'type' | grep -q 'string'"

    # Conditionals
    run_test "if-then-else" \
        "echo '5' | bin/qq 'if . > 3 then \"big\" else \"small\" end' | grep -q 'big'"

    # Null handling
    run_test "null coalescing" \
        "echo '{\"a\":null}' | bin/qq '.a // \"default\"' | grep -q 'default'"

    # Sorting
    run_test "sort array" \
        "echo '[3,1,2]' | bin/qq 'sort' | jq -e '. == [1,2,3]'"

    run_test "sort reverse" \
        "echo '[1,2,3]' | bin/qq 'sort | reverse' | jq -e '. == [3,2,1]'"

    # Group by
    run_test "group_by operation" \
        "echo '[{\"k\":\"a\",\"v\":1},{\"k\":\"b\",\"v\":2},{\"k\":\"a\",\"v\":3}]' | bin/qq 'group_by(.k) | length' | grep -q '^2$'"

    # Min/Max
    run_test "max function" \
        "echo '[1,5,3,2,4]' | bin/qq 'max' | grep -q '^5$'"

    run_test "min function" \
        "echo '[1,5,3,2,4]' | bin/qq 'min' | grep -q '^1$'"
}

# Test streaming functionality (jq-compatible streaming)
test_streaming_functionality() {
    print "yellow" "Testing streaming functionality..."

    # Basic streaming with simple object
    run_test "streaming simple object" \
        "echo '{\"name\":\"test\",\"id\":1}' | bin/qq --stream | grep -q 'name'"

    # Streaming with array
    run_test "streaming array" \
        "echo '[1,2,3]' | bin/qq --stream | grep -q '0'"

    # Streaming with filter - select only path-value pairs (length == 2)
    run_test "streaming with filter" \
        "echo '{\"a\":1,\"b\":2}' | bin/qq --stream 'select(length == 2)' | grep -q '\"a\"'"

    # Streaming with nested structure
    run_test "streaming nested structure" \
        "echo '{\"user\":{\"name\":\"Bob\"}}' | bin/qq --stream | grep -q 'user'"

    # Streaming from file
    local test_file="/tmp/qq_stream_test.json"
    echo '{"test":123}' > "$test_file"
    run_test "streaming from file" \
        "bin/qq --stream '.' $test_file | grep -q 'test'"

    # Streaming + interactive validation (should fail)
    local exit_code=0
    echo '{}' | bin/qq --stream --interactive &>/dev/null || exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        passed_tests=$((passed_tests + 1))
        total_tests=$((total_tests + 1))
        print "green" "  ✓ PASS: streaming + interactive correctly rejected"
    else
        failed_tests=$((failed_tests + 1))
        total_tests=$((total_tests + 1))
        print "red" "  ✗ FAIL: streaming + interactive should fail"
    fi

    # Cleanup
    rm -f "$test_file"
}

# Test slurp functionality
test_slurp_functionality() {
    print "yellow" "Testing slurp functionality..."

    # Slurp multiple JSON values
    run_test "slurp multiple JSON values" \
        "echo -e '{\"id\":1}\n{\"id\":2}\n{\"id\":3}' | bin/qq -s 'length' | grep -q '3'"

    # Slurp with transformation
    run_test "slurp with map" \
        "echo -e '{\"id\":1}\n{\"id\":2}' | bin/qq -s 'map(.id) | add' | grep -q '3'"

    # Slurp JSONL file
    local test_file="/tmp/qq_slurp_test.jsonl"
    echo -e '{"name":"Alice"}\n{"name":"Bob"}' > "$test_file"
    run_test "slurp JSONL file" \
        "bin/qq -s -i jsonl 'map(.name) | join(\", \")' $test_file | grep -q 'Alice, Bob'"

    # Slurp + stream validation (should fail)
    local exit_code=0
    echo '{}' | bin/qq -s --stream &>/dev/null || exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        passed_tests=$((passed_tests + 1))
        total_tests=$((total_tests + 1))
        print "green" "  ✓ PASS: slurp + stream correctly rejected"
    else
        failed_tests=$((failed_tests + 1))
        total_tests=$((total_tests + 1))
        print "red" "  ✗ FAIL: slurp + stream should fail"
    fi

    # Cleanup
    rm -f "$test_file"
}

# Test exit-status functionality
test_exit_status_functionality() {
    print "yellow" "Testing exit-status functionality..."

    # Exit status with true value (should succeed)
    local exit_code=0
    echo '{"active":true}' | bin/qq -e '.active' > /dev/null 2>&1 || exit_code=$?
    if [[ $exit_code -eq 0 ]]; then
        passed_tests=$((passed_tests + 1))
        total_tests=$((total_tests + 1))
        print "green" "  ✓ PASS: exit-status returns 0 for true"
    else
        failed_tests=$((failed_tests + 1))
        total_tests=$((total_tests + 1))
        print "red" "  ✗ FAIL: exit-status should return 0 for true, got $exit_code"
    fi

    # Exit status with false value (should fail with 1)
    exit_code=0
    echo '{"active":false}' | bin/qq -e '.active' > /dev/null 2>&1 || exit_code=$?
    if [[ $exit_code -eq 1 ]]; then
        passed_tests=$((passed_tests + 1))
        total_tests=$((total_tests + 1))
        print "green" "  ✓ PASS: exit-status returns 1 for false"
    else
        failed_tests=$((failed_tests + 1))
        total_tests=$((total_tests + 1))
        print "red" "  ✗ FAIL: exit-status should return 1 for false, got $exit_code"
    fi

    # Exit status with no output (should fail with 4)
    exit_code=0
    echo '5' | bin/qq -e 'select(. > 10)' > /dev/null 2>&1 || exit_code=$?
    if [[ $exit_code -eq 4 ]]; then
        passed_tests=$((passed_tests + 1))
        total_tests=$((total_tests + 1))
        print "green" "  ✓ PASS: exit-status returns 4 for no output"
    else
        failed_tests=$((failed_tests + 1))
        total_tests=$((total_tests + 1))
        print "red" "  ✗ FAIL: exit-status should return 4 for no output, got $exit_code"
    fi

    # Exit status in conditional
    run_test "exit-status in conditional" \
        "echo '{\"ready\":true}' | bin/qq -e '.ready' && echo 'success' | grep -q 'success'"
}

# Progress summary
print_summary() {
    echo
    print "blue" "=== Test Summary ==="
    print "green" "Passed: $passed_tests"
    print "yellow" "Skipped: $skipped_tests"
    print "red" "Failed: $failed_tests"
    print "blue" "Total: $total_tests"
    echo
}

# Main execution
main() {
    print "blue" "Starting qq codec tests..."
    echo

    validate_prerequisites

    # Run gojq functionality tests
    test_gojq_functionality
    echo

    # Run streaming functionality tests
    test_streaming_functionality
    echo

    # Run slurp functionality tests
    test_slurp_functionality
    echo

    # Run exit-status functionality tests
    test_exit_status_functionality
    echo

    # Get test extensions, excluding shell scripts and ini files
    local extensions
    if ! extensions=$(find tests -maxdepth 1 -type f ! -name "*.sh" ! -name "*.ini" 2>/dev/null); then
        print "red" "Error: Failed to find test files"
        exit 1
    fi
    
    if [[ -z "$extensions" ]]; then
        print "yellow" "Warning: No test files found in tests/ directory"
        exit 0
    fi
    
    # Test all format conversions
    for input_file in $extensions; do
        local input_ext=""
        input_ext="${input_file##*.}"
        
        # Skip if we can't determine extension
        if [[ -z "$input_ext" || "$input_ext" == "$input_file" ]]; then
            continue
        fi
        
        print "yellow" "Testing conversions from $input_ext format..."
        
        for output_file in $extensions; do
            local output_ext=""
            output_ext="${output_file##*.}"
            
            # Skip if we can't determine extension
            if [[ -z "$output_ext" || "$output_ext" == "$output_file" ]]; then
                continue
            fi
            
            # Check if conversion should be skipped
            local skip_reason=""
            if skip_reason=$(should_skip_conversion "$input_ext" "$output_ext"); then
                skipped_tests=$((skipped_tests + 1))
                print "yellow" "  -> Skipping $input_ext->$output_ext: $skip_reason"
                continue
            fi
            
            # Run the conversion test
            local test_name="$input_ext -> $output_ext"
            local command
            if [[ "$input_ext" == "parquet" || "$input_ext" == "msgpack" ]]; then
                # Parquet files are binary, use qq directly 
                command="bin/qq '$input_file' | bin/qq -o '$output_ext'"
            else
                command="cat '$input_file' | grep -v '#' | bin/qq -i '$input_ext' -o '$output_ext'"
            fi
            run_test "$test_name" "$command"
        done
        
        # Test embedded test cases (lines with # comments) - skip for binary files
        if [[ "$input_ext" != "parquet" || "$input_ext" != "msgpack" ]] && grep -q "#" "$input_file" 2>/dev/null; then
            # Process each line with a # comment individually
            while IFS= read -r line; do
                if [[ "$line" =~ ^#[[:space:]]*(.+)$ ]]; then
                    local case="${BASH_REMATCH[1]}"
                    # Remove leading/trailing whitespace
                    case=$(echo "$case" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
                    if [[ -n "$case" ]]; then
                        local test_name="$input_ext embedded case: $case"
                        local command="cat '$input_file' | grep -v '^#' | bin/qq -i '$input_ext' | jq $case"
                        run_test "$test_name" "$command"
                    fi
                fi
            done < "$input_file"
        fi
    done
    
    # Test jq pipeline conversions (excluding codecs with constrained structures/binary formats)
    print "yellow" "Testing jq pipeline conversions..."
    local previous_ext="json"
    for file in $extensions; do
        local current_ext
        current_ext="${file##*.}"
        
        if [[ "$current_ext" == "csv" || "$current_ext" == "parquet" || "$current_ext" == "msgpack" ]]; then
            continue
        fi
        
        # Skip if target format doesn't support complex structures
        local skip_reason=""
        if skip_reason=$(should_skip_conversion "$current_ext" "$previous_ext"); then
            skipped_tests=$((skipped_tests + 1))
            continue
        fi
        
        local test_name="jq pipeline: $current_ext -> $previous_ext"
        local command="cat '$file' | grep -v '^#' | bin/qq -i '$current_ext' | jq . | bin/qq -o '$previous_ext'"
        run_test "$test_name" "$command"
        
        previous_ext="$current_ext"
    done
    
    print_summary
    
    if [[ $failed_tests -eq 0 ]]; then
        print "green" "All tests passed! 🎉"
    fi
}

main "$@"

