package trans

import (
	"bytes"
	"cmp"
	"fmt"
	"invido-site/src/idl"
	"path"
	"slices"
	"strings"
	"text/template"
)

type mdhtTagPostsNode struct {
	MdhtLineNode
	mapLinks *idl.MapPagePostsLinks
}

func NewTagPostsNode(preline string, maplinks *idl.MapPagePostsLinks) *mdhtTagPostsNode {
	res := mdhtTagPostsNode{
		mapLinks: maplinks,
	}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func (ln *mdhtTagPostsNode) AddParamString(parVal string) error {
	return fmt.Errorf("[mdhtTagPostsNode - AddParamString] no parameter")
}

func (ln *mdhtTagPostsNode) AddblockHtml(val string) error {
	if ln.after_link != "" {
		return fmt.Errorf("[AddblockHtml] already set")
	}
	ln.after_link = val
	return nil
}

func (ln *mdhtTagPostsNode) Transform(templDir string) error {
	if templDir == "" {
		return fmt.Errorf("[Transform - TagPosts] templ dir is not set")
	}
	if ln.mapLinks == nil {
		return fmt.Errorf("[Transform - TagPosts] map links are not set")
	}
	templName := path.Join(templDir, "transform.html")
	tmplPage := template.Must(template.New("Trf").ParseFiles(templName))
	var partMerged bytes.Buffer

	if len(ln.mapLinks.Tags) > 0 {
		partFirst, err := ln.transformTags(tmplPage)
		if err != nil {
			return err
		}
		partFirst.WriteTo(&partMerged)
		res := fmt.Sprintf("%s%s%s", ln.before_link, partMerged.String(), ln.after_link)
		ln.block = res
	} else {
		ln.block = ""
	}

	return nil
}

func (ln *mdhtTagPostsNode) transformTags(tmplPage *template.Template) (*bytes.Buffer, error) {
	CtxFirst := struct {
		Tags []idl.TagItem
	}{
		Tags: ln.mapLinks.Tags,
	}
	// Sort by DateTime (ascending)
	slices.SortFunc(CtxFirst.Tags, func(a, b idl.TagItem) int {
		return cmp.Compare(
			strings.ToLower(a.Title),
			strings.ToLower(b.Title),
		)
	})
	var partFirst bytes.Buffer
	if err := tmplPage.ExecuteTemplate(&partFirst, "tagposts", CtxFirst); err != nil {
		return nil, err
	}
	return &partFirst, nil
}

func (ln *mdhtTagPostsNode) HasJsonBlock() bool {
	return false
}

func (ln *mdhtTagPostsNode) JsonBlock() string {
	return ""
}

func (ln *mdhtTagPostsNode) JsonBlockType() string {
	return ""
}
