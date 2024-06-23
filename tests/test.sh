set -e

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


for i in $(ls -1 tests/* | grep -v '.sh'); do
    echo "Testing $i"
    for extension in json yaml toml xml tf hcl; do
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
        bin/qq ${case} $i
    done
done

