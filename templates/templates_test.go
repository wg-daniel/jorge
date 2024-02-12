package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTemplate(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
<p>Hello World!</p>
`

	file := newFile("test*.html", input)
	defer os.Remove(file.Name())

	templ, err := Parse(file.Name())
	assertEqual(t, err, nil)

	assertEqual(t, templ.Type, PAGE)
	assertEqual(t, templ.Ext(), ".html")
	assertEqual(t, templ.Metadata["title"], "my new post")
	assertEqual(t, templ.Metadata["subtitle"], "a blog post")
	assertEqual(t, templ.Metadata["tags"].([]interface{})[0], "software")
	assertEqual(t, templ.Metadata["tags"].([]interface{})[1], "web")

	content, err := templ.Render(nil)
	assertEqual(t, err, nil)
	assertEqual(t, string(content), "<p>Hello World!</p>\n")
}

func TestNonTemplate(t *testing.T) {
	// not identified as front matter, leaving file as is
	input := `+++
title: my new post
subtitle: a blog post
+++
<p>Hello World!</p>`

	file := newFile("test*.html", input)
	defer os.Remove(file.Name())

	templ, err := Parse(file.Name())
	assertEqual(t, err, nil)
	assertEqual(t, templ.Type, STATIC)

	// not first thing in file, leaving as is
	input = `#+OPTIONS: toc:nil num:nil
---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
<p>Hello World!</p>`

	file = newFile("test*.html", input)
	defer os.Remove(file.Name())

	templ, err = Parse(file.Name())
	assertEqual(t, err, nil)
	assertEqual(t, templ.Type, STATIC)
}

func TestInvalidFrontMatter(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
`
	file := newFile("test*.html", input)
	defer os.Remove(file.Name())
	_, err := Parse(file.Name())

	assertEqual(t, err.Error(), "front matter not closed")

	input = `---
title
tags: ["software", "web"]
---
<p>Hello World!</p>`

	file = newFile("test*.html", input)
	defer os.Remove(file.Name())
	_, err = Parse(file.Name())
	assert(t, strings.Contains(err.Error(), "invalid yaml"))
}

func TestRenderLiquid(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
<h1>{{ page.title }}</h1>
<h2>{{ page.subtitle }}</h2>
<ul>{% for tag in page.tags %}
<li>{{tag}}</li>{% endfor %}
</ul>
`

	file := newFile("test*.html", input)
	defer os.Remove(file.Name())

	templ, err := Parse(file.Name())
	assertEqual(t, err, nil)
	ctx := map[string]interface{}{"page": templ.Metadata}
	content, err := templ.Render(ctx)
	assertEqual(t, err, nil)
	expected := `<h1>my new post</h1>
<h2>a blog post</h2>
<ul>
<li>software</li>
<li>web</li>
</ul>
`
	assertEqual(t, string(content), expected)
}

func TestRenderOrg(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
#+OPTIONS: toc:nil num:nil
* My title
** my Subtitle
- list 1
- list 2
`

	file := newFile("test*.org", input)
	defer os.Remove(file.Name())

	templ, err := Parse(file.Name())
	assertEqual(t, err, nil)
	assertEqual(t, templ.Ext(), ".html")

	content, err := templ.Render(nil)
	assertEqual(t, err, nil)
	expected := `<div id="outline-container-headline-1" class="outline-2">
<h2 id="headline-1">
My title
</h2>
<div id="outline-text-headline-1" class="outline-text-2">
<div id="outline-container-headline-2" class="outline-3">
<h3 id="headline-2">
my Subtitle
</h3>
<div id="outline-text-headline-2" class="outline-text-3">
<ul>
<li>list 1</li>
<li>list 2</li>
</ul>
</div>
</div>
</div>
</div>
`
	assertEqual(t, string(content), expected)
}

func TestRenderLiquidLayout(t *testing.T) {
	input := `---
title: base layout
---
<!DOCTYPE html>
<html>
<body>
<p>this is the {{layout.title}} that wraps the content of {{ page.title}}</p>
{{ content }}
</body>
</html>
`

	base := newFile("layouts/base*.html", input)
	defer os.Remove(base.Name())
	baseTempl, err := Parse(base.Name())
	assertEqual(t, err, nil)
	assertEqual(t, baseTempl.Type, LAYOUT)

	context := map[string]interface{}{
		"layouts": map[string]Template{
			"base": *baseTempl,
		},
	}

	input = `---
title: my very first post
layout: base
date: 2023-12-01
---
<h1>{{page.title}}</h1>`

	post := newFile("src/post1*.html", input)
	defer os.Remove(post.Name())

	templ, err := Parse(post.Name())
	assertEqual(t, err, nil)
	assertEqual(t, templ.Type, POST)
	content, err := templ.Render(context)
	assertEqual(t, err, nil)
	expected := `<!DOCTYPE html>
<html>
<body>
<p>this is the base layout that wraps the content of my very first post</p>
<h1>my very first post</h1>

</body>
</html>
`
	assertEqual(t, string(content), expected)
}

func TestRenderOrgLayout(t *testing.T) {
	// TODO
}

func TestRenderLayoutLayout(t *testing.T) {
	// TODO
}

// ------ HELPERS --------

func newFile(path string, contents string) *os.File {
	parts := strings.Split(path, string(filepath.Separator))
	name := parts[len(parts)-1]
	path = filepath.Join(parts[:len(parts)-1]...)
	path = filepath.Join(os.TempDir(), path)
	os.MkdirAll(path, 0777)

	file, _ := os.CreateTemp(path, name)
	file.WriteString(contents)
	return file
}

// TODO move to assert package
func assert(t *testing.T, cond bool) {
	t.Helper()
	if !cond {
		t.Fatalf("%v is false", cond)
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}