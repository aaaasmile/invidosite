package watch

import (
	"bytes"
	"fmt"
	"html/template"
	"invido-site/src/conf"
	"invido-site/src/db"
	"invido-site/src/util"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func NewPost(title string, datepost string, watch_for_changes bool) error {
	if title == "" {
		return fmt.Errorf("title could not be null")
	}
	tit_compr := strings.ReplaceAll(title, " ", "")
	tit_compr = strings.ReplaceAll(tit_compr, ":", "_")
	tit_compr = strings.ReplaceAll(tit_compr, ";", "_")
	tit_compr = strings.ReplaceAll(tit_compr, ".", "_")
	tit_compr = strings.ReplaceAll(tit_compr, "-", "_")
	post := Post{
		Title:         title,
		TitleCompress: tit_compr,
		DatetimeOrig:  datepost,
		templDir:      "templates/mdhtml",
	}
	if err := post.setDateTimeFromString(datepost); err != nil {
		return err
	}

	if err := post.createNewPost("../posts-src"); err != nil {
		return err
	}
	if watch_for_changes {
		if err := post.editPost("../posts-src"); err != nil {
			return err
		}
	}
	return nil
}

func (pp *Post) setDateTimeFromString(datepost string) error {
	// expected something like: 2024-11-08 19:00
	//                          2024-11-08
	arr := strings.Split(datepost, " ")
	parsStr := "2006-01-02"
	if len(arr) == 2 {
		parsStr = "2006-01-02 15:00"
	}
	dt, err := time.Parse(parsStr, datepost)
	if err != nil {
		return err
	}
	pp.Datetime = dt

	return nil
}

func (pp *Post) createNewPost(targetRootDir string) error {
	log.Printf("[createNewPost] create new post '%s' on '%s'", pp.Title, pp.Datetime)
	var err error

	if pp.liteDB, err = db.OpenSqliteDatabase(fmt.Sprintf("..\\..\\%s", conf.Current.Database.DbFileName),
		conf.Current.Database.SQLDebug); err != nil {
		return err
	}
	if pp.mapLinks, err = CreateMapLinks(pp.liteDB); err != nil {
		return err
	}
	yy := fmt.Sprintf("%d", pp.Datetime.Year())
	mm := fmt.Sprintf("%02d", pp.Datetime.Month())
	dd := fmt.Sprintf("%02d", pp.Datetime.Day())
	contentDir := filepath.Join(targetRootDir, yy, mm, dd)
	log.Println("source post content dir ", contentDir)

	if err := os.MkdirAll(contentDir, 0700); err != nil {
		return err
	}
	log.Println("dir created ", contentDir)
	files, err := os.ReadDir(contentDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		itemAbs := filepath.Join(contentDir, f.Name())
		if info, err := os.Stat(itemAbs); err == nil && info.IsDir() {
			fmt.Println("*** ignore dir is ", f.Name())
		} else {
			return fmt.Errorf("[createNewPost] in this directory there is already some content %s", itemAbs)
		}
	}
	pp.contentDir = contentDir
	pp.postId = fmt.Sprintf("%d-%s-%s-%s-PS", pp.Datetime.Year()-2000, mm, dd, pp.TitleCompress)
	pp.mdhtmlName = fmt.Sprintf("%d-%s-%s-%s.mdhtml", pp.Datetime.Year()-2000, mm, dd, pp.TitleCompress)
	log.Println("content dir is empty, lets generate the file", pp.mdhtmlName)
	if err := pp.createMdHtml(); err != nil {
		return err
	}
	for ix, item := range pp.mapLinks.ListPost {
		if item.DateTime.Before(pp.Datetime) {
			prev_ix := ix - 1
			if prev_ix >= 0 {
				prev_Item := pp.mapLinks.ListPost[prev_ix]
				log.Println("[createNewPost] previous post_id is also invalid", prev_Item.PostId)
				pp.liteDB.DeleteTagOnPostId(prev_Item.PostId)
				pp.liteDB.DeletePostId(prev_Item.PostId)
			}
			log.Println("[createNewPost] post_id is before, now is invalid", item.PostId)
			pp.liteDB.DeleteTagOnPostId(item.PostId)
			pp.liteDB.DeletePostId(item.PostId)
			break
		}
	}
	return nil
}

func (pp *Post) createMdHtml() error {
	templName := path.Join(pp.templDir, "newpost.html")
	var partFirst bytes.Buffer
	tmplPage := template.Must(template.New("PostSrc").ParseFiles(templName))
	CtxFirst := struct {
		Title    string
		DateTime string
		Id       string
		DateLoc  string
	}{
		Title:    pp.Title,
		DateTime: pp.DatetimeOrig,
		Id:       pp.postId,
		DateLoc:  util.FormatDateIt(pp.Datetime),
	}

	if err := tmplPage.ExecuteTemplate(&partFirst, "postnew", CtxFirst); err != nil {
		return err
	}

	fname := path.Join(pp.contentDir, pp.mdhtmlName)
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
