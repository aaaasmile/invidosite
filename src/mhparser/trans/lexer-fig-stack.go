package trans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"invido-site/src/idl"
	"log"
	"path"
	"strings"
	"text/template"
)

type mdhtFigStackNode struct {
	MdhtLineNode
	figItems    []string
	jsonImgPart string
}

// implements IMdhtmlTransfNode

func NewFigStackNode(preline string) *mdhtFigStackNode {
	res := mdhtFigStackNode{figItems: make([]string, 0)}
	arr := strings.Split(preline, "[")
	if len(arr) > 0 {
		res.before_link = arr[0]
	}
	return &res
}

func (ln *mdhtFigStackNode) AddParamString(parVal string) error {
	if parVal == "" {
		return fmt.Errorf("param is empty")
	}
	ln.figItems = append(ln.figItems, parVal)
	return nil
}

func (ln *mdhtFigStackNode) AddblockHtml(val string) error {
	if ln.after_link != "" {
		return fmt.Errorf("[AddblockHtml] already set")
	}
	ln.after_link = val
	return nil
}

func (ln *mdhtFigStackNode) Transform(templDir string) error {
	if templDir == "" {
		return fmt.Errorf("[Transform] templ dir is not set")
	}
	figs := make([]idl.ImgDataItem, 0)
	is_next_caption := false
	new_fig := idl.ImgDataItem{}
	for ix, item := range ln.figItems {
		if !is_next_caption {
			new_fig = idl.ImgDataItem{Name: item, Id: fmt.Sprintf("%02d", ix)}
			if err := new_fig.CalcReduced(); err != nil {
				return err
			}
			is_next_caption = true
		} else {
			new_fig.Caption = item
			is_next_caption = false
			figs = append(figs, new_fig)
		}
	}
	templName := path.Join(templDir, "transform.html")
	tmplPage := template.Must(template.New("FigStack").ParseFiles(templName))
	Ctx := idl.ImgDataItems{Images: figs}
	var partStack bytes.Buffer
	if err := tmplPage.ExecuteTemplate(&partStack, "figstack", Ctx); err != nil {
		return err
	}

	res := fmt.Sprintf("%s%s%s", ln.before_link, partStack.String(), ln.after_link)
	ln.block = res

	bdata, err := json.Marshal(Ctx)
	if err != nil {
		log.Println("[Transform] marshal error")
		return err
	}
	ln.jsonImgPart = string(bdata)
	return nil
}

func (ln *mdhtFigStackNode) JsonBlock() string {
	return ln.jsonImgPart
}

func (ln *mdhtFigStackNode) HasJsonBlock() bool {
	return len(ln.jsonImgPart) > 0
}

func (ln *mdhtFigStackNode) JsonBlockType() string {
	return "TtJsonImgs"
}
