package trans

import (
	"bytes"
	"fmt"
	"invido-site/src/idl"
	"invido-site/src/util"
	"log"
	"path"
	"slices"
	"strings"
	"text/template"
)

type mdhtSingleTaggedPostsNode struct {
	MdhtLineNode
	mapLinks *idl.MapPagePostsLinks
	_title   string
}

func NewSingleTaggedPostsNode(preline string, maplinks *idl.MapPagePostsLinks) *mdhtSingleTaggedPostsNode {
	res := mdhtSingleTaggedPostsNode{
		mapLinks: maplinks,
	}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func (ln *mdhtSingleTaggedPostsNode) AddParamString(parVal string) error {
	if ln._title != "" {
		return fmt.Errorf("[AddParamString] parameter already set")
	}
	ln._title = parVal
	return nil
}

func (ln *mdhtSingleTaggedPostsNode) AddblockHtml(val string) error {
	if ln.after_link != "" {
		return fmt.Errorf("[AddblockHtml] already set")
	}
	ln.after_link = val
	return nil
}

func (ln *mdhtSingleTaggedPostsNode) Transform(templDir string) error {
	if templDir == "" {
		return fmt.Errorf("[Transform - SingleTaggedPosts] templ dir is not set")
	}
	if ln.mapLinks == nil {
		return fmt.Errorf("[Transform - SingleTaggedPosts] map links are not set")
	}
	if ln._title == "" {
		return fmt.Errorf("[Transform - SingleTaggedPosts] tag title is not set")
	}
	templName := path.Join(templDir, "transform.html")
	tmplPage := template.Must(template.New("Trf").ParseFiles(templName))
	tagPosts := []*PostWithData{}
	var partMerged bytes.Buffer
	for _, item := range ln.mapLinks.MapTag[ln._title] {
		pwd := PostWithData{
			DateFormatted: util.FormatDateIt(item.DateTime),
			DateTimeTxt:   item.DateTime.Format("2006-01-02 15:00"),
			DateTime:      item.DateTime,
			Title:         item.Title,
			Link:          item.Uri,
		}
		tagPosts = append(tagPosts, &pwd)
	}
	if len(tagPosts) > 0 {
		log.Println("[Transform - SingleTaggedPosts] working on tags", ln._title, len(tagPosts))
		partFirst, err := ln.transformPost(tmplPage, ln._title, tagPosts)
		if err != nil {
			return err
		}
		partFirst.WriteTo(&partMerged)
	}

	res := fmt.Sprintf("%s%s%s", ln.before_link, partMerged.String(), ln.after_link)
	ln.block = res
	return nil
}

func (ln *mdhtSingleTaggedPostsNode) transformPost(tmplPage *template.Template, current_title string, tagPosts []*PostWithData) (*bytes.Buffer, error) {
	CtxFirst := struct {
		Title      string
		NumOfPosts int
		TagPosts   []*PostWithData
	}{
		Title:      current_title,
		NumOfPosts: len(tagPosts),
		TagPosts:   tagPosts,
	}
	// Sort by DateTime (descending)
	slices.SortFunc(CtxFirst.TagPosts, func(a, b *PostWithData) int {
		if a.DateTime.Before(b.DateTime) {
			return 1
		}
		if a.DateTime.After(b.DateTime) {
			return -1
		}
		return 0
	})
	var partFirst bytes.Buffer
	if err := tmplPage.ExecuteTemplate(&partFirst, "singletaggedposts", CtxFirst); err != nil {
		return nil, err
	}
	return &partFirst, nil
}

func (ln *mdhtSingleTaggedPostsNode) HasJsonBlock() bool {
	return false
}

func (ln *mdhtSingleTaggedPostsNode) JsonBlock() string {
	return ""
}

func (ln *mdhtSingleTaggedPostsNode) JsonBlockType() string {
	return ""
}
