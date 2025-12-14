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

type Post struct {
	Datetime      time.Time
	DatetimeOrig  string
	Title         string
	TitleCompress string
	mdhtmlName    string
	contentDir    string
	templDir      string
	postId        string
	mapLinks      *idl.MapPagePostsLinks
	liteDB        *db.LiteDB
}

func EditPost(datepost string) error {
	post := Post{
		DatetimeOrig: datepost,
	}
	if err := post.setDateTimeFromString(datepost); err != nil {
		return err
	}
	var err error
	if post.liteDB, err = db.OpenSqliteDatabase(fmt.Sprintf("..\\..\\%s", conf.Current.Database.DbFileName),
		conf.Current.Database.SQLDebug); err != nil {
		return err
	}
	if err := post.editPost("../posts-src"); err != nil {
		return err
	}
	return nil
}

func (pp *Post) editPost(contentRootDir string) error {
	log.Printf("[editPost] on '%s'", pp.Datetime)
	yy := fmt.Sprintf("%d", pp.Datetime.Year())
	mm := fmt.Sprintf("%02d", pp.Datetime.Month())
	dd := fmt.Sprintf("%02d", pp.Datetime.Day())
	contentDir := filepath.Join(contentRootDir, yy, mm, dd)
	log.Println("source post content dir ", contentDir)
	info, err := os.Stat(contentDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("[editPost] expected dir on %s", contentDir)
	}
	mapLinks, err := CreateMapLinks(pp.liteDB)
	if err != nil {
		return err
	}
	if err := RunWatcher(contentDir, conf.Current.PostDestSubDir, false, mapLinks); err != nil {
		log.Println("[editPost] error on watch")
		return err
	}
	return nil
}
