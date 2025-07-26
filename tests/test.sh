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
            if [[ "$input_ext" == "parquet" ]]; then
                # Parquet files are binary, use qq directly 
                command="bin/qq '$input_file' | bin/qq -o '$output_ext'"
            else
                command="cat '$input_file' | grep -v '#' | bin/qq -i '$input_ext' -o '$output_ext'"
            fi
            run_test "$test_name" "$command"
        done
        
        # Test embedded test cases (lines with # comments) - skip for binary files
        if [[ "$input_ext" != "parquet" ]] && grep -q "#" "$input_file" 2>/dev/null; then
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
    
    # Test jq pipeline conversions (excluding CSV)
    print "yellow" "Testing jq pipeline conversions..."
    local previous_ext="json"
    for file in $extensions; do
        local current_ext
        current_ext="${file##*.}"
        
        if [[ "$current_ext" == "csv" || "$current_ext" == "parquet" ]]; then
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

