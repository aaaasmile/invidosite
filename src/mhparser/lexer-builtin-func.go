package mhparser

import "sort"

func buildDescrInLex(l *L) {
	arr := make([]DescrFnItem, 0)

	fn := DescrFnItem{
		KeyName:       "link",
		Labels:        []string{"Url"},
		ItemTokenType: itemLinkSimple,
		IsMultiline:   false,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	fn = DescrFnItem{
		KeyName:       "figstack",
		ItemTokenType: itemFigStack,
		Labels:        []string{},
		IsMultiline:   true,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	fn = DescrFnItem{
		KeyName:       "linkcap",
		Labels:        []string{"Caption", "Url"},
		ItemTokenType: itemLinkCaption,
		IsMultiline:   false,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	fn = DescrFnItem{
		KeyName:       "linknext",
		Labels:        []string{"Caption", "Url"},
		ItemTokenType: itemLinkNext,
		IsMultiline:   false,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	fn = DescrFnItem{
		KeyName:       "linkimgnext",
		Labels:        []string{"Caption", "Url"},
		ItemTokenType: itemLinkImgNext,
		IsMultiline:   false,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	fn = DescrFnItem{
		KeyName:       "youtube",
		Labels:        []string{"VideoID"},
		ItemTokenType: itemYouTubeEmbed,
		IsMultiline:   false,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	fn = DescrFnItem{
		KeyName:       "latest_posts",
		Labels:        []string{"Title", "NumOfPosts"},
		ItemTokenType: itemLatestPosts,
		IsMultiline:   false,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	fn = DescrFnItem{
		KeyName:       "archive_posts",
		Labels:        []string{},
		ItemTokenType: itemArchivePosts,
		IsMultiline:   false,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	fn = DescrFnItem{
		KeyName:       "tag_posts",
		Labels:        []string{},
		ItemTokenType: itemTagPosts,
		IsMultiline:   false,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	fn = DescrFnItem{
		KeyName:       "single_taggedposts",
		Labels:        []string{"Title"},
		ItemTokenType: itemSingleTaggedPosts,
		IsMultiline:   false,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	fn = DescrFnItem{
		KeyName:       "img_link_run",
		Labels:        []string{"ImgSrc", "Url", "RunCaption"},
		ItemTokenType: itemImgLinkRun,
		IsMultiline:   false,
	}
	fn.NumParam = len(fn.Labels)
	arr = append(arr, fn)

	//
	// use arr2 for id calculation
	arr2 := make([]DescrFnItem, 0, len(arr))
	for ix, v := range arr {
		v.CustomID = ix + 1
		arr2 = append(arr2, v)
	}

	// final sort
	sort.Slice(arr2, func(i, j int) bool {
		return len(arr2[i].KeyName) > len(arr2[j].KeyName)
	})
	l.descrFns = arr2
}
