package f_test

import (
	"github.com/angenalZZZ/gofunc/f"
	"net/http"
	"testing"
)

func TestHtml_Decode(t *testing.T) {
	type example struct {
		Title string   `html:"h1"`
		Files []string `html:"table.files tbody tr.js-navigation-item td.content,text"`
	}

	res, err := http.Get("https://github.com/andrewstuart/goq")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()

	var ex example

	err = f.NewHtmlDecoder(res.Body).Decode(&ex)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ex.Files[:6])
}

func TestJson_Get(t *testing.T) {
	json := f.Json(` {
	   "name": {"first": "Tom", "last": "Anderson"},
	   "age":37,
	   "children": ["Sara","Alex","Jack"],
	   "friends": [
	     {"first": "James", "last": "Murphy"},
	     {"first": "Roger", "last": "Craig"}
	   ]
	 }`)

	t.Log(json.Get("name.last"))
	t.Log(json.Get("age"))
	t.Log(json.Get("children"))
	t.Log(json.Get("children.#"))
	t.Log(json.Get("children.1"))
	t.Log(json.Get("child*.2"))
	t.Log(json.Get("c?ildren.0"))
	t.Log(json.Get("friends.#.first"))
	t.Log(json.GetMany("friends")[0])

	json = json.Sets("name.last", "Chinese")
	t.Log(json.Get("name.last"))
	json = json.Sets("age", 20)
	t.Log(json.Get("age"))
	json = json.Deletes("friends.1")
	t.Log(json.GetMany("friends")[0])
}

func TestJson_Jwt(t *testing.T) {
	o := struct {
		Token string
		Exp   int
	}{"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJleHAiOjE1OTUyMTA2MjUsImlzcyI6bnVsbCwic3ViIjpudWxsLCJhdWQiOm51bGwsIm5iZiI6bnVsbCwiaWF0IjpudWxsLCJqdGkiOm51bGx9.EN_oGUhyzGlbRJkMr0YpAj-6Uoxqkq2FT1lJYFno1iU", 3600}
	if p, err := f.EncodeJson(o); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("%s\n", p)
	}
	if m, ok := f.IsJwtToken(o.Token); ok {
		t.Logf("%s\n", f.EncodedMap(m))
	} else {
		t.Fatal("Error", 401)
	}
}
