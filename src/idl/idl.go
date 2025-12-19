package idl

import (
	"fmt"
	"path"
	"strings"
	"time"
)

var (
	Appname = "invido-site"
	Buildnr = "00.001.20251214-00"
)

type ImgDataItem struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Redux   string `json:"redux"`
	Caption string `json:"caption"`
	SectId  string `json:"sectid"`
}

type ImgDataItems struct {
	Images []ImgDataItem `json:"images"`
}

type ImgDataSection struct {
	SectId string        `json:"id"`
	Val    []ImgDataItem `json:"val"`
}

type ImgDataSections struct {
	Images []ImgDataSection `json:"images"`
}

func (fg *ImgDataItem) CalcReduced() error {
	ext := path.Ext(fg.Name)
	if ext == "" {
		return fmt.Errorf("[calcReduced] extension on %s is empty, this is not supported", fg.Name)
	}
	bare_name := strings.Replace(fg.Name, ext, "", -1)
	fg.Redux = fmt.Sprintf("%s_320%s", bare_name, ext)
	fg.Id = fmt.Sprintf("%s_%s", bare_name, fg.Id)
	return nil
}

type PostItem struct {
	Id             int64
	PostId         string
	Title          string
	TitleImgUri    string
	DateTime       time.Time
	DateTimeRfC822 string
	Abstract       string
	Uri            string
	Md5            string
}

type PostLinks struct {
	PrevLink   string
	PrevPostID string
	NextLink   string
	NextPostID string
	Item       *PostItem
}

type MapPagePostsLinks struct {
	MapPost  map[string]PostLinks
	MapPage  map[string]*PageItem
	ListPost []PostItem
	ListPage []PageItem
	Tags     []TagItem
	MapTag   map[string][]*PostItem
}

type PageItem struct {
	Id             int64
	PageId         string
	Title          string
	Uri            string
	Md5            string
	Path           string
	DateTime       time.Time
	DateTimeRfC822 string
}

type TagItem struct {
	Id             int64
	Title          string
	NumOfPosts     int
	DateTime       time.Time
	DateTimeRfC822 string
	Uri            string
	Md5            string
}
