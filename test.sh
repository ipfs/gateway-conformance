#! /usr/bin/env bash
set -x

make test-kubo; mv output.json output-kubo.json
cat output-kubo.json | jq --raw-output 'select(.Action == "run") | .Package + " - " + .Test' | sort > test-kubo.run

# if test-kubo.run is empty, something went wrong
if [ ! -s test-kubo.run ]; then
    echo "test-kubo.run is empty"
    exit 1
fi

make test-randomizer; mv output.json output-randomizer.json
cat output-randomizer.json | jq --raw-output 'select(.Action == "run") | .Package + " - " + .Test' | sort > test-randomizer.run
cat output-randomizer.json | jq --raw-output 'select(.Action == "pass") | .Package + " - " + .Test' | sort > test-randomizer.pass

# detect tests that are not running during randomizer
# if there is a difference, something went wrong
diff test-kubo.run test-randomizer.run

if [ $? -ne 0 ]; then
    echo "test-kubo.run and test-randomizer.run are different"
    exit 1
fi

# run randomizer test 5 times and detect the tests that are always passing:
cp test-randomizer.pass always_passing

for i in {1..3}; do
    # if always_passing is empty -> we're done.
    if [ ! -s always_passing ]; then
        break
    fi

    make test-randomizer; mv output.json output-randomizer.json
    cat output-randomizer.json | jq --raw-output 'select(.Action == "pass") | .Package + " - " + .Test' | sort > test-randomizer.pass
    comm -12 always_passing test-randomizer.pass > temp.log
    mv temp.log always_passing
done

# if always_passing is not empty, something went wrong:
if [ -s always_passing ]; then
    echo "always_passing is not empty"
    exit 1
else 
    echo "all tests failed at least once :clap:"
    exit 0
fi
