package trans

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"
)

type mdhtLinkCaptionNode struct {
	MdhtLineNode
	caption    string
	href_arg   string
	is_caption bool
	is_next    bool
	is_image   bool
}

func NewLinkCaptionNode(preline string) *mdhtLinkCaptionNode {
	res := mdhtLinkCaptionNode{is_caption: true}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func NewLinkNextNode(preline string) *mdhtLinkCaptionNode {
	res := mdhtLinkCaptionNode{is_caption: true, is_next: true}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func NewLinkImgNextNode(preline string) *mdhtLinkCaptionNode {
	res := mdhtLinkCaptionNode{is_caption: true, is_next: true, is_image: true}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func (ln *mdhtLinkCaptionNode) AddParamString(parVal string) error {
	if ln.is_caption {
		if ln.caption != "" {
			return fmt.Errorf("[AddParamString] parameter already set")
		}
		ln.caption = parVal
		ln.is_caption = false
		return nil
	}
	if ln.href_arg != "" {
		return fmt.Errorf("[AddParamString] parameter already set")
	}
	ln.href_arg = parVal
	return nil
}

func (ln *mdhtLinkCaptionNode) AddblockHtml(val string) error {
	if ln.after_link != "" {
		return fmt.Errorf("[AddblockHtml] already set")
	}
	ln.after_link = val
	return nil
}

func (ln *mdhtLinkCaptionNode) Transform(templDir string) error {
	if templDir == "" {
		return fmt.Errorf("[Transform] templ dir is not set")
	}
	templName := path.Join(templDir, "transform.html")
	tmplPage := template.Must(template.New("Link").ParseFiles(templName))
	CtxFirst := struct {
		HrefLink    string
		DisplayLink string
		OpenNewPage bool
	}{
		HrefLink:    ln.href_arg,
		DisplayLink: ln.caption,
		OpenNewPage: true,
	}
	if ln.is_next {
		CtxFirst.OpenNewPage = false
	}
	tempSection := "linkbase"
	if ln.is_image {
		CtxFirst.DisplayLink = fmt.Sprintf("<img src=\"%s\" />", ln.caption)
	}
	var partFirst bytes.Buffer
	if err := tmplPage.ExecuteTemplate(&partFirst, tempSection, CtxFirst); err != nil {
		return err
	}

	res := fmt.Sprintf("%s%s%s", ln.before_link, partFirst.String(), ln.after_link)
	ln.block = res
	return nil
}

func (ln *mdhtLinkCaptionNode) HasJsonBlock() bool {
	return false
}

func (ln *mdhtLinkCaptionNode) JsonBlock() string {
	return ""
}

func (ln *mdhtLinkCaptionNode) JsonBlockType() string {
	return ""
}
