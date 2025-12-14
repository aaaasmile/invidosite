package syncdir

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type SyncDir struct {
	itemsInDir []os.FileInfo
	dirName    string
}

func (sd *SyncDir) contains(fi os.FileInfo) (os.FileInfo, bool) {
	finame := fi.Name()
	for _, fi_trg := range sd.itemsInDir {
		tgtname := fi_trg.Name()
		if strings.Compare(tgtname, finame) == 0 {
			return fi_trg, true
		}
	}
	return nil, false
}

func (sd *SyncDir) preserveOnlySameSrc(src *SyncDir) error {
	for _, fi_trg := range sd.itemsInDir {
		src_item, ok := src.contains(fi_trg)
		to_remove := !ok
		if ok {
			if fi_trg.ModTime().Before(src_item.ModTime()) {
				log.Println("target asset is obsolete", fi_trg.Name())
				to_remove = true
			}
		}
		if to_remove {
			fn_to_delete := filepath.Join(sd.dirName, fi_trg.Name())
			if err := os.Remove(fn_to_delete); err != nil {
				return err
			}
			log.Println("removed file ", fn_to_delete)
		}
	}
	return nil
}

func (sd *SyncDir) copyInDir(src_dirname string, fi_src os.FileInfo) (string, error) {
	full_src_name := filepath.Join(src_dirname, fi_src.Name())
	from, err := os.Open(full_src_name)
	if err != nil {
		return "", err
	}
	defer from.Close()

	target_full_name := filepath.Join(sd.dirName, fi_src.Name())
	to, err := os.OpenFile(target_full_name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer to.Close()

	if _, err := io.Copy(to, from); err != nil {
		return "", err
	}

	return target_full_name, nil
}

func (sd *SyncDir) copyNewItemsFromSrc(src *SyncDir) error {
	for _, fi_src := range src.itemsInDir {
		if _, ok := sd.contains(fi_src); !ok {

			full_trg_name, err := sd.copyInDir(src.dirName, fi_src)
			if err != nil {
				return err
			}
			log.Println("file copied into the target ", full_trg_name)
		}
	}
	return nil
}

func (sd *SyncDir) populateDir(dir string) error {
	sd.dirName = dir
	sd.itemsInDir = make([]fs.FileInfo, 0)
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ext := path.Ext(f.Name())
		if (ext != ".jpg") && (ext != ".png") && (ext != ".pdf") {
			continue
		}
		fi, err := os.Stat(filepath.Join(dir, f.Name()))
		if err != nil {
			return err
		}
		sd.itemsInDir = append(sd.itemsInDir, fi)
	}
	log.Printf("assets in dir %s count %d ", dir, len(sd.itemsInDir))
	return nil
}

func SynchTargetDirWithSrcDir(targetDir string, srcDir string) error {
	srcFiInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	if !srcFiInfo.IsDir() {
		return fmt.Errorf("cannot get src dir on %s", srcDir)
	}
	src := SyncDir{}
	dst := SyncDir{}
	if err := src.populateDir(srcDir); err != nil {
		return err
	}
	if err := dst.populateDir(targetDir); err != nil {
		return err
	}
	if err := dst.preserveOnlySameSrc(&src); err != nil {
		return err
	}
	if err := dst.copyNewItemsFromSrc(&src); err != nil {
		return err
	}

	return nil
}
