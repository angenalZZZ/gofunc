package f

import "testing"

func TestJson_Get(t *testing.T) {
	json := Json(` {
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
