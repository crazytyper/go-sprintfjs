package sprintfjs_test

import (
	"fmt"
	"regexp"
	"testing"

	"brainloop/pe/util/sprintfjs"
)

func TestFormat(t *testing.T) {
	pi := 3.141592653589793

	type testcase struct {
		Expected string
		Format string
		Args []interface{}
	}
	tc := func(expected, format string, args... interface{}) testcase {
		return testcase{expected,format,args}
	}

	testcases := []testcase {
		tc(`%`,`%%`),
		tc(`10`,`%b`, 2),
		tc(`A`,`%c`, 65),

		tc(`2`,`%d`, 2),
		tc(`2`,`%i`, 2),
	 	tc(`2`,`%d`, "2"),
	 	tc(`2`,`%i`, "2"),

		tc(`{"foo":"bar"}`,`%j`, map[string]interface{}{"foo": "bar"}),
		tc(`["foo","bar"]`,`%j`, []string{"foo", "bar"}),

		tc(`2e+0`,`%e`, 2),
		tc(`2`,`%u`, 2),
		tc(`4294967294`,`%u`, -2),

		tc(`2.2`,`%f`, 2.2),
		tc(`3.141592653589793`,`%g`, pi),

		tc(`10`,`%o`, 8),
	 	tc(`37777777770`,`%o`, -8),
		tc(`%s`,`%s`, "%s"),

		tc(`ff`,`%x`, 255),
	 	tc(`ffffff01`,`%x`, -255),
		tc(`FF`,`%X`, 255),
	 	tc(`FFFFFF01`,`%X`, -255),

		tc(`Polly wants a cracker`,`%2$s %3$s a %1$s`, "cracker", "Polly", "wants"),
		tc(`Hello world!`,`Hello %(who)s!`, map[string]interface{}{"who": "world"}),

		tc(`true`,`%t`, true),
		tc(`t`,`%.1t`, true),
		tc(`true`,`%t`, "true"),
		tc(`true`,`%t`, 1),
		tc(`false`,`%t`, false),
		tc(`f`,`%.1t`, false),
		tc(`false`,`%t`, ""),
		tc(`false`,`%t`, 0),

		tc(`null`,`%T`, nil),
		tc(`boolean`,`%T`, true),
		tc(`number`,`%T`, 42),
		tc(`string`,`%T`, "This is a string"),
		tc(`function`,`%T`, t.Fatal),
		tc(`array`,`%T`, []int{1, 2, 3}),
		tc(`object`,`%T`, map[string]interface{}{"foo": "bar"}),
		tc(`regexp`,`%T`, regexp.MustCompile(`<('[^']*'|'[^']*'|[^''>])*>`)),

		tc(`true`,`%v`, true),
		tc(`42`,`%v`, 42),
		tc(`This is a string`,`%v`, "This is a string"),
		tc(`[1 2 3]`,`%v`, []int{1, 2, 3}), // <- differs from sprintf.js
		tc(`map[foo:bar]`,`%v`, map[string]interface{}{"foo": "bar"}),// <- differs from sprintf.js
		tc(`<("[^"]*"|'[^']*'|[^'">])*>`,`%v`, regexp.MustCompile(`<("[^"]*"|'[^']*'|[^'">])*>`)),// <- differs from sprintf.js
		tc(`[1 2 3]`,`%v`, []int{1, 2, 3}),

		// sign
		tc(`2`,`%d`, 2),
		tc(`-2`,`%d`, -2),
		tc(`+2`,`%+d`, 2),
		tc(`-2`,`%+d`, -2),
		tc(`2`,`%i`, 2),
		tc(`-2`,`%i`, -2),
		tc(`+2`,`%+i`, 2),
		tc(`-2`,`%+i`, -2),
		tc(`2.2`,`%f`, 2.2),
		tc(`-2.2`,`%f`, -2.2),
		tc(`+2.2`,`%+f`, 2.2),
		tc(`-2.2`,`%+f`, -2.2),
		tc(`-2.3`,`%+.1f`, -2.34),
		tc(`-0.0`,`%+.1f`, -0.01),
		tc(`3.14159`,`%.6g`, pi),
		tc(`3.14`,`%.3g`, pi),
		tc(`3`,`%.1g`, pi),
		tc(`-000000123`,`%+010d`, -123),
		tc(`______-123`,"%+'_10d", -123),
		tc(`-234.34 123.2`,`%f %f`, -234.34, 123.2),

		// padding
		tc(`-0002`,`%05d`, -2),
		tc(`-0002`,`%05i`, -2),
		tc(`    <`,`%5s`, "<"),
		tc(`0000<`,`%05s`, "<"),
		tc(`____<`,"%'_5s", "<"),
		tc(`>    `,`%-5s`, ">"),
		tc(`>0000`,`%0-5s`, ">"),
		tc(`>____`,"%'_-5s", ">"),
		tc(`xxxxxx`,`%5s`, "xxxxxx"),
		tc(`1234`,`%02u`, 1234),
		tc(` -10.235`,`%8.3f`, -10.23456),
		tc(`-12.34 xxx`,`%f %s`, -12.34, "xxx"),
		tc("{\n  \"foo\": \"bar\"\n}",`%2j`, map[string]interface{}{"foo": "bar"}),
		tc("[\n  \"foo\",\n  \"bar\"\n]",`%2j`, []string{"foo", "bar"}),

		// precision
		tc(`2.3`,`%.1f`, 2.345),
		tc(`xxxxx`,`%5.5s`, "xxxxxx"),
		tc(`    x`,`%5.1s`, "xxxxxx"),
	}
	for i := range testcases {
		tc := testcases[i]
		t.Run(
			fmt.Sprintf("%s(%s)", tc.Expected, tc.Format),
			func(t *testing.T){
				actual, err := sprintfjs.Format(tc.Format, tc.Args...)
				if err != nil {
					t.Fatalf("%v", err)
				}
				if tc.Expected != actual {
					t.Fatalf("expected %q had %q", tc.Expected, actual)
				}
		})
	}
}

func TestFormatAST(t *testing.T) {
	ast, err := sprintfjs.Parse("Hello %(to)s!")
	if err != nil {
		t.Fatal(err)
	}

	actual, err := sprintfjs.FormatAST(
		ast,
		map[string]interface{}{
			"to": "world",
		})
	if err != nil {
		t.Fatal(err)
	}

	expected := "Hello world!"
	if expected != actual {
		t.Fatalf("Expected %q has %q", expected, actual)
	}
}
