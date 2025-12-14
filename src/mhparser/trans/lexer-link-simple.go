package trans

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"
)

// -- Link Simple, implements  IMdhtmlTransfNode

type mdhtLinkSimpleNode struct {
	MdhtLineNode
	href_arg string
}

func NewLinkSimpleNode(preline string) *mdhtLinkSimpleNode {
	res := mdhtLinkSimpleNode{}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func (ln *mdhtLinkSimpleNode) AddParamString(parVal string) error {
	if ln.href_arg != "" {
		return fmt.Errorf("[AddParamString] parameter already set")
	}
	ln.href_arg = parVal
	return nil
}

func (ln *mdhtLinkSimpleNode) AddblockHtml(val string) error {
	if ln.after_link != "" {
		return fmt.Errorf("[AddblockHtml] already set")
	}
	ln.after_link = val
	return nil
}

func (ln *mdhtLinkSimpleNode) Transform(templDir string) error {
	if templDir == "" {
		return fmt.Errorf("[Transform] templ dir is not set")
	}
	templName := path.Join(templDir, "transform.html")
	tmplPage := template.Must(template.New("Link").ParseFiles(templName))
	CtxFirst := struct {
		HrefLink    string
		DisplayLink string
	}{
		HrefLink:    ln.href_arg,
		DisplayLink: ln.href_arg,
	}
	var partFirst bytes.Buffer
	if err := tmplPage.ExecuteTemplate(&partFirst, "linkbase", CtxFirst); err != nil {
		return err
	}

	res := fmt.Sprintf("%s%s%s", ln.before_link, partFirst.String(), ln.after_link)
	ln.block = res
	return nil
}

func (ln *mdhtLinkSimpleNode) HasJsonBlock() bool {
	return false
}

func (ln *mdhtLinkSimpleNode) JsonBlock() string {
	return ""
}

func (ln *mdhtLinkSimpleNode) JsonBlockType() string {
	return ""
}
