package mhproc

import "testing"

func TestVerySimpleHtml(t *testing.T) {
	str := `title: Prossima gara Wien Rundumadum
datetime: 2024-11-08 19:00
id: 20241108-00
---
<h1>Hello</h1>
<p>Something to say!</p>`
	prc := MdHtmlProcess{debug: true, validateMandatory: true, templDir: "../templates/htmlgen"}
	if err := prc.ProcessToHtml(str); err != nil {
		t.Error("Process error", err)
		return
	}
}

func TestSimpleHtmlWithNoDataAndTemplate(t *testing.T) {
	str := `---
<h1>Hello</h1>
<p>Something to say!</p>`
	prc := MdHtmlProcess{debug: true, templDir: ""}
	if err := prc.ProcessToHtml(str); err != nil {
		t.Error("Process error", err)
		return
	}
}
