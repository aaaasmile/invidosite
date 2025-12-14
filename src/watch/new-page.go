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

func NewPage(name string, datepage string, watch_for_changes bool) error {
	if name == "" {
		return fmt.Errorf("title could not be null")
	}

	name_compr := strings.ReplaceAll(name, " ", "")
	name_compr = strings.ReplaceAll(name_compr, ":", "-")
	name_compr = strings.ReplaceAll(name_compr, ";", "-")
	name_compr = strings.ReplaceAll(name_compr, ".", "-")
	page := Page{
		Name:         name,
		NameCompress: name_compr,
		DatetimeOrig: datepage,
		templDir:     "src/templates/mdhtml",
	}
	if err := page.setDateTimeFromString(datepage); err != nil {
		return err
	}
	page_src := conf.Current.ContentPage

	if err := page.createNewPage(page_src); err != nil {
		return err
	}
	if watch_for_changes {
		if err := page.editPage(page_src); err != nil {
			return err
		}
	}
	return nil
}

func (pg *Page) setDateTimeFromString(datepost string) error {
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
	pg.Datetime = dt

	return nil
}

func (pg *Page) createNewPage(targetRootDir string) error {
	log.Printf("[createNewPage] create new page '%s' on '%s'", pg.Name, pg.Datetime)
	var err error
	if pg.liteDB, err = db.OpenSqliteDatabase(conf.Current.Database.DbFileName,
		conf.Current.Database.SQLDebug); err != nil {
		return err
	}
	if pg.mapLinks, err = CreateMapLinks(pg.liteDB); err != nil {
		return err
	}

	contentDir := filepath.Join(targetRootDir, pg.NameCompress)
	log.Println("source page content dir ", contentDir)

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
			return fmt.Errorf("[createNewPage] in this directory there is already some content %s", itemAbs)
		}
	}
	pg.contentDir = contentDir
	pg.Id = fmt.Sprintf("%s-PG", pg.NameCompress)
	pg.mdhtmlName = fmt.Sprintf("%s.mdhtml", pg.NameCompress)
	log.Println("content dir is empty, lets generate the file", pg.mdhtmlName)
	if err := pg.createMdHtml(); err != nil {
		return err
	}
	return nil
}

func (pg *Page) createMdHtml() error {
	templName := path.Join(pg.templDir, "newpage.html")
	var partFirst bytes.Buffer
	tmplPage := template.Must(template.New("PagetSrc").ParseFiles(templName))
	CtxFirst := struct {
		Title    string
		DateTime string
		Id       string
		DateLoc  string
	}{
		Title:    pg.Name,
		DateTime: pg.DatetimeOrig,
		Id:       pg.Id,
		DateLoc:  util.FormatDateIt(pg.Datetime),
	}

	if err := tmplPage.ExecuteTemplate(&partFirst, "pagenew", CtxFirst); err != nil {
		return err
	}

	fname := path.Join(pg.contentDir, pg.mdhtmlName)
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
