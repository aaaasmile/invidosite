package mhparser

import (
	"invido-site/src/idl"
	"strings"
	"testing"
	"time"
)

func createFakeLinks() *idl.MapPagePostsLinks {
	loc := time.Now().Location()
	mapLinks := &idl.MapPagePostsLinks{
		MapPost:  map[string]idl.PostLinks{},
		MapPage:  map[string]*idl.PageItem{},
		MapTag:   map[string][]*idl.PostItem{},
		ListPost: []idl.PostItem{},
		ListPage: []idl.PageItem{},
		Tags:     []idl.TagItem{},
	}
	mapLinks.ListPost = append(mapLinks.ListPost, idl.PostItem{Title: "A ti", Uri: "A uri", DateTime: time.Date(2010, 2, 15, 0, 0, 0, 0, loc)})
	mapLinks.ListPost = append(mapLinks.ListPost, idl.PostItem{Title: "B ti", Uri: "B uri", DateTime: time.Date(2010, 3, 31, 0, 0, 0, 0, loc)})
	mapLinks.ListPost = append(mapLinks.ListPost, idl.PostItem{Title: "C ti", Uri: "C uri", DateTime: time.Date(2011, 6, 10, 0, 0, 0, 0, loc)})
	mapLinks.Tags = append(mapLinks.Tags, idl.TagItem{Title: "Gedanken", NumOfPosts: 6}, idl.TagItem{Title: "Ultra", NumOfPosts: 7})
	lstPostTags := []*idl.PostItem{}
	lstPostTags = append(lstPostTags, &idl.PostItem{Title: "C ci", Uri: "C uri"})
	lstPostTags = append(lstPostTags, &idl.PostItem{Title: "D di", Uri: "D uri"})
	mapLinks.MapTag["MaratonaGara"] = lstPostTags
	return mapLinks
}

func TestParseData(t *testing.T) {
	str := `title: Prossima gara Wien Rundumadum
datetime: 2024-11-08 19:00
id: 20241108-00`

	lex := ScriptGrammar{
		Debug: true,
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	if lex.Id != "20241108-00" {
		t.Error("unexpected id", lex.Id)
	}
	if lex.Title != "Prossima gara Wien Rundumadum" {
		t.Error("unexpected Title", lex.Title)
	}
	if lex.Datetime.Year() != 2024 {
		t.Error("unexpected Year", lex.Datetime)
	}
	if lex.Datetime.Hour() != 19 {
		t.Error("unexpected Hour", lex.Datetime)
	}
}

func TestParseCustomData(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
frasefamosa : non dire gatto
`

	lex := ScriptGrammar{
		Debug: true,
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}

	if frfam, ok := lex.CustomData["frasefamosa"]; ok {
		if frfam != "non dire gatto" {
			t.Error("unexpected custom data", lex.CustomData)
		}
	} else {
		t.Error("custom data missed", lex.CustomData)
	}
}

func TestParseSimpleHtml(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>Pa</p>
il nuovo`

	lex := ScriptGrammar{
		Debug: true,
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	if len(ll.ArrayValue) != 2 {
		t.Errorf("expected two html lines, but %d", len(ll.ArrayValue))
		return
	}
}

func TestParseHtmlLinkBlock(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>Pa</p>
<p>Tracker: [link 'https://wien-rundumadum-2024-130k.legendstracking.com/']</p>`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	if len(ll.ArrayValue) != 2 {
		t.Errorf("expected two html lines, but %d", len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[1]
	if !strings.Contains(secline, "<p>Tracker: <a href=\"https://wien-rundumadum-2024-130k.legendstracking.com/\" target=\"_blank\">https://wien-rundumadum-2024-130k.legendstracking.com/</a></p>") {
		t.Errorf("expected  <a> in generated  html, but %s ", secline)
	}
}

func TestParseHtmlLinkBlockThreeLines(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>Pa</p>
<p>Tracker: [link 'https://wien-rundumadum-2024-130k.legendstracking.com/']</p>
<div>hello</div>`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	if len(ll.ArrayValue) != 3 {
		t.Errorf("expected 3 html lines, but %d", len(ll.ArrayValue))
		return
	}
	secline0 := ll.ArrayValue[0]
	if !strings.Contains(secline0, "<p>Pa</p>") {
		t.Errorf("expected <p>Pa</p> in generated  html, but %s ", secline0)
	}
	secline := ll.ArrayValue[1]
	if !strings.Contains(secline, "<p>Tracker: <a href=\"https://wien-rundumadum-2024-130k.legendstracking.com/\" target=\"_blank\">https://wien-rundumadum-2024-130k.legendstracking.com/</a></p>") {
		t.Errorf("expected  <a href> in generated  html, but %s ", secline)
	}
	secline = ll.ArrayValue[2]
	if !strings.Contains(secline, "<div>hello</div>") {
		t.Errorf("expected  <div>hello</div> in generated  html, but %s ", secline)
	}
}

func TestParseHtmlLinkBlockOneLine(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
[link 'https://wien-rundumadum-2024-130k.legendstracking.com/']`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	if len(ll.ArrayValue) != 1 {
		t.Errorf("expected one html lines, but %d", len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[0]
	if !strings.Contains(secline, "<a href=\"https://wien-rundumadum-2024-130k.legendstracking.com/\" target=\"_blank\">https://wien-rundumadum-2024-130k.legendstracking.com/</a>") {
		t.Errorf("expected  <a> in generated  html, but %s ", secline)
	}
}

func TestParseHtmlLinkBlockTwoLines(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
[link 'https://wien-rundumadum-2024-130k.legendstracking.com/']<p>
hello</p>`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	if len(ll.ArrayValue) != 2 {
		t.Errorf("expected 2 html lines, but %d", len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[0]
	if !strings.Contains(secline, "<a href=\"https://wien-rundumadum-2024-130k.legendstracking.com/\" target=\"_blank\">https://wien-rundumadum-2024-130k.legendstracking.com/</a><p>") {
		t.Errorf("expected  <a> in generated  html, but %s ", secline)
	}
	secline = ll.ArrayValue[1]
	if !strings.Contains(secline, "hello</p>") {
		t.Errorf("expected  hello</p> in generated  html, but %s ", secline)
	}
}

func TestParseHtmlFigStack(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>Ciao</p>
[figstack
  'AustriaBackyardUltra2024011.jpg', 'Partenza mondiale Backyard',
  'backyard_award.png', 'Certificato finale'
]
<p>hello</p>`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	if len(ll.ArrayValue) != 3 {
		t.Errorf("expected 3 html lines, but %d", len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[0]
	if !strings.Contains(secline, "<p>Ciao</p>") {
		t.Errorf("expected <p>Ciao</p> in generated  html, but %s ", secline)
		return
	}
	secline = ll.ArrayValue[1]
	if !strings.Contains(secline, `id="AustriaBackyardUltra2024011_00" onclick="appGallery.displayImage`) {
		t.Errorf("expected AustriaBackyardUltra2024011 in generated  html, but %s ", secline)
		return
	}
	if !strings.Contains(secline, `id="backyard_award_02" onclick="appGallery.displayImage`) {
		t.Errorf("expected backyard_award.png in generated  html, but %s ", secline)
		return
	}

	secline = ll.ArrayValue[2]
	if !strings.Contains(secline, "<p>hello</p>") {
		t.Errorf("expected  <p>hello</p> in generated  html, but %s ", secline)
	}
}

func TestParseZeroTags(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
tags:
`
	lex := ScriptGrammar{
		Debug: true,
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}

	if len(lex.Tags) != 0 {
		t.Error("expected zero tags")
	}
}

func TestParseTwoTags(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
tags: ultra,adamello
`
	lex := ScriptGrammar{
		Debug: true,
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}

	if len(lex.Tags) != 2 {
		t.Error("expected two tags")
		return
	}
	if strings.Compare(lex.Tags[0], "ultra") != 0 {
		t.Error("expected first tag ultra, but ", lex.Tags[0])
		return
	}
	if strings.Compare(lex.Tags[1], "adamello") != 0 {
		t.Error("expected second tag adamello, but ", lex.Tags[0])
		return
	}
}

func TestSimpleHtmlZeroTags(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
tags:
---
<header class="withimg">
  <div>
    <h1>Quo Vadis</h1>
    <time>4 Gennaio 2023</time>
  </div>
  <img src="bestage.jpg" /> 
</header>
`
	lex := ScriptGrammar{
		Debug: true,
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}

	if len(lex.Tags) != 0 {
		t.Error("expected zero tags")
	}
}

func TestParseHtmlLinkWithCaptionBlock(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>Pa</p>
<p>Tracker: [linkcap 'Tracker', 'https://wien-rundumadum-2024-130k.legendstracking.com/']</p>`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	if len(ll.ArrayValue) != 2 {
		t.Errorf("expected two html lines, but %d", len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[1]
	if !strings.Contains(secline, "<p>Tracker: <a href=\"https://wien-rundumadum-2024-130k.legendstracking.com/\" target=\"_blank\">Tracker</a></p>") {
		t.Errorf("expected  <a> in generated  html, but %s ", secline)
	}
}

func TestVideoYoutube(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>Pa</p>
<p>Video: [youtube 'IOP7RhDnLnw']</p>`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	if len(ll.ArrayValue) != 2 {
		t.Errorf("expected two html lines, but %d", len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[1]
	if !strings.Contains(secline, "<p>Video: <iframe allowfullscreen=\"allowfullscreen\" frameborder=\"0\" height=\"266\" mozallowfullscreen=\"mozallowfullscreen\" src=\"https://www.youtube.com/embed/IOP7RhDnLnw") {
		t.Errorf("expected  <iframe in generated  html, but %s ", secline)
	}
}

func TestParseHtmlLastPost(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>first line</p>
s<p>[latest_posts 'Invido SIte', '7']</p>e`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
		MapLinks: createFakeLinks(),
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	len_exp := 2
	if len(ll.ArrayValue) != len_exp {
		t.Errorf("expected %d html lines, but have %d lines", len_exp, len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[1]
	if !strings.Contains(secline, "Ultimi post") {
		t.Errorf("expected  Ultimi post, but %s ", secline)
	}
}

func TestParseHtmlArchivePost(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>first line</p>
s<p>[archive_posts]</p>e`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
		MapLinks: createFakeLinks(),
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	len_exp := 2
	if len(ll.ArrayValue) != len_exp {
		t.Errorf("expected %d html lines, but have %d lines", len_exp, len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[1]
	if !strings.Contains(secline, "Anno 2010") {
		t.Errorf("expected  Archivio anno 2010, but %s ", secline)
	}
	if !strings.Contains(secline, "Anno 2011") {
		t.Errorf("expected  Archivio anno 2011, but %s ", secline)
	}
}

func TestParseHtmlTagPost(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>first line</p>
s<p>[tag_posts]</p>e`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
		MapLinks: createFakeLinks(),
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	len_exp := 2
	if len(ll.ArrayValue) != len_exp {
		t.Errorf("expected %d html lines, but have %d lines", len_exp, len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[1]
	if !strings.Contains(secline, "Gedanken") {
		t.Errorf("expected  Gedanken, but %s ", secline)
	}
	if !strings.Contains(secline, "Ultra") {
		t.Errorf("expected  Ultra, but %s ", secline)
	}
}

func TestParseHtmlTaggedPosts(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>first line</p>
s<p>[single_taggedposts 'MaratonaGara']</p>e`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
		MapLinks: createFakeLinks(),
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	len_exp := 2
	if len(ll.ArrayValue) != len_exp {
		t.Errorf("expected %d html lines, but have %d lines", len_exp, len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[1]
	if !strings.Contains(secline, "C ci") {
		t.Errorf("expected  C ci, but %s ", secline)
	}
	if !strings.Contains(secline, "D di") {
		t.Errorf("expected  D di, but %s ", secline)
	}
}

func TestParseHtmlImgLinkRun(t *testing.T) {
	str := `title: Un altro post entusiasmante
datetime: 2024-12-23
id: 20241108-00
---
<p>first line</p>
s<p>[img_link_run 'foto01_320.jpg', 'https://cup.invido.it/#/', 'RUN nel Browser']</p>e`

	lex := ScriptGrammar{
		Debug:    true,
		TemplDir: "../templates/htmlgen",
		MapLinks: createFakeLinks(),
	}
	err := lex.ParseScript(str)
	if err != nil {
		t.Error("Error is: ", err)
		return
	}

	err = lex.CheckNorm()
	if err != nil {
		t.Error("Error in parser norm ", err)
		return
	}
	err = lex.EvaluateParams()
	if err != nil {
		t.Error("Error in evaluate ", err)
		return
	}
	nrm := lex.Norm["main"]
	lastFns := len(nrm.FnsList) - 1
	stFns := nrm.FnsList[lastFns]
	if len(stFns.Params) != 1 && !stFns.Params[0].IsArray {
		t.Error("expected one array param with lines")
		return
	}
	ll := &stFns.Params[0]
	len_exp := 2
	if len(ll.ArrayValue) != len_exp {
		t.Errorf("expected %d html lines, but have %d lines", len_exp, len(ll.ArrayValue))
		return
	}
	secline := ll.ArrayValue[1]
	if !strings.Contains(secline, "button type=") {
		t.Errorf("expected  button type=, but %s ", secline)
	}
}
