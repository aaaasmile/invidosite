package trans

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"
)

type mdhtYouTubeNode struct {
	MdhtLineNode
	videoid_arg string
}

func NewYouTubeNode(preline string) *mdhtYouTubeNode {
	res := mdhtYouTubeNode{}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func (ln *mdhtYouTubeNode) AddParamString(parVal string) error {
	if ln.videoid_arg != "" {
		return fmt.Errorf("[AddParamString] parameter already set")
	}
	ln.videoid_arg = parVal
	return nil
}

func (ln *mdhtYouTubeNode) AddblockHtml(val string) error {
	if ln.after_link != "" {
		return fmt.Errorf("[AddblockHtml] already set")
	}
	ln.after_link = val
	return nil
}

func (ln *mdhtYouTubeNode) Transform(templDir string) error {
	if templDir == "" {
		return fmt.Errorf("[Transform] templ dir is not set")
	}
	templName := path.Join(templDir, "transform.html")
	tmplPage := template.Must(template.New("Link").ParseFiles(templName))
	CtxFirst := struct {
		VideoID string
	}{
		VideoID: ln.videoid_arg,
	}
	var partFirst bytes.Buffer
	if err := tmplPage.ExecuteTemplate(&partFirst, "youtube", CtxFirst); err != nil {
		return err
	}

	res := fmt.Sprintf("%s%s%s", ln.before_link, partFirst.String(), ln.after_link)
	ln.block = res
	return nil
}

func (ln *mdhtYouTubeNode) HasJsonBlock() bool {
	return false
}

func (ln *mdhtYouTubeNode) JsonBlock() string {
	return ""
}

func (ln *mdhtYouTubeNode) JsonBlockType() string {
	return ""
}
