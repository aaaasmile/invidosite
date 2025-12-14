package watch

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"invido-site/src/conf"
	"invido-site/src/idl"
	"invido-site/src/mhproc"
	"invido-site/src/syncdir"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/image/draw"
)

type WatcherMdHtml struct {
	CreatedHtmlFile string
	debug           bool
	dirContent      string
	staticSiteDir   string
	staticSubDir    string
	filesToIgnore   []string
	is_page         bool
	mapLinks        *idl.MapPagePostsLinks
}

func RunWatcher(targetDir string, subDir string, is_page bool, maplinks *idl.MapPagePostsLinks) error {
	if targetDir == "" {
		return fmt.Errorf("target dir is empty")
	}
	log.Println("watching ", targetDir)
	fs, err := os.Stat(targetDir)
	if err != nil {
		return err
	}
	if !fs.IsDir() {
		return fmt.Errorf("watch make sense only on a directory with content and images")
	}

	chShutdown := make(chan struct{}, 1)
	go func(chs chan struct{}) {
		wmh := WatcherMdHtml{dirContent: targetDir,
			debug:         conf.Current.Debug,
			staticSiteDir: conf.Current.StaticSiteDir,
			staticSubDir:  subDir,
			is_page:       is_page,
			mapLinks:      maplinks,
		}
		if err := wmh.doWatch(); err != nil {
			log.Println("Server is not watching anymore because: ", err)
		}
		log.Println("watch end")
		chs <- struct{}{}
	}(chShutdown)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	log.Println("Enter in server blocking loop")

loop:
	for {
		select {
		case <-sig:
			log.Println("stop because interrupt")
			break loop
		case <-chShutdown:
			log.Println("stop because service shutdown on watch")
			break loop
		}
	}

	log.Println("Bye, service")
	return nil
}

func (wmh *WatcherMdHtml) doWatch() error {
	log.Printf("setup watch on src: %s, update sub dir in static %s", wmh.dirContent, wmh.staticSubDir)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	err = watcher.Add(wmh.dirContent)
	if err != nil {
		return err
	}
	last_proc_eventname := ""
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watch event failed")
			}
			//log.Println("event:", event)

			if event.Has(fsnotify.Write) {
				log.Println("WRITE modified file:", event.Name)
				go func(ffname string) {
					if strings.Compare(last_proc_eventname, ffname) == 0 {
						log.Println("igone write event beacuse is duplicated")
						return
					}
					last_proc_eventname = ffname
					time.Sleep(200 * time.Millisecond)
					if err := wmh.processMdHtmlChange(ffname); err != nil {
						log.Println("[doWatch] error in processMdHtmlChange: ", err)
					}
					last_proc_eventname = ""
				}(event.Name)
			}
			if event.Has(fsnotify.Create) {
				log.Println("CREATE file:", event.Name)
				go func() {
					time.Sleep(200 * time.Millisecond) // some delay to wait until the writing FS process is finished
					if err := wmh.processNewImage(event.Name); err != nil {
						log.Println("[doWatch] error in processNewImage: ", err)
					}
				}()
			}
			if event.Has(fsnotify.Remove) {
				log.Println("REMOVE file:", event.Name)
				// do nothing: removing an asset items means that the mdhtml file should be also updated, in this case the image is synch
			}
			if event.Has(fsnotify.Rename) {
				log.Println("RENAME file:", event.Name)
				// do nothing: that is followed by a create event and the synch is done with the modification of the mdhtmlfile
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return err
			}
			log.Println("error:", err)
		}
	}
}

func (wmh *WatcherMdHtml) processNewImage(newFname string) error {
	_, err := os.Stat(newFname)
	if err != nil {
		return err
	}

	ext := filepath.Ext(newFname)
	if wmh.debug {
		log.Println("[processNewImage] extension new file ", ext)
	}
	isPng := strings.HasPrefix(ext, ".png")
	isJpeg := strings.HasPrefix(ext, ".jpg")
	isJpeg = isJpeg || strings.HasPrefix(ext, ".JPG")
	if !(isJpeg || isPng) {
		log.Println("[processNewImage] file ignored", newFname)
		return nil
	}
	for _, ignItem := range wmh.filesToIgnore {
		if strings.Compare(ignItem, newFname) == 0 {
			log.Println("[processNewImage] ignore file because already processed ", ignItem)
			return nil
		}
	}

	imageBytes, err := os.ReadFile(newFname)
	if err != nil {
		return err
	}
	newWidth := 320
	base_ff := filepath.Base(newFname)
	ff := strings.Replace(base_ff, ext, "", 1)
	if isJpeg {
		ff = fmt.Sprintf("%s_%d.jpg", ff, newWidth)
	} else if isPng {
		ff = fmt.Sprintf("%s_%d.png", ff, newWidth)
	} else {
		return fmt.Errorf("[processNewImage] image format %s not supported", ext)
	}
	ff_full_reduced := filepath.Join(wmh.dirContent, ff)

	var original_image image.Image
	if isJpeg {
		if original_image, err = jpeg.Decode(bytes.NewReader(imageBytes)); err != nil {
			return err
		}
	} else if isPng {
		if original_image, err = png.Decode(bytes.NewReader(imageBytes)); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("[processNewImage] image format %s not supported", ext)
	}
	if original_image.Bounds().Max.X <= newWidth {
		log.Println("[processNewImage] image is already on resize width or smaller", newWidth)
		newWidth = original_image.Bounds().Max.X
	}

	output, _ := os.Create(ff_full_reduced)
	defer output.Close()
	log.Println("[processNewImage] current image size ", original_image.Bounds().Max)
	ratiof := float32(original_image.Bounds().Max.X) / float32(newWidth)
	if ratiof == 0.0 {
		return fmt.Errorf("[processNewImage] invalid source image, attempt division by zero")
	}
	newHeightf := float32(original_image.Bounds().Max.Y) / ratiof
	newHeight := int(newHeightf)
	log.Printf("[processNewImage] new rect width %d height %d ratio %f ", newWidth, newHeight, ratiof)
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Rect, original_image, original_image.Bounds(), draw.Over, nil)
	if isJpeg {
		jpOpt := jpeg.Options{Quality: 100}
		if err = jpeg.Encode(output, dst, &jpOpt); err != nil {
			return err
		}
	} else if isPng {
		if err = png.Encode(output, dst); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("[processNewImage] image format %s not supported", ext)
	}
	wmh.filesToIgnore = append(wmh.filesToIgnore, ff_full_reduced)
	log.Println("[processNewImage] image created: ", ff_full_reduced)

	return nil
}

func (wmh *WatcherMdHtml) BuildFromMdHtml(mdHtmlFilename string) error {
	return wmh.processMdHtmlChange(mdHtmlFilename)
}

func (wmh *WatcherMdHtml) processMdHtmlChange(srcMdHtmlFname string) error {
	start := time.Now()
	if wmh.staticSiteDir == "" {
		return fmt.Errorf("[processMdHtmlChange] destination err: static site dir is empty")
	}
	if wmh.staticSubDir == "" {
		return fmt.Errorf("[processMdHtmlChange] destination err: sub dir is empty")
	}
	fi_src, err := os.Stat(srcMdHtmlFname)
	if err != nil {
		return err
	}
	src_path_name := fi_src.Name()
	ext := filepath.Ext(srcMdHtmlFname)
	if !strings.HasPrefix(ext, ".mdhtml") {
		log.Println("[processMdHtmlChange] file ignored", srcMdHtmlFname)
		return nil
	}
	mdhtml, err := os.ReadFile(srcMdHtmlFname)
	if err != nil {
		return err
	}
	//log.Println("read: ", mdhtml)
	prc := mhproc.NewMdHtmlProcess(false, wmh.mapLinks)
	if err := prc.ProcessToHtml(string(mdhtml)); err != nil {
		log.Println("[processMdHtmlChange] HTML error: ", err)
		return nil
	}
	grm := prc.GetScriptGrammar()
	if item, ok := grm.CustomData["path"]; ok {
		prc.RootStaticDir = fmt.Sprintf("static\\%s%s", wmh.staticSiteDir, item)
		src_path_name = item
	} else if rrpath, ok := grm.CustomData["rootpath"]; ok {
		prc.RootStaticDir = fmt.Sprintf("static\\%s%s", wmh.staticSiteDir, rrpath)
	} else {
		prc.RootStaticDir = fmt.Sprintf("static\\%s\\%s", wmh.staticSiteDir, wmh.staticSubDir)
	}
	log.Println("[processMdHtmlChange] Root dir is ", prc.RootStaticDir)
	if wmh.is_page {
		if err = prc.PageCreateOrUpdateStaticHtml(srcMdHtmlFname, src_path_name); err != nil {
			return err
		}
	} else {
		if err = prc.PostCreateOrUpdateStaticHtml(srcMdHtmlFname); err != nil {
			return err
		}
	}
	wmh.CreatedHtmlFile = prc.CreatedFnHtml
	if err := syncdir.SynchTargetDirWithSrcDir(prc.TargetDir, prc.SourceDir); err != nil {
		return err
	}
	log.Printf("[processMdHtmlChange] update traget html with size %d, duration %s ", len(prc.HtmlGen), time.Since(start))
	return nil
}
