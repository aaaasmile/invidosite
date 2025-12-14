package watch

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"fmt"
	"invido-site/src/conf"
	"invido-site/src/db"
	"invido-site/src/idl"
	"invido-site/src/mhproc"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type Builder struct {
	mdsFn    []string
	pages    []string
	liteDB   *db.LiteDB
	tx       *sql.Tx
	mapLinks *idl.MapPagePostsLinks
	force    bool
	debug    bool
}

func RebuildAll() error {
	start := time.Now()
	log.Println("[RebuildAll] the full site")
	post_src := conf.Current.ContentPost
	page_src := conf.Current.ContentPage

	bb := Builder{force: true}
	if err := bb.InitDBData(); err != nil {
		return err
	}
	if err := bb.scanPostsMdHtml(post_src); err != nil {
		return err
	}
	if err := bb.scanPageMdHtml(page_src); err != nil {
		return err
	}
	var err error
	if bb.mapLinks, err = CreateMapLinks(bb.liteDB); err != nil {
		return err
	}
	if err := bb.buildPosts(post_src); err != nil {
		return err
	}
	if err := bb.buildFeed(); err != nil {
		return err
	}
	if err := bb.buildTags(); err != nil {
		return err
	}
	if err := bb.buildPages(page_src); err != nil {
		return err
	}
	log.Println("[RebuildAll] completed, elapsed time ", time.Since(start))
	return nil
}

func PrepareForRsync(debug bool) error {
	start := time.Now()
	log.Println("[PrepareForRsync] start")
	if err := ScanContent(false, debug); err != nil {
		return err
	}
	if err := BuildTags(); err != nil {
		return err
	}
	if err := BuildPosts(); err != nil {
		return err
	}
	if err := BuildPages(false); err != nil {
		return err
	}
	if err := BuildMain(); err != nil {
		return err
	}
	log.Println("[PrepareForRsync] completed, elapsed time ", time.Since(start))
	return nil
}

func BuildTags() error {
	start := time.Now()
	log.Println("[BuildTags] start")

	bb := Builder{}
	if err := bb.InitDBData(); err != nil {
		return err
	}

	if err := bb.buildTags(); err != nil {
		return err
	}
	log.Println("[BuildFeed] completed, elapsed time ", time.Since(start))
	return nil
}

func BuildFeed() error {
	start := time.Now()
	log.Println("[BuildFeed] start")

	bb := Builder{}
	if err := bb.InitDBData(); err != nil {
		return err
	}

	if err := bb.buildFeed(); err != nil {
		return err
	}
	log.Println("[BuildFeed] completed, elapsed time ", time.Since(start))
	return nil
}

func BuildPosts() error {
	start := time.Now()
	log.Println("[BuildPosts] start")
	post_src := conf.Current.ContentPost

	bb := Builder{
		debug: conf.Current.Debug,
	}
	if err := bb.InitDBData(); err != nil {
		return err
	}

	if err := bb.buildPosts(post_src); err != nil {
		return err
	}
	if err := bb.buildFeed(); err != nil {
		return err
	}
	log.Println("[BuildPosts] completed, elapsed time ", time.Since(start))
	return nil
}

func BuildPages(force bool) error {
	start := time.Now()
	log.Println("[BuildPages] start")
	page_src := conf.Current.ContentPage

	bb := Builder{
		force: force,
	}
	if err := bb.InitDBData(); err != nil {
		return err
	}
	if err := bb.buildPages(page_src); err != nil {
		return err
	}

	log.Println("[BuildPages] completed, elapsed time ", time.Since(start))
	return nil
}

func BuildMain() error {
	start := time.Now()
	log.Println("[BuildMain] started")
	page_src := conf.Current.ContentPage

	bb := Builder{force: true}
	if err := bb.InitDBData(); err != nil {
		return err
	}
	if err := bb.builMdHtmlInDir(page_src + "/main"); err != nil {
		return err
	}
	if err := bb.builMdHtmlInDir(page_src + "/archivio"); err != nil {
		return err
	}
	log.Println("[BuildMain] completed, elapsed time ", time.Since(start))
	return nil
}

func (bb *Builder) InitDBData() error {
	var err error
	if bb.liteDB, err = db.OpenSqliteDatabase(conf.Current.Database.DbFileName,
		conf.Current.Database.SQLDebug); err != nil {
		return err
	}
	if bb.mapLinks, err = CreateMapLinks(bb.liteDB); err != nil {
		return err
	}
	return nil
}

func (bb *Builder) buildTags() error {
	//  create a mdhtml page pro tag
	//  create a page entry into the database
	var err error
	tx, err := bb.liteDB.GetTransaction()
	if err != nil {
		return err
	}
	pageItem := idl.PageItem{}
	for _, tag_item := range bb.mapLinks.Tags {
		if err := bb.createTagMdhtml(&tag_item, &pageItem); err != nil {
			return err
		}
		err = bb.liteDB.InsertNewPageIfNotExist(tx, &pageItem)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (bb *Builder) createTagMdhtml(tag_item *idl.TagItem, pageItem *idl.PageItem) error {
	mdhtmlName := fmt.Sprintf("%s.mdhtml", tag_item.Title)
	page_src := conf.Current.ContentPage
	contentDir := page_src + "/tags"

	templDir := "src/templates/mdhtml"
	templName := path.Join(templDir, "newtag.html")
	var partFirst bytes.Buffer
	tmplPage := template.Must(template.New("TagSrc").ParseFiles(templName))
	CtxFirst := struct {
		Title    string
		DateTime string
		PageId   string
		Path     string
	}{
		Title:    tag_item.Title,
		DateTime: tag_item.DateTime.Format("2006-01-02 15:00"),
		PageId:   fmt.Sprintf("tags-%s-PG", tag_item.Title),
		Path:     fmt.Sprintf("/tags/%s/", tag_item.Title),
	}
	pageItem.Title = CtxFirst.Title
	pageItem.PageId = CtxFirst.PageId
	pageItem.Uri = fmt.Sprintf("/tags/%s/#", CtxFirst.Title)
	pageItem.DateTime = tag_item.DateTime
	fname := path.Join(contentDir, mdhtmlName)
	if _, err := os.Stat(fname); err == nil {
		fmt.Printf("Tag File already %s exists\n", fname)
		return nil
	}

	if err := tmplPage.ExecuteTemplate(&partFirst, "tagnew", CtxFirst); err != nil {
		return err
	}

	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(partFirst.Bytes()); err != nil {
		return err
	}
	log.Println("file created ", fname)
	return nil
}

func (bb *Builder) buildFeed() error {
	log.Println("[rebuildFeed] start ")
	templDir := "src/templates/xml"
	templName := path.Join(templDir, "feed.xml")
	var partFirst bytes.Buffer
	tmplPage := template.Must(template.New("FeedSrc").ParseFiles(templName))

	if err := tmplPage.ExecuteTemplate(&partFirst, "feedbeg", bb.mapLinks.ListPost); err != nil {
		return err
	}
	rootStaticDir := fmt.Sprintf("static\\%s\\", conf.Current.StaticSiteDir)
	fname := path.Join(rootStaticDir, "feed")
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(partFirst.Bytes()); err != nil {
		return err
	}
	log.Println("file created ", fname)

	return nil
}

func (bb *Builder) buildPosts(srcDir string) error {
	bb.mdsFn = make([]string, 0)
	var err error
	bb.mdsFn, err = getFilesinDir(srcDir, bb.mdsFn)
	if err != nil {
		return err
	}
	bb.tx, err = bb.liteDB.GetTransaction()
	if err != nil {
		return err
	}
	log.Printf("%d mdhtml posts  found ", len(bb.mdsFn))
	for _, item := range bb.mdsFn {
		if err := bb.buildPost(item); err != nil {
			return err
		}
	}
	bb.tx.Commit()
	log.Printf("%d posts processed ", len(bb.mdsFn))
	return nil
}

func (bb *Builder) buildPages(srcDir string) error {
	bb.pages = make([]string, 0)
	var err error
	bb.pages, err = getFilesinDir(srcDir, bb.pages)
	if err != nil {
		return err
	}
	bb.tx, err = bb.liteDB.GetTransaction()
	if err != nil {
		return err
	}
	log.Printf("%d mdhtml pages found ", len(bb.pages))
	for _, item := range bb.pages {
		if err := bb.buildPage(item); err != nil {
			return err
		}
	}
	bb.tx.Commit()
	log.Printf("%d pages processed ", len(bb.pages))
	return nil
}

func (bb *Builder) builMdHtmlInDir(srcDir string) error {
	bb.pages = make([]string, 0)
	var err error
	bb.pages, err = getMdHtmlInDir(srcDir, bb.pages)
	if err != nil {
		return err
	}
	bb.tx, err = bb.liteDB.GetTransaction()
	if err != nil {
		return err
	}
	log.Printf("%d mdhtml pages found ", len(bb.pages))
	for _, item := range bb.pages {
		if err := bb.buildPage(item); err != nil {
			return err
		}
	}
	bb.tx.Commit()
	log.Printf("%d pages processed ", len(bb.pages))
	return nil
}

func (bb *Builder) buildPost(mdHtmlFname string) error {
	var err error
	wmh := WatcherMdHtml{
		debug:         conf.Current.Debug,
		staticSiteDir: conf.Current.StaticSiteDir,
		is_page:       false,
		mapLinks:      bb.mapLinks,
	}
	is_same := true
	postItem := &idl.PostItem{}
	postItem, is_same, err = bb.getPostItemAndCheckMd5(mdHtmlFname)
	if err != nil {
		return err
	}
	wmh.staticSubDir = conf.Current.PostDestSubDir
	if !bb.force && is_same {
		if bb.debug {
			log.Println("[buildPost] ignore because unchanged", mdHtmlFname)
		}
		return nil
	}
	if err := wmh.BuildFromMdHtml(mdHtmlFname); err != nil {
		return err
	}
	if (postItem.PostId != "") && !is_same {
		if err := bb.liteDB.UpdateMd5Post(bb.tx, postItem); err != nil {
			return err
		}
	}
	log.Println("[buildPost] created HTML: ", wmh.CreatedHtmlFile)
	return nil
}

func (bb *Builder) buildPage(mdHtmlFname string) error {
	wmh := WatcherMdHtml{
		debug:         conf.Current.Debug,
		staticSiteDir: conf.Current.StaticSiteDir,
		is_page:       true,
		mapLinks:      bb.mapLinks,
	}
	wmh.staticSubDir = conf.Current.PageDestSubDir
	pageItem, is_same, err := bb.getPageItemAndCheckMd5(mdHtmlFname)
	if err != nil {
		return err
	}

	if !bb.force && is_same {
		if bb.debug {
			log.Println("[buildPage] ignore because unchanged", mdHtmlFname)
		}
		return nil
	}
	if err := wmh.BuildFromMdHtml(mdHtmlFname); err != nil {
		return err
	}
	if (pageItem.PageId != "") && !is_same {
		if err := bb.liteDB.UpdateMd5Page(bb.tx, pageItem); err != nil {
			return err
		}
	}
	log.Println("[buildPage] created HTML: ", wmh.CreatedHtmlFile)
	return nil
}

func (bb *Builder) getPostItemAndCheckMd5(mdHtmlFname string) (*idl.PostItem, bool, error) {
	f, err := os.Open(mdHtmlFname)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, false, err
	}
	mMd5 := string(h.Sum(nil))

	mdhtml, err := os.ReadFile(mdHtmlFname)
	if err != nil {
		return nil, false, err
	}
	prc := mhproc.NewMdHtmlProcess(false, bb.mapLinks)
	if err := prc.ProcessToHtml(string(mdhtml)); err != nil {
		log.Println("[hasSamePostMd5] ProcessToHtml error: ", err)
		return nil, false, err
	}
	gr := prc.GetScriptGrammar()
	mMd5Db, ok := bb.mapLinks.MapPost[gr.Id]
	if !ok {
		return nil, false, fmt.Errorf("[hasSamePostMd5] post id %s not found in MapLinks. Is the post table in db syncronized?", gr.Id)
	}
	same := mMd5 == mMd5Db.Item.Md5
	postItem := idl.PostItem{PostId: gr.Id, Md5: mMd5}
	return &postItem, same, nil
}

func (bb *Builder) getPageItemAndCheckMd5(mdHtmlFname string) (*idl.PageItem, bool, error) {
	f, err := os.Open(mdHtmlFname)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, false, err
	}
	mMd5 := string(h.Sum(nil))

	mdhtml, err := os.ReadFile(mdHtmlFname)
	if err != nil {
		return nil, false, err
	}
	prc := mhproc.NewMdHtmlProcess(false, bb.mapLinks)
	if err := prc.ProcessToHtml(string(mdhtml)); err != nil {
		log.Println("[getPageItemAndCheckMd5] ProcessToHtml error: ", err)
		return nil, false, err
	}
	gr := prc.GetScriptGrammar()
	mMd5Db, ok := bb.mapLinks.MapPage[gr.Id]
	if !ok {
		return nil, false, fmt.Errorf("[getPageItemAndCheckMd5] post id %s not found in MapLinks. Is the post table in db syncronized?", gr.Id)
	}
	same := mMd5 == mMd5Db.Md5
	pageItem := idl.PageItem{
		PageId: gr.Id,
		Md5:    mMd5,
	}
	if item, ok := gr.CustomData["path"]; ok {
		pageItem.Path = item
	}
	return &pageItem, same, nil
}

func getFilesinDir(dirAbs string, ini []string) ([]string, error) {
	r := ini
	//log.Println("Scan dir ", dirAbs)
	files, err := os.ReadDir(dirAbs)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		itemAbs := path.Join(dirAbs, f.Name())
		if info, err := os.Stat(itemAbs); err == nil && info.IsDir() {
			//fmt.Println("** Sub dir found ", f.Name())
			r, err = getFilesinDir(itemAbs, r)
			if err != nil {
				return nil, err
			}
		} else {
			//fmt.Println("** file is ", f.Name())
			ext := filepath.Ext(itemAbs)
			if strings.HasPrefix(ext, ".mdhtml") {
				r = append(r, path.Join(dirAbs, f.Name()))
			}
		}
	}
	return r, nil
}

func getMdHtmlInDir(dirAbs string, ini []string) ([]string, error) {
	r := ini
	//log.Println("Scan dir ", dirAbs)
	files, err := os.ReadDir(dirAbs)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		itemAbs := path.Join(dirAbs, f.Name())
		if info, err := os.Stat(itemAbs); err == nil && info.IsDir() {
			//fmt.Println("** Sub dir found ", f.Name())
			r, err = getMdHtmlInDir(itemAbs, r)
			if err != nil {
				return nil, err
			}
		} else {
			//fmt.Println("** file is ", f.Name())
			ext := filepath.Ext(itemAbs)
			if strings.HasPrefix(ext, ".mdhtml") {
				r = append(r, path.Join(dirAbs, f.Name()))
			}
		}
	}
	return r, nil
}
