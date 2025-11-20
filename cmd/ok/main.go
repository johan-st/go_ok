package main

import (
	"flag"
	"strings"

	"github.com/johan-st/gook/examples"
)

func main() {
	runValues, keys := getRunFuncs()
	example := flag.String("run", "", "the example to run. (available: "+strings.Join(keys,", ")+")")
	testString := flag.String("on", "test@example.com", "the string to test on the example")
	flag.Parse()

	if _, ok := runValues[*example]; !ok {
		flag.PrintDefaults()
		return
	}
	runValues[*example](*testString)
}

func getRunFuncs() (map[string]func(string), []string) {
	funcRunBasicExamples := func(string) { examples.RunBasicExamples() }
	funcEmail := func(testString string) { examples.Email(testString) }
	funcNumeric := func(testString string) { examples.Numeric(testString) }
	funcWip := func(testString string) { examples.Wip(testString) }

	runValues := map[string]func(string){
		"basic": funcRunBasicExamples,
		"email": funcEmail,
		"num": funcNumeric,
		"wip": funcWip,
	}
	keys := make([]string, 0, len(runValues))
	for k := range runValues {
		keys = append(keys, k)
	}
	return runValues, keys
}