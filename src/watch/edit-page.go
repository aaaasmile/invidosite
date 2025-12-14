package watch

import (
	"fmt"
	"invido-site/src/conf"
	"invido-site/src/db"
	"invido-site/src/idl"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Page struct {
	Datetime     time.Time
	DatetimeOrig string
	Name         string
	NameCompress string
	mdhtmlName   string
	contentDir   string
	templDir     string
	Id           string
	liteDB       *db.LiteDB
	mapLinks     *idl.MapPagePostsLinks
}

func EditPage(name string) error {
	if name == "" {
		return fmt.Errorf("[EditPage] page name could not be empty")
	}
	page := Page{
		Name: name,
	}
	var err error
	if page.liteDB, err = db.OpenSqliteDatabase(fmt.Sprintf("..\\..\\%s", conf.Current.Database.DbFileName),
		conf.Current.Database.SQLDebug); err != nil {
		return err
	}
	if err := page.editPage("../page-src"); err != nil {
		return err
	}

	return nil
}

func (pg *Page) editPage(targetRootDir string) error {
	log.Printf("[editPage] on '%s'", pg.Name)
	contentDir := filepath.Join(targetRootDir, pg.Name)
	log.Println("source page content dir ", contentDir)
	log.Println("destination page is ", conf.Current.PageDestSubDir)
	info, err := os.Stat(contentDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("[editPage] expected dir on %s", contentDir)
	}
	mapLinks, err := CreateMapLinks(pg.liteDB)
	if err != nil {
		return err
	}
	if err := RunWatcher(contentDir, conf.Current.PageDestSubDir, true, mapLinks); err != nil {
		log.Println("[editPage] error on watch")
		return err
	}
	return nil
}
