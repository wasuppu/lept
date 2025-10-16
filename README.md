# lept

A tiny json parser

## Overview

Code structure adapted from the [json-tutorial](https://github.com/miloyip/json-tutorial), while its parsing logic comes from the book [interpreterbook](https://interpreterbook.com/).

Additionally, some utility functions have been added for user convenience.

Current version has partial string support, unicode is omitted in test cases.

## Usage

```go
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

// Directly get the value
fmt.Println(v.Get("weight").NUMBER())
fmt.Println(v.Get("author").ARRAY()[1].STRING())
fmt.Println(v.Get("author").ARRAY().Index(2).STRING())
fmt.Println(v.Get("publisher").Get("Company").STRING())
fmt.Println(v.Seek("publisher", "Company")) // Access value from nested object structure

// Decode values from struct using Unmarshal
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

book := Book{}
lept.Unmarshal(v, &book)
fmt.Print(book.Publisher)
```
