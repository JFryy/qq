set -e

if [ -z $(which jq) ]; then
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


extensions=$(ls -1 tests/* | grep -Ev '.sh|csv')
for i in ${extensions}; do
    echo "Testing $i"
    for f in ${extensions}; do
        extension=$(echo $f | cut -d. -f2)
        input=$(echo $i | cut -d. -f2)
        # csv is not supported for toml and xml output
        case $input in
            csv)
                if [ $(echo $extension | grep -E 'toml') ]; then
                    continue
                fi
                ;;
        esac
        print "" "============================================"
        print "" "Executing: cat $i | grep -v '#' | bin/qq -i $input -o $extension"
        print "" "============================================"
        cat $i | grep -v "#" | bin/qq -i $(echo $i | cut -d. -f2) -o $extension
        print "green" "============================================"
        print "green" "Success."
        print "green" "============================================"
    done

    test_cases=$(cat $i | grep "#" | cut -d# -f2)
    for case in ${test_cases}; do
        print "" "============================================"
        print "yellow" "Testing case: qq $case $i"
        print "" "============================================"
        echo $test_cases
        cat $i | grep -v \# | bin/qq ${case} $i
    done
done

# conversions to jq and back
previous_ext="json"
for file in ${extensions}; do
    print "" "============================================"
    print "" "Executing: cat tests/test.xml | jq . | bin/qq -o $extension"
    print "" "============================================"
    bin/qq tests/test.xml | jq . | bin/qq -o $previous_ext
    print "green" "============================================"
    print "green" "Success."
    print "green" "============================================"
    previous_ext=$(echo $file | cut -d. -f2)
done
