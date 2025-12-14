package trans

import (
	"bytes"
	"fmt"
	"invido-site/src/idl"
	"invido-site/src/util"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type PostWithData struct {
	DateFormatted string
	DateTimeTxt   string
	DateTime      time.Time
	Title         string
	Link          string
}

type mdhtLatestPostsNode struct {
	MdhtLineNode
	title_arg    string
	num_of_posts int
	mapLinks     *idl.MapPagePostsLinks
}

func NewLatestPostsNode(preline string, maplinks *idl.MapPagePostsLinks) *mdhtLatestPostsNode {
	res := mdhtLatestPostsNode{
		mapLinks: maplinks,
	}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func (ln *mdhtLatestPostsNode) AddParamString(parVal string) error {
	if ln.title_arg != "" {
		i, err := strconv.Atoi(parVal)
		if err != nil {
			return err
		}
		ln.num_of_posts = i
	} else {
		ln.title_arg = parVal
	}
	return nil
}

func (ln *mdhtLatestPostsNode) AddblockHtml(val string) error {
	if ln.after_link != "" {
		return fmt.Errorf("[AddblockHtml] already set")
	}
	ln.after_link = val
	return nil
}

func (ln *mdhtLatestPostsNode) Transform(templDir string) error {
	if templDir == "" {
		return fmt.Errorf("[Transform - LatestPosts] templ dir is not set")
	}
	if ln.num_of_posts == 0 {
		return fmt.Errorf("[Transform - LatestPosts] num of post is not defined or zero")
	}
	if ln.mapLinks == nil {
		return fmt.Errorf("[Transform - LatestPosts] map links are not set")
	}
	templName := path.Join(templDir, "transform.html")
	tmplPage := template.Must(template.New("Trf").ParseFiles(templName))
	latestPosts := []*PostWithData{}
	for ix, item := range ln.mapLinks.ListPost {
		pwd := PostWithData{
			DateFormatted: util.FormatDateIt(item.DateTime),
			DateTimeTxt:   item.DateTime.Format("2006-01-02 15:00"),
			DateTime:      item.DateTime,
			Title:         item.Title,
			Link:          item.Uri,
		}
		latestPosts = append(latestPosts, &pwd)
		if ix >= ln.num_of_posts {
			break
		}
	}
	CtxFirst := struct {
		Title       string
		LatestPosts []*PostWithData
	}{
		Title:       ln.title_arg,
		LatestPosts: latestPosts,
	}
	var partFirst bytes.Buffer
	if err := tmplPage.ExecuteTemplate(&partFirst, "latestposts", CtxFirst); err != nil {
		return err
	}

	res := fmt.Sprintf("%s%s%s", ln.before_link, partFirst.String(), ln.after_link)
	ln.block = res
	return nil
}

func (ln *mdhtLatestPostsNode) HasJsonBlock() bool {
	return false
}

func (ln *mdhtLatestPostsNode) JsonBlock() string {
	return ""
}

func (ln *mdhtLatestPostsNode) JsonBlockType() string {
	return ""
}
