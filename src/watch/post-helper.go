package watch

import (
	"invido-site/src/db"
	"invido-site/src/idl"
	"log"
)

func CreateMapLinks(liteDB *db.LiteDB) (*idl.MapPagePostsLinks, error) {
	mapLinks := &idl.MapPagePostsLinks{
		MapPost:  map[string]idl.PostLinks{},
		MapPage:  map[string]*idl.PageItem{},
		ListPost: []idl.PostItem{},
		ListPage: []idl.PageItem{},
		Tags:     []idl.TagItem{},
	}
	var err error
	mapLinks.ListPost, err = liteDB.GetPostList()
	if err != nil {
		return nil, err
	}
	mapLinks.ListPage, err = liteDB.GetPageList()
	if err != nil {
		return nil, err
	}
	mapLinks.Tags, err = liteDB.GetTagList()
	if err != nil {
		return nil, err
	}
	mapLinks.MapTag, err = liteDB.GetTagPostMap(mapLinks.Tags)
	if err != nil {
		return nil, err
	}

	//fmt.Println("*** Posts ", mapLinks.ListPost)
	last_ix := len(mapLinks.ListPost) - 1
	prev_item := &idl.PostItem{}
	next_item := &idl.PostItem{}

	for ix, item := range mapLinks.ListPost {
		// remember that the order is descendig: most recent firts, oldest at the end
		postLinks := idl.PostLinks{
			Item: &item,
		}
		if last_ix > 0 {
			// at least 2 or more elements
			switch ix {
			case 0:
				prev_item = &mapLinks.ListPost[ix+1]
				postLinks.PrevLink = prev_item.Uri
				postLinks.PrevPostID = prev_item.PostId
			case last_ix:
				postLinks.NextLink = prev_item.Uri
				postLinks.NextPostID = prev_item.PostId
			default:
				next_item = &mapLinks.ListPost[ix+1]
				postLinks.NextLink = prev_item.Uri
				postLinks.NextPostID = prev_item.PostId
				postLinks.PrevLink = next_item.Uri
				postLinks.PrevPostID = next_item.PostId
			}
			prev_item = &mapLinks.ListPost[ix]
		}
		mapLinks.MapPost[item.PostId] = postLinks
	}
	//fmt.Println("*** map ", mapLinks.MapPost)
	log.Printf("[CreateMapLinks] Built map with %d posts", len(mapLinks.MapPost))

	for ix, item := range mapLinks.ListPage {
		mapLinks.MapPage[item.PageId] = &mapLinks.ListPage[ix]
	}
	log.Printf("[CreateMapLinks] Built map with %d page", len(mapLinks.MapPage))
	return mapLinks, nil
}
