package lept_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/wasuppu/lept"
)

func TestBaseType(t *testing.T) {
	t.Run("null", func(t *testing.T) {
		v, err := lept.Parse("null")
		if err != nil {
			t.Fatal("test parse null failed", err)
		}

		assertValue(t, v.Type, lept.TypeNull)
	})

	t.Run("true", func(t *testing.T) {
		v, err := lept.Parse("true")
		if err != nil {
			t.Fatal("test parse true failed", err)
		}

		assertValue(t, v.Type, lept.TypeTrue)
	})

	t.Run("false", func(t *testing.T) {
		v, err := lept.Parse("false")
		if err != nil {
			t.Error("test parse false failed", err)
		}

		assertValue(t, v.Type, lept.TypeFalse)
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
	v, err := lept.Parse(number)
	if err != nil {
		t.Fatal("test parse number failed", err)
	}
	assertValue(t, v.Type, lept.TypeNumber)
	assertValue(t, v.NUMBER(), want)
}

func testString(t *testing.T, want string, str string) {
	v, err := lept.Parse(str)
	if err != nil {
		t.Error("test parse string failed", err)
	}
	assertValue(t, v.Type, lept.TypeString)
	assertValue(t, v.STRING(), want)
}

func TestArray(t *testing.T) {
	v, err := lept.Parse("[]")
	if err != nil {
		t.Error("test parse array failed", err)
	}
	assertValue(t, v.Type, lept.TypeArray)
	assertValue(t, len(v.ARRAY()), 0)

	v, err = lept.Parse("[ null , false , true , 123 , \"abc\" ]")
	if err != nil {
		t.Error("test parse array failed", err)
	}
	assertValue(t, v.Type, lept.TypeArray)
	assertValue(t, len(v.ARRAY()), 5)
	assertValue(t, v.ARRAY().Index(0).Type, lept.TypeNull)
	assertValue(t, v.ARRAY().Index(1).Type, lept.TypeFalse)
	assertValue(t, v.ARRAY().Index(2).Type, lept.TypeTrue)
	assertValue(t, v.ARRAY().Index(3).Type, lept.TypeNumber)
	assertValue(t, v.ARRAY().Index(4).Type, lept.TypeString)
	assertValue(t, v.ARRAY().Index(3).NUMBER(), 123.0)
	assertValue(t, v.ARRAY().Index(4).STRING(), "abc")

	v, err = lept.Parse("[ [ ] , [ 0 ] , [ 0 , 1 ] , [ 0 , 1 , 2 ] ]")
	if err != nil {
		t.Error("test parse array failed", err)
	}
	assertValue(t, v.Type, lept.TypeArray)
	assertValue(t, len(v.ARRAY()), 4)
	for i := range 4 {
		a := v.ARRAY().Index(i)
		assertValue(t, a.Type, lept.TypeArray)
		assertValue(t, len(a.ARRAY()), i)
		for j := range i {
			e := a.ARRAY().Index(j)
			assertValue(t, e.Type, lept.TypeNumber)
			assertValue(t, e.NUMBER(), float64(j))
		}
	}
}

func TestObject(t *testing.T) {
	v, err := lept.Parse(" { } ")
	if err != nil {
		t.Error("test parse object failed", err)
	}
	assertValue(t, v.Type, lept.TypeObject)
	assertValue(t, len(v.OBJECT()), 0)

	v, err = lept.Parse(`
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
	assertValue(t, v.Type, lept.TypeObject)
	assertValue(t, len(v.OBJECT()), 7)
	assertValue(t, v.OBJECT().Index(0).K, "n")
	assertValue(t, v.OBJECT().Index(0).V.Type, lept.TypeNull)
	assertValue(t, v.OBJECT().Index(1).K, "f")
	assertValue(t, v.OBJECT().Index(1).V.Type, lept.TypeFalse)
	assertValue(t, v.OBJECT().Index(2).K, "t")
	assertValue(t, v.OBJECT().Index(2).V.Type, lept.TypeTrue)
	assertValue(t, v.OBJECT().Index(3).K, "i")
	assertValue(t, v.OBJECT().Index(3).V.Type, lept.TypeNumber)
	assertValue(t, v.OBJECT().Index(3).V.NUMBER(), 123.0)
	assertValue(t, v.OBJECT().Index(4).K, "s")
	assertValue(t, v.OBJECT().Index(4).V.Type, lept.TypeString)
	assertValue(t, v.OBJECT().Index(4).V.STRING(), "abc")
	assertValue(t, v.OBJECT().Index(5).K, "a")
	assertValue(t, v.OBJECT().Index(5).V.Type, lept.TypeArray)

	for i := range 3 {
		e := v.OBJECT().Index(5).V.ARRAY().Index(i)
		assertValue(t, e.Type, lept.TypeNumber)
		assertValue(t, e.NUMBER(), float64(i)+1)
	}

	assertValue(t, v.OBJECT().Index(6).K, "o")
	o := v.OBJECT().Index(6).V
	assertValue(t, o.Type, lept.TypeObject)
	for i := range 3 {
		ov := o.Get(strconv.Itoa(i + 1))
		assertValue(t, o.OBJECT()[i].K, strconv.Itoa(i+1))
		assertValue(t, ov.Type, lept.TypeNumber)
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

	v, err := lept.Parse(data)
	if err != nil {
		t.Error("parse json data failed", err)
	}

	expected := &lept.Value{
		Type: lept.TypeObject,
		U: lept.Object{
			lept.Member{K: "title", V: &lept.Value{Type: lept.TypeString, U: "Design Patterns"}},
			lept.Member{K: "subtitle", V: &lept.Value{Type: lept.TypeString, U: "Elements of Reusable Object-Oriented Software"}},
			lept.Member{K: "author", V: &lept.Value{Type: lept.TypeArray, U: lept.Array{
				&lept.Value{Type: lept.TypeString, U: "Erich Gamma"},
				&lept.Value{Type: lept.TypeString, U: "Richard Helm"},
				&lept.Value{Type: lept.TypeString, U: "Ralph Johnson"},
				&lept.Value{Type: lept.TypeString, U: "John Vlissides"},
			}}},
			lept.Member{K: "year", V: &lept.Value{Type: lept.TypeNumber, U: 2009}},
			lept.Member{K: "weight", V: &lept.Value{Type: lept.TypeNumber, U: 1.8}},
			lept.Member{K: "hardcover", V: &lept.Value{Type: lept.TypeTrue, U: true}},
			lept.Member{K: "publisher", V: &lept.Value{Type: lept.TypeObject, U: lept.Object{
				lept.Member{K: "Company", V: &lept.Value{Type: lept.TypeString, U: "Pearson Education"}},
				lept.Member{K: "Country", V: &lept.Value{Type: lept.TypeString, U: "India"}},
			}}},
			lept.Member{K: "website", V: &lept.Value{Type: lept.TypeNull, U: &lept.Value{"null", lept.TypeNull}}},
		}}

	if v.String() != expected.String() {
		t.Errorf("got %v want %v", v.String(), expected.String())
	}

	builder := lept.NewObject().Set("title", lept.NewString("Design Patterns")).
		Set("subtitle", lept.NewString("Elements of Reusable Object-Oriented Software")).
		Set("author", lept.NewArray().
			Append([]*lept.Value{
				lept.NewString("Erich Gamma"),
				lept.NewString("Richard Helm"),
				lept.NewString("Ralph Johnson")}...).
			Append(lept.NewString("John Vlissides"))).
		Set("year", lept.NewNumber(2009)).
		Set("weight", lept.NewNumber(1.8)).
		Set("hardcover", lept.NewBool(true)).
		Set("publisher", lept.NewObject().
			Set("Company", lept.NewString("Pearson Education")).
			Set("Country", lept.NewString("India"))).
		Set("website", lept.NewNull())

	if builder.String() != expected.String() {
		t.Errorf("got %v want %v", builder.String(), expected.String())
	}
}

func TestUnmarshal(t *testing.T) {
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

	type Book struct {
		Title     string    `json:"title"`
		Subtitle  string    `json:"subtitle"`
		Author    [4]string `json:"author"`
		Year      int       `json:"year"`
		Weight    float64   `json:"weight"`
		Hardcover bool      `json:"hardcover"`
		Publisher struct {
			Company string `json:"Company"`
			Country string `json:"Country"`
		} `json:"publisher"`
		Website string `json:"website"`
	}

	want := Book{
		Title:    "Design Patterns",
		Subtitle: "Elements of Reusable Object-Oriented Software",
		Author: [4]string{
			"Erich Gamma",
			"Richard Helm",
			"Ralph Johnson",
			"John Vlissides",
		},
		Year:      2009,
		Weight:    1.8,
		Hardcover: true,
		Publisher: struct {
			Company string `json:"Company"`
			Country string `json:"Country"`
		}{
			Company: "Pearson Education",
			Country: "India",
		},
		Website: "",
	}

	v, err := lept.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	book := Book{}
	if err := lept.Unmarshal(v, &book); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(book, want) {
		t.Errorf("got %v want %v", book, want)
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

	v, _ := lept.Parse(data)
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
		lept.Parse(data)
	}
}
