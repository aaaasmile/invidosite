package mhproc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"invido-site/src/idl"
	"invido-site/src/mhparser"
	"invido-site/src/util"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type MdHtmlProcess struct {
	debug             bool
	scrGramm          mhparser.ScriptGrammar
	HtmlGen           string
	ImgJsonGen        string
	templDir          string
	validateMandatory bool
	SourceDir         string
	RootStaticDir     string
	TargetDir         string
	CreatedFnHtml     string
	prevLink          string
	nextLink          string
	mapLinks          *idl.MapPagePostsLinks
}

func NewMdHtmlProcess(debug bool, mpLks *idl.MapPagePostsLinks) *MdHtmlProcess {
	res := MdHtmlProcess{
		debug:             debug,
		validateMandatory: true,
		templDir:          "templates/htmlgen",
		mapLinks:          mpLks,
	}
	return &res
}

func (mp *MdHtmlProcess) GetScriptGrammar() *mhparser.ScriptGrammar {
	return &mp.scrGramm
}

func (mp *MdHtmlProcess) ProcessToHtml(script string) error {
	if mp.debug {
		log.Println("[ProcessToHtml] is called with a script len ", len(script))
	}
	if script == "" {
		return fmt.Errorf("[ProcessToHtml] script is empty")
	}
	mp.scrGramm = mhparser.ScriptGrammar{
		Debug:    mp.debug,
		TemplDir: mp.templDir,
		MapLinks: mp.mapLinks,
	}
	if err := mp.scrGramm.ParseScript(script); err != nil {
		log.Println("[ProcessToHtml] Parser error")
		return err
	}
	if err := mp.scrGramm.CheckNorm(); err != nil {
		log.Println("[ProcessToHtml] Script structure error")
		return err
	}
	if err := mp.scrGramm.EvaluateParams(); err != nil {
		log.Println("[ProcessToHtml] EvaluateParams error")
		return err
	}
	if mp.validateMandatory {
		if mp.scrGramm.Title == "" {
			return fmt.Errorf("[ProcessToHtml] field 'title' in mdhtml is empty")
		}
		if mp.scrGramm.Id == "" {
			return fmt.Errorf("[ProcessToHtml] field 'id' in mdhtml is empty")
		}
		if mp.scrGramm.Datetime.Year() < 2010 {
			return fmt.Errorf("[ProcessToHtml] field 'datetime' is empty or invalid")
		}
	}
	if mp.debug {
		main_norm := mp.scrGramm.Norm["main"]
		log.Println("[ProcessToHtml] Parser nodes found: ", len(main_norm.FnsList))
	}
	return mp.parsedToHtml()
}

func (mp *MdHtmlProcess) parsedToHtml() error {
	if mp.debug {
		log.Println("create the HTML using parsed info")
	}
	normPrg := mp.scrGramm.Norm["main"]
	lines := []string{}
	img_data_items := []idl.ImgDataItem{}
	for _, stItem := range normPrg.FnsList {
		if stItem.Type == mhparser.TtHtmlVerbatim {
			lines = append(lines, stItem.Params[0].ArrayValue...)
		}
		if stItem.Type == mhparser.TtJsonBlock {
			labelJson := stItem.Params[0].Label
			if labelJson == "TtJsonImgs" {
				img_arr := idl.ImgDataItems{}
				bb := []byte(stItem.Params[0].Value)
				if err := json.Unmarshal(bb, &img_arr); err != nil {
					log.Println("[parsedToHtml] Unmarshal error")
					return err
				}
				img_data_items = append(img_data_items, img_arr.Images...)
			} else {
				return fmt.Errorf("[parsedToHtml] %s json block not supported", labelJson)
			}
			//fmt.Println("*** json item", stItem.Params[0].Value)
		}
	}
	if len(img_data_items) > 0 {
		imgs := idl.ImgDataItems{Images: img_data_items}
		data_img, err := json.Marshal(imgs)
		if err != nil {
			log.Println("[parsedToHtml] Marshal error")
			return err
		}
		mp.ImgJsonGen = string(data_img)
	}

	if mp.templDir != "" {
		return mp.htmlFromTemplate(lines)
	}
	// no template
	mp.HtmlGen = strings.Join(lines, "\n")
	mp.printGenHTML()

	return nil
}

func (mp *MdHtmlProcess) printGenHTML() {
	if mp.debug {
		fmt.Printf("***HTML***\n%s\n", mp.HtmlGen)
	}
}

func (mp *MdHtmlProcess) htmlFromTemplate(lines []string) error {
	templName := path.Join(mp.templDir, "post_or_page.html")
	var partFirst, partSecond, partThird, partMerged bytes.Buffer
	tmplPage := template.Must(template.New("Page").ParseFiles(templName))
	linesNoCr := []string{}
	for _, item := range lines {
		llnocr := strings.ReplaceAll(item, "\r", "")
		llnocr = strings.ReplaceAll(llnocr, "\n", "")
		if llnocr != "" {
			linesNoCr = append(linesNoCr, llnocr)
		}
	}
	//fmt.Println("*** lines ", strings.Join(linesNoCr, ","))
	CtxFirst := struct {
		Title string
		Lines []string
	}{
		Title: mp.scrGramm.Title,
		Lines: linesNoCr,
	}

	if err := tmplPage.ExecuteTemplate(&partFirst, "postbeg", CtxFirst); err != nil {
		return err
	}
	if err := mp.calculatePrevNextLink(); err != nil {
		return err
	}

	CtxSecond := struct {
		DateFormatted string
		DateTime      string
		PostId        string
		HasGallery    bool
		HasComments   bool
		HasPrev       bool
		HasNext       bool
		PrevURI       string
		NextURI       string
	}{
		DateTime:      mp.scrGramm.Datetime.Format("2006-01-02 15:00"),
		DateFormatted: util.FormatDateIt(mp.scrGramm.Datetime),
		PostId:        mp.scrGramm.Id,
		HasGallery:    len(mp.ImgJsonGen) > 0,
		HasComments:   true,
		HasPrev:       len(mp.prevLink) > 0,
		HasNext:       len(mp.nextLink) > 0,
		PrevURI:       mp.prevLink,
		NextURI:       mp.nextLink,
	}
	if val, ok := mp.scrGramm.CustomData["comments"]; ok {
		if val == "no" || val == "false" {
			CtxSecond.HasComments = false
		}
	}
	if CtxSecond.HasComments {
		if err := tmplPage.ExecuteTemplate(&partSecond, "postfinal", CtxSecond); err != nil {
			return err
		}
	} else {
		CtxSecond.DateFormatted = util.FormatDateTimeIt(time.Now())
		CtxSecond.DateTime = time.Now().Format("2006-01-02 15:00")
		if err := tmplPage.ExecuteTemplate(&partSecond, "pagefinal", CtxSecond); err != nil {
			return err
		}
	}
	if err := tmplPage.ExecuteTemplate(&partThird, "footer", CtxSecond); err != nil {
		return err
	}
	partFirst.WriteTo(&partMerged)
	partSecond.WriteTo(&partMerged)
	partThird.WriteTo(&partMerged)
	mp.HtmlGen = partMerged.String()
	mp.printGenHTML()
	return nil
}

func (mp *MdHtmlProcess) calculatePrevNextLink() error {
	mp.nextLink = ""
	mp.prevLink = ""
	if mp.mapLinks == nil {
		return nil
	}
	mapLinks := mp.mapLinks.MapPost
	if links, ok := mapLinks[mp.scrGramm.Id]; ok {
		mp.nextLink = links.NextLink
		mp.prevLink = links.PrevLink
	}
	return nil
}

func (mp *MdHtmlProcess) PageCreateOrUpdateStaticHtml(srcMdFullName string, fname string) error {
	mp.SourceDir = filepath.Dir(srcMdFullName)
	log.Println("[PageCreateOrUpdateStaticHtml] source dir for PAGE", mp.SourceDir)

	ext := filepath.Ext(fname)
	dir_for_target := strings.Replace(fname, ext, "", -1)
	dir_stack := []string{dir_for_target}
	if err := mp.checkOrCreateOutDir(dir_stack); err != nil {
		return err
	}
	log.Println("[PageCreateOrUpdateStaticHtml] target dir for PAGE", mp.TargetDir)
	if err := mp.createIndexHtml(); err != nil {
		return err
	}
	return nil
}

func (mp *MdHtmlProcess) CreateOnlyIndexStaticHtml() error {
	if mp.TargetDir == "" {
		return fmt.Errorf("target dir not set")
	}
	if mp.HtmlGen == "" {
		return fmt.Errorf("html dir not set")
	}
	log.Println("target dir for Index", mp.TargetDir)
	if err := mp.createIndexHtml(); err != nil {
		return err
	}
	return nil
}

func GetDirNameArray(sourceName string) ([]string, error) {
	sourceNameAgn := strings.ReplaceAll(sourceName, "\\", "/")
	arr := strings.Split(sourceNameAgn, "/")
	if len(arr) < 4 {
		return nil, fmt.Errorf("source filename is not conform to expected path: <optional/>yyyy/mm/dd/fname.mdhtml, but it is %s", sourceNameAgn)
	}
	//log.Println("Processing stack from source ", arr)
	last_ix := len(arr) - 1
	ext := path.Ext(arr[last_ix])
	last_dir := strings.Replace(arr[last_ix], ext, "", -1)
	arr[last_ix] = last_dir
	return arr, nil
}

func (mp *MdHtmlProcess) PostCreateOrUpdateStaticHtml(sourceName string) error {
	log.Println("[PostCreateOrUpdateStaticHtml] on ", sourceName)
	arr, err := GetDirNameArray(sourceName)
	if err != nil {
		return err
	}
	last_ix := len(arr) - 1
	dir_stack := []string{arr[last_ix-3], arr[last_ix-2], arr[last_ix-1], arr[last_ix]}
	if mp.debug {
		log.Println("dir structure for output ", dir_stack)
	}
	if err := mp.checkOrCreateOutDir(dir_stack); err != nil {
		return err
	}
	log.Println("target dir", mp.TargetDir)
	if err := mp.createIndexHtml(); err != nil {
		return err
	}
	if err := mp.createImageGalleryJson(); err != nil {
		return err
	}

	src_arr := make([]string, 0)
	src_arr = append(src_arr, arr[0:last_ix]...)
	mp.SourceDir = strings.Join(src_arr, "\\")
	log.Println("source dir", mp.SourceDir)

	return nil
}

func (mp *MdHtmlProcess) createIndexHtml() error {
	fname := path.Join(mp.TargetDir, "index.html")
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(mp.HtmlGen); err != nil {
		return err
	}
	log.Println("file created ", fname)
	mp.CreatedFnHtml = fname
	return nil
}

func (mp *MdHtmlProcess) createImageGalleryJson() error {
	if mp.ImgJsonGen == "" {
		return nil
	}
	fname := path.Join(mp.TargetDir, "photos.json")
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(mp.ImgJsonGen); err != nil {
		return err
	}
	log.Println("[createImageGalleryJson] file created ", fname)
	return nil
}

func (mp *MdHtmlProcess) checkOrCreateOutDir(dir_stack []string) error {
	dir_path := mp.RootStaticDir
	for _, item := range dir_stack {
		dir_path = path.Join(dir_path, item)
		//log.Println("check if out dir is here ", dir_path)
		if info, err := os.Stat(dir_path); err == nil && info.IsDir() {
			if mp.debug {
				log.Println("dir exist", dir_path)
			}
		} else {
			if mp.debug {
				log.Println("create dir", dir_path)
			}
			os.MkdirAll(dir_path, 0700)
		}
	}
	mp.TargetDir = dir_path
	return nil
}
