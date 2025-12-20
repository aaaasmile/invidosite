package trans

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"
)

type mdhtImgLinkRunNode struct {
	MdhtLineNode
	_args []string
}

func NewImgLinkRunNode(preline string) *mdhtImgLinkRunNode {
	res := mdhtImgLinkRunNode{_args: []string{}}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func (ln *mdhtImgLinkRunNode) AddParamString(parVal string) error {
	if len(ln._args) == 3 {
		return fmt.Errorf("invalid param size, already got 3 parameters")
	}
	ln._args = append(ln._args, parVal)
	return nil
}

func (ln *mdhtImgLinkRunNode) AddblockHtml(val string) error {
	if ln.after_link != "" {
		return fmt.Errorf("[AddblockHtml] already set")
	}
	ln.after_link = val
	return nil
}

func (ln *mdhtImgLinkRunNode) Transform(templDir string) error {
	if templDir == "" {
		return fmt.Errorf("[Transform] templ dir is not set")
	}
	if len(ln._args) < 2 {
		return fmt.Errorf("[Transform] at least 2 areguments are expected")
	}
	templName := path.Join(templDir, "transform.html")
	tmplPage := template.Must(template.New("ImgLinkRun").ParseFiles(templName))
	CtxFirst := struct {
		ImgSrc     string
		HrefLink   string
		RunCaption string
	}{
		ImgSrc:     ln._args[0],
		HrefLink:   ln._args[1],
		RunCaption: "RUN in Browser",
	}
	if len(ln._args) == 3 {
		CtxFirst.RunCaption = ln._args[2]
	}
	var partFirst bytes.Buffer
	if err := tmplPage.ExecuteTemplate(&partFirst, "imglinkrun", CtxFirst); err != nil {
		return err
	}

	res := fmt.Sprintf("%s%s%s", ln.before_link, partFirst.String(), ln.after_link)
	ln.block = res
	return nil
}

func (ln *mdhtImgLinkRunNode) HasJsonBlock() bool {
	return false
}

func (ln *mdhtImgLinkRunNode) JsonBlock() string {
	return ""
}

func (ln *mdhtImgLinkRunNode) JsonBlockType() string {
	return ""
}
