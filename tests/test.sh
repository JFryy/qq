#!/bin/bash

set -e

if [ -z "$(which jq)" ]; then
    echo "jq is not installed. Please install jq."
    exit 1
fi

print() {
    case $1 in
        red)
            echo -e "\033[0;31m$2\033[0m"
            ;;
        green)
            echo -e "\033[0;32m$2\033[0m"
            ;;
        yellow)
            echo -e "\033[0;33m$2\033[0m"
            ;;
        *)
            echo -e "\033[0;33m$2\033[0m"
            ;;
    esac
}

extensions=$(ls -1 tests/* | grep -Ev '.sh|ini')
for i in ${extensions}; do
    echo "Testing $i"
    input=$(echo $i | cut -d. -f2)

    for f in ${extensions}; do
        extension=$(echo $f | cut -d. -f2)

        if [[ "$input" == "csv" && "$extension" != "csv" ]]
        then
            print "yellow" "Skipping unsupported conversion from CSV to non-CSV compatible structure"
            continue
        fi

        if [[ "$input" != csv && $extension == "csv" ]]
        then
            print "yellow" "Skipping unsupported conversion from CSV to non-CSV compatible structure"
            continue
        fi

        print "" "============================================"
        print "" "Executing: cat $i | grep -v '#' | bin/qq -i $input -o $extension"
        print "" "============================================"
        cat "$i" | grep -v "#" | bin/qq -i "$input" -o "$extension"
        print "green" "============================================"
        print "green" "Success."
        print "green" "============================================"
    done

    test_cases=$(cat $i | grep "#" | cut -d# -f2)
    for case in ${test_cases}; do
        print "" "============================================"
        print "yellow" "Testing case: qq $case $i"
        print "" "============================================"
        cat "$i" | grep -v \# | bin/qq "${case}" "$i"
    done
done

previous_ext="json"
for file in ${extensions}; do
    if [[ $(echo -n $file | grep csv) ]]
    then
        continue
    fi
    print "" $file
    print "" "============================================"
    print "" "Executing: cat $file | jq . | bin/qq -o $previous_ext"
    print "" "============================================"
    bin/qq "$file" | jq . | bin/qq -o "$previous_ext"
    print "green" "============================================"
    print "green" "Success."
    print "green" "============================================"
    previous_ext=$(echo "$file" | cut -d. -f2)
done

