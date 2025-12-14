package trans

import (
	"bytes"
	"fmt"
	"invido-site/src/idl"
	"invido-site/src/util"
	"path"
	"strings"
	"text/template"
)

type mdhtArchivePostsNode struct {
	MdhtLineNode
	mapLinks *idl.MapPagePostsLinks
}

func NewArchivePostsNode(preline string, maplinks *idl.MapPagePostsLinks) *mdhtArchivePostsNode {
	res := mdhtArchivePostsNode{
		mapLinks: maplinks,
	}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func (ln *mdhtArchivePostsNode) AddParamString(parVal string) error {
	return fmt.Errorf("[mdhtArchivePostsNode - AddParamString] no parameter")
}

func (ln *mdhtArchivePostsNode) AddblockHtml(val string) error {
	if ln.after_link != "" {
		return fmt.Errorf("[AddblockHtml] already set")
	}
	ln.after_link = val
	return nil
}

func (ln *mdhtArchivePostsNode) Transform(templDir string) error {
	if templDir == "" {
		return fmt.Errorf("[Transform - ArchivePosts] templ dir is not set")
	}
	if ln.mapLinks == nil {
		return fmt.Errorf("[Transform - ArchivePosts] map links are not set")
	}
	templName := path.Join(templDir, "transform.html")
	tmplPage := template.Must(template.New("Trf").ParseFiles(templName))
	yearPosts := []*PostWithData{}
	var partMerged bytes.Buffer
	current_year := 0
	for _, item := range ln.mapLinks.ListPost {
		pwd := PostWithData{
			DateFormatted: util.FormatDateIt(item.DateTime),
			DateTimeTxt:   item.DateTime.Format("2006-01-02 15:00"),
			DateTime:      item.DateTime,
			Title:         item.Title,
			Link:          item.Uri,
		}
		post_year := item.DateTime.Year()
		if current_year == 0 {
			current_year = post_year
			yearPosts = append(yearPosts, &pwd)
			continue
		}
		if current_year != post_year {
			partFirst, err := ln.transformPost(tmplPage, current_year, yearPosts)
			if err != nil {
				return err
			}
			partFirst.WriteTo(&partMerged)

			current_year = post_year
			yearPosts = []*PostWithData{}
		}
		yearPosts = append(yearPosts, &pwd)
	}
	if len(yearPosts) > 0 {
		partFirst, err := ln.transformPost(tmplPage, current_year, yearPosts)
		if err != nil {
			return err
		}
		partFirst.WriteTo(&partMerged)
	}

	res := fmt.Sprintf("%s%s%s", ln.before_link, partMerged.String(), ln.after_link)
	ln.block = res
	return nil
}

func (ln *mdhtArchivePostsNode) transformPost(tmplPage *template.Template, current_year int, yearPosts []*PostWithData) (*bytes.Buffer, error) {
	CtxFirst := struct {
		Year       int
		NumOfPosts int
		YearPosts  []*PostWithData
	}{
		Year:       current_year,
		NumOfPosts: len(yearPosts),
		YearPosts:  yearPosts,
	}
	var partFirst bytes.Buffer
	if err := tmplPage.ExecuteTemplate(&partFirst, "archiveposts", CtxFirst); err != nil {
		return nil, err
	}
	return &partFirst, nil
}

func (ln *mdhtArchivePostsNode) HasJsonBlock() bool {
	return false
}

func (ln *mdhtArchivePostsNode) JsonBlock() string {
	return ""
}

func (ln *mdhtArchivePostsNode) JsonBlockType() string {
	return ""
}
