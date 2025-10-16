package lept

import (
	"fmt"
	"strconv"
	"testing"
)

func TestBaseType(t *testing.T) {
	t.Run("null", func(t *testing.T) {
		v, err := Parse("null")
		if err != nil {
			t.Fatal("test parse null failed", err)
		}

		assertValue(t, v.typ, LEPT_NULL)
	})

	t.Run("true", func(t *testing.T) {
		v, err := Parse("true")
		if err != nil {
			t.Fatal("test parse true failed", err)
		}

		assertValue(t, v.typ, LEPT_TRUE)
	})

	t.Run("false", func(t *testing.T) {
		v, err := Parse("false")
		if err != nil {
			t.Error("test parse false failed", err)
		}

		assertValue(t, v.typ, LEPT_FALSE)
	})

	t.Run("number", func(t *testing.T) {
		testNumber(t, 0.0, "0")
		testNumber(t, 0.0, "-0")
		testNumber(t, 0.0, "-0.0")
		testNumber(t, 1.0, "1")
		testNumber(t, -1.0, "-1")
		testNumber(t, 1.5, "1.5")
		testNumber(t, -1.5, "-1.5")
		testNumber(t, 3.1416, "3.1416")
		testNumber(t, 1e10, "1E10")
		testNumber(t, 1e10, "1e10")
		testNumber(t, 1e+10, "1E+10")
		testNumber(t, 1e-10, "1E-10")
		testNumber(t, -1e10, "-1E10")
		testNumber(t, -1e10, "-1e10")
		testNumber(t, -1e+10, "-1E+10")
		testNumber(t, -1e-10, "-1E-10")
		testNumber(t, 1.234e+10, "1.234E+10")
		testNumber(t, 1.234e-10, "1.234E-10")
		testNumber(t, 0.0, "1e-10000") /* must underflow */

		testNumber(t, 1.0000000000000002, "1.0000000000000002")           /* the smallest number > 1 */
		testNumber(t, 4.9406564584124654e-324, "4.9406564584124654e-324") /* minimum denormal */
		testNumber(t, -4.9406564584124654e-324, "-4.9406564584124654e-324")
		testNumber(t, 2.2250738585072009e-308, "2.2250738585072009e-308") /* Max subnormal double */
		testNumber(t, -2.2250738585072009e-308, "-2.2250738585072009e-308")
		testNumber(t, 2.2250738585072014e-308, "2.2250738585072014e-308") /* Min normal positive double */
		testNumber(t, -2.2250738585072014e-308, "-2.2250738585072014e-308")
		testNumber(t, 1.7976931348623157e+308, "1.7976931348623157e+308") /* Max double */
		testNumber(t, -1.7976931348623157e+308, "-1.7976931348623157e+308")
	})

	t.Run("string", func(t *testing.T) {
		testString(t, ``, "\"\"")
		testString(t, `Hello`, "\"Hello\"")
		testString(t, `Hello\nWorld`, "\"Hello\\nWorld\"")
	})
}

func testNumber(t *testing.T, want float64, number string) {
	v, err := Parse(number)
	if err != nil {
		t.Fatal("test parse number failed", err)
	}
	assertValue(t, v.typ, LEPT_NUMBER)
	assertValue(t, v.NUMBER(), want)
}

func testString(t *testing.T, want string, str string) {
	v, err := Parse(str)
	if err != nil {
		t.Error("test parse string failed", err)
	}
	assertValue(t, v.typ, LEPT_STRING)
	assertValue(t, v.STRING(), want)
}

func TestArray(t *testing.T) {
	v, err := Parse("[]")
	if err != nil {
		t.Error("test parse array failed", err)
	}
	assertValue(t, v.typ, LEPT_ARRAY)
	assertValue(t, len(v.ARRAY()), 0)

	v, err = Parse("[ null , false , true , 123 , \"abc\" ]")
	if err != nil {
		t.Error("test parse array failed", err)
	}
	assertValue(t, v.typ, LEPT_ARRAY)
	assertValue(t, len(v.ARRAY()), 5)
	assertValue(t, v.ARRAY().Index(0).typ, LEPT_NULL)
	assertValue(t, v.ARRAY().Index(1).typ, LEPT_FALSE)
	assertValue(t, v.ARRAY().Index(2).typ, LEPT_TRUE)
	assertValue(t, v.ARRAY().Index(3).typ, LEPT_NUMBER)
	assertValue(t, v.ARRAY().Index(4).typ, LEPT_STRING)
	assertValue(t, v.ARRAY().Index(3).NUMBER(), 123.0)
	assertValue(t, v.ARRAY().Index(4).STRING(), "abc")

	v, err = Parse("[ [ ] , [ 0 ] , [ 0 , 1 ] , [ 0 , 1 , 2 ] ]")
	if err != nil {
		t.Error("test parse array failed", err)
	}
	assertValue(t, v.typ, LEPT_ARRAY)
	assertValue(t, len(v.ARRAY()), 4)
	for i := range 4 {
		a := v.ARRAY().Index(i)
		assertValue(t, a.typ, LEPT_ARRAY)
		assertValue(t, len(a.ARRAY()), i)
		for j := range i {
			e := a.ARRAY().Index(j)
			assertValue(t, e.typ, LEPT_NUMBER)
			assertValue(t, e.NUMBER(), float64(j))
		}
	}
}

func TestObject(t *testing.T) {
	v, err := Parse(" { } ")
	if err != nil {
		t.Error("test parse object failed", err)
	}
	assertValue(t, v.typ, LEPT_OBJECT)
	assertValue(t, len(v.OBJECT()), 0)

	v, err = Parse(`
    {
        "n": null ,
        "f": false ,
        "t": true ,
        "i": 123 ,
        "s": "abc" ,
        "a": [ 1, 2, 3 ] ,
        "o": { "1": 1, "2": 2, "3": 3 }
    } `)
	if err != nil {
		t.Error("test parse object failed", err)
	}
	assertValue(t, v.typ, LEPT_OBJECT)
	assertValue(t, len(v.OBJECT()), 7)
	assertValue(t, v.OBJECT().Index(0).K, "n")
	assertValue(t, v.OBJECT().Index(0).V.typ, LEPT_NULL)
	assertValue(t, v.OBJECT().Index(1).K, "f")
	assertValue(t, v.OBJECT().Index(1).V.typ, LEPT_FALSE)
	assertValue(t, v.OBJECT().Index(2).K, "t")
	assertValue(t, v.OBJECT().Index(2).V.typ, LEPT_TRUE)
	assertValue(t, v.OBJECT().Index(3).K, "i")
	assertValue(t, v.OBJECT().Index(3).V.typ, LEPT_NUMBER)
	assertValue(t, v.OBJECT().Index(3).V.NUMBER(), 123.0)
	assertValue(t, v.OBJECT().Index(4).K, "s")
	assertValue(t, v.OBJECT().Index(4).V.typ, LEPT_STRING)
	assertValue(t, v.OBJECT().Index(4).V.STRING(), "abc")
	assertValue(t, v.OBJECT().Index(5).K, "a")
	assertValue(t, v.OBJECT().Index(5).V.typ, LEPT_ARRAY)

	for i := range 3 {
		e := v.OBJECT().Index(5).V.ARRAY().Index(i)
		assertValue(t, e.typ, LEPT_NUMBER)
		assertValue(t, e.NUMBER(), float64(i)+1)
	}

	assertValue(t, v.OBJECT().Index(6).K, "o")
	o := v.OBJECT().Index(6).V
	assertValue(t, o.typ, LEPT_OBJECT)
	for i := range 3 {
		ov := o.Get(strconv.Itoa(i + 1))
		assertValue(t, o.OBJECT()[i].K, strconv.Itoa(i+1))
		assertValue(t, ov.typ, LEPT_NUMBER)
		assertValue(t, ov.NUMBER(), float64(i)+1)
	}
}

func assertValue[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestParse(t *testing.T) {
	data := `
	{
	    "title": "Design Patterns",
	    "subtitle": "Elements of Reusable Object-Oriented Software",
	    "author": [
	        "Erich Gamma",
	        "Richard Helm",
	        "Ralph Johnson",
	        "John Vlissides"
	    ],
	    "year": 2009,
	    "weight": 1.8,
	    "hardcover": true,
	    "publisher": {
	        "Company": "Pearson Education",
	        "Country": "India"
	    },
	    "website": null
	}
	    `

	v, err := Parse(data)
	if err != nil {
		t.Error("parse json data failed", err)
	}

	expected := &leptValue{
		typ: LEPT_OBJECT,
		u: leptObject{
			leptMember{K: "title", V: &leptValue{typ: LEPT_STRING, u: "Design Patterns"}},
			leptMember{K: "subtitle", V: &leptValue{typ: LEPT_STRING, u: "Elements of Reusable Object-Oriented Software"}},
			leptMember{K: "author", V: &leptValue{typ: LEPT_ARRAY, u: leptArray{
				&leptValue{typ: LEPT_STRING, u: "Erich Gamma"},
				&leptValue{typ: LEPT_STRING, u: "Richard Helm"},
				&leptValue{typ: LEPT_STRING, u: "Ralph Johnson"},
				&leptValue{typ: LEPT_STRING, u: "John Vlissides"},
			}}},
			leptMember{K: "year", V: &leptValue{typ: LEPT_NUMBER, u: 2009}},
			leptMember{K: "weight", V: &leptValue{typ: LEPT_NUMBER, u: 1.8}},
			leptMember{K: "hardcover", V: &leptValue{typ: LEPT_TRUE, u: true}},
			leptMember{K: "publisher", V: &leptValue{typ: LEPT_OBJECT, u: leptObject{
				leptMember{K: "Company", V: &leptValue{typ: LEPT_STRING, u: "Pearson Education"}},
				leptMember{K: "Country", V: &leptValue{typ: LEPT_STRING, u: "India"}},
			}}},
			leptMember{K: "website", V: &leptValue{typ: LEPT_NULL, u: &leptValue{"null", LEPT_NULL}}},
		}}

	if v.String() != expected.String() {
		t.Errorf("got %v want %v", v.String(), expected.String())
	}

	builder := NewObject().Set("title", NewString("Design Patterns")).
		Set("subtitle", NewString("Elements of Reusable Object-Oriented Software")).
		Set("author", NewArray().
			Append([]*leptValue{
				NewString("Erich Gamma"),
				NewString("Richard Helm"),
				NewString("Ralph Johnson")}...).
			Append(NewString("John Vlissides"))).
		Set("year", NewNumber(2009)).
		Set("weight", NewNumber(1.8)).
		Set("hardcover", NewBool(true)).
		Set("publisher", NewObject().
			Set("Company", NewString("Pearson Education")).
			Set("Country", NewString("India"))).
		Set("website", NewNull())

	if builder.String() != expected.String() {
		t.Errorf("got %v want %v", builder.String(), expected.String())
	}
}

func ExampleParse() {
	data := `
	{
	    "title": "Design Patterns",
	    "subtitle": "Elements of Reusable Object-Oriented Software",
	    "author": [
	        "Erich Gamma",
	        "Richard Helm",
	        "Ralph Johnson",
	        "John Vlissides"
	    ],
	    "year": 2009,
	    "weight": 1.8,
	    "hardcover": true,
	    "publisher": {
	        "Company": "Pearson Education",
	        "Country": "India"
	    },
	    "website": null
	}
	    `

	v, _ := Parse(data)
	fmt.Println(v)
	// Output: {"title": "Design Patterns", "subtitle": "Elements of Reusable Object-Oriented Software", "author": ["Erich Gamma", "Richard Helm", "Ralph Johnson", "John Vlissides"], "year": 2009, "weight": 1.8, "hardcover": true, "publisher": {"Company": "Pearson Education", "Country": "India"}, "website": null}
}

func BenchmarkParse(b *testing.B) {
	data := `
	{
	    "title": "Design Patterns",
	    "subtitle": "Elements of Reusable Object-Oriented Software",
	    "author": [
	        "Erich Gamma",
	        "Richard Helm",
	        "Ralph Johnson",
	        "John Vlissides"
	    ],
	    "year": 2009,
	    "weight": 1.8,
	    "hardcover": true,
	    "publisher": {
	        "Company": "Pearson Education",
	        "Country": "India"
	    },
	    "website": null
	}
	    `

	for b.Loop() {
		Parse(data)
	}
}
