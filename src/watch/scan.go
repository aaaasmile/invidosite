package watch

import (
	"database/sql"
	"fmt"
	"invido-site/src/conf"
	"invido-site/src/idl"
	"invido-site/src/mhproc"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ScanContent(force bool, debug bool) error {
	start := time.Now()
	post_src := conf.Current.ContentPost
	page_src := conf.Current.ContentPage
	bb := Builder{
		force: force,
		debug: debug,
	}
	if err := bb.InitDBData(); err != nil {
		return err
	}
	if err := bb.scanPostsMdHtml(post_src); err != nil {
		return err
	}
	if err := bb.scanPageMdHtml(page_src); err != nil {
		return err
	}
	if err := bb.liteDB.UpdateNumOfPostInTags(); err != nil {
		return err
	}
	var err error
	if bb.mapLinks, err = CreateMapLinks(bb.liteDB); err != nil {
		return err
	}
	log.Println("[ScanContent] completed, elapsed time ", time.Since(start))
	return nil
}

func (bb *Builder) scanPostsMdHtml(srcDir string) error {
	var err error
	bb.mdsFn = make([]string, 0)
	bb.mdsFn, err = getFilesinDir(srcDir, bb.mdsFn)
	if err != nil {
		return err
	}
	log.Printf("%d mdhtml posts  found ", len(bb.mdsFn))
	if bb.force {
		bb.liteDB.DeleteAllTagsToPost()
		bb.liteDB.DeleteAllPostItem()
		bb.liteDB.DeleteAllTags()
	}
	tx, err := bb.liteDB.GetTransaction()
	if err != nil {
		return err
	}
	if bb.mapLinks, err = CreateMapLinks(bb.liteDB); err != nil {
		return err
	}

	for _, item := range bb.mdsFn {
		if err := bb.scanPostItem(item, tx); err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("%d posts processed ", len(bb.mdsFn))
	return nil
}

func (bb *Builder) scanPostItem(mdHtmlFname string, tx *sql.Tx) error {
	if bb.debug {
		log.Println("[scanPostItem] file is ", mdHtmlFname)
	}

	mdhtml, err := os.ReadFile(mdHtmlFname)
	if err != nil {
		return err
	}
	//log.Println("read: ", mdhtml)
	prc := mhproc.NewMdHtmlProcess(false, bb.mapLinks)
	if err := prc.ProcessToHtml(string(mdhtml)); err != nil {
		log.Println("[scanPostItem] HTML error: ", err)
		return err
	}
	grm := prc.GetScriptGrammar()
	postItem := idl.PostItem{
		Title:    grm.Title,
		PostId:   grm.Id,
		DateTime: grm.Datetime,
	}

	subDir := conf.Current.PostDestSubDir
	arr, err := mhproc.GetDirNameArray(mdHtmlFname)
	if err != nil {
		return err
	}
	last_ix := len(arr) - 1
	dir_stack := []string{arr[last_ix-3], arr[last_ix-2], arr[last_ix-1], arr[last_ix]}
	remain := strings.Join(dir_stack, "/")
	postItem.Uri = fmt.Sprintf("/%s/%s/#", subDir, remain)
	//fmt.Println("*** uri is ", postItem.Uri)
	//fmt.Println("*** title is ", postItem.Title)
	bufRead := strings.NewReader(prc.HtmlGen)
	doc, err := html.Parse(bufRead)
	if err != nil {
		return err
	}
	traversePost(doc, &postItem)
	plink, ok := bb.mapLinks.MapPost[postItem.PostId]
	if !ok {
		err = bb.liteDB.InsertNewPost(tx, &postItem)
		if err != nil {
			return err
		}
	} else {
		postItem.Id = plink.Item.Id
		if bb.debug {
			log.Printf("[scanPostItem] ignore %s because already up to date", postItem.PostId)
		}
	}
	for _, single_tag := range grm.Tags {
		inserted, err := bb.liteDB.InsertOrUpdateTag(tx, single_tag, &postItem)
		if err != nil {
			return err
		}
		if inserted {
			pgid := fmt.Sprintf("tags-%s-PG", single_tag)
			pageItem := idl.PageItem{PageId: pgid, Md5: ""}
			err = bb.liteDB.UpdateMd5Page(tx, &pageItem)
			if err != nil {
				return err
			}
		} else {
			log.Printf("[scanPostItem] ignore changes for Tag %s on post id %s", single_tag, postItem.PostId)
		}
	}

	return nil
}

func traversePost(doc *html.Node, postItem *idl.PostItem) {
	// We need here the title, abstract and header image
	// Information are from parsing the mdhtml file
	section_first := false
	title_first := false
	has_title_img := false
	for n := range doc.Descendants() {
		if !title_first && n.Type == html.ElementNode && n.DataAtom == atom.Header {
			for _, a := range n.Attr {
				if a.Key == "class" {
					if a.Val == "withimg" {
						//fmt.Println("** has an image in title ")
						has_title_img = true
					}
					break
				}
			}
		}
		if !title_first && n.Type == html.ElementNode && n.DataAtom == atom.H1 {
			if n.FirstChild != nil {
				title := n.FirstChild.Data
				//fmt.Println("** title ", title)
				postItem.Title = title
			}
			title_first = true
		}
		if has_title_img && n.Type == html.ElementNode && n.DataAtom == atom.Img {
			has_title_img = false
			for _, a := range n.Attr {
				if a.Key == "src" {
					img_src := a.Val
					//fmt.Println("*** image in title ", img_src)
					postItem.TitleImgUri = strings.TrimRight(postItem.Uri, "/#")
					postItem.TitleImgUri = fmt.Sprintf("%s/%s", postItem.TitleImgUri, img_src)
					//fmt.Println("*** TitleImgUri ", postItem.TitleImgUri)
					break
				}
			}
		}
		if n.Type == html.ElementNode && n.DataAtom == atom.Section {
			section_first = true
			has_title_img = false
		}
		if section_first && n.Type == html.ElementNode && n.DataAtom == atom.P {
			if n.FirstChild != nil {
				abstract := n.FirstChild.Data
				abstract = strings.Trim(abstract, " ")
				abstract = strings.Trim(abstract, "\n")
				abstract = strings.Trim(abstract, " ")

				maxlen := 40
				cutPos := maxlen - 3
				if len(abstract) > cutPos {
					cutted := abstract[0:cutPos]
					rest := abstract[cutPos:]
					next_space_ix := strings.Index(rest, " ")
					if next_space_ix > 0 {
						cutted = fmt.Sprintf("%s%s", cutted, rest[0:next_space_ix])
					}
					abstract = fmt.Sprintf("%s...", cutted)
				}
				//fmt.Println("** abstract ", abstract)
				if len(abstract) > 4 {
					postItem.Abstract = abstract
				}
			}
			return
		}
	}
}

// pages

func (bb *Builder) scanPageMdHtml(srcDir string) error {
	var err error
	bb.mdsFn = make([]string, 0)
	bb.mdsFn, err = getFilesinDir(srcDir, bb.mdsFn)
	if err != nil {
		return err
	}
	log.Printf("[scanPageMdHtml] %d mdhtml pages found ", len(bb.mdsFn))
	if bb.force {
		bb.liteDB.DeleteAllPageItem()
	}
	tx, err := bb.liteDB.GetTransaction()
	if err != nil {
		return err
	}
	if bb.mapLinks, err = CreateMapLinks(bb.liteDB); err != nil {
		return err
	}

	for _, item := range bb.mdsFn {
		if err := bb.scanPageItem(srcDir, item, tx); err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("[scanPageMdHtml] %d page processed ", len(bb.mdsFn))
	return nil
}

func (bb *Builder) scanPageItem(srcDir string, mdHtmlFname string, tx *sql.Tx) error {
	if bb.debug {
		log.Println("[scanPageItem] file is ", mdHtmlFname)
	}

	mdhtml, err := os.ReadFile(mdHtmlFname)
	if err != nil {
		return err
	}
	//log.Println("read: ", mdhtml)
	prc := mhproc.NewMdHtmlProcess(false, bb.mapLinks)
	if err := prc.ProcessToHtml(string(mdhtml)); err != nil {
		log.Println("[scanPageItem] HTML error: ", err)
		return err
	}
	grm := prc.GetScriptGrammar()
	pageItem := idl.PageItem{
		Title:    grm.Title,
		PageId:   grm.Id,
		DateTime: grm.Datetime,
	}
	if item, ok := grm.CustomData["path"]; ok {
		pageItem.Uri = fmt.Sprintf("%s#", item)
	} else {
		subDir := conf.Current.PageDestSubDir
		arr, err := mhproc.GetDirNameArray(mdHtmlFname)
		if err != nil {
			return err
		}
		arr_src := strings.Split(srcDir, "/")
		equal_ix := 0
		for ini_ix, pp := range arr {
			if len(arr_src) <= ini_ix {
				break
			}
			if arr_src[ini_ix] == pp {
				equal_ix += 1
				continue
			} else {
				break
			}
		}
		last_ix := len(arr) - 1
		if arr[equal_ix] == "tags" {
			last_ix = len(arr)
			remain := strings.Join(arr[equal_ix:last_ix], "/")
			pageItem.Uri = fmt.Sprintf("/%s/#", remain)
		} else {
			remain := strings.Join(arr[equal_ix:last_ix], "/")
			pageItem.Uri = fmt.Sprintf("/%s/%s/#", subDir, remain)
		}
	}
	if _, ok := bb.mapLinks.MapPage[pageItem.PageId]; !ok {
		err = bb.liteDB.InsertNewPage(tx, &pageItem)
		if err != nil {
			return err
		}
	} else {
		if bb.debug {
			log.Printf("[scanPageItem] ignore %s because already up to date", pageItem.PageId)
		}
	}

	return nil
}
