package mhparser

import "fmt"

func (i TokenType) String() string {
	switch i {
	case itemText:
		return "itemText"
	case itemBuiltinFunction:
		return "itemBuiltinFunction"
	case itemVarValue:
		return "itemVarValue"
	case itemAssign:
		return "itemAssign"
	case itemComment:
		return "itemComment"
	case itemEmptyString:
		return "itemEmptyString"
	case itemEndOfStatement:
		return "itemEndOfStatement"
	case itemError:
		return "itemError"
	case itemFunctionName:
		return "itemFunctionName"
	case itemFunctionStartBlock:
		return "itemFunctionStartBlock"
	case itemFunctionEnd:
		return "itemFunctionEnd"
	case itemEOF:
		return "itemEOF"
	case itemArrayBegin:
		return "itemArrayBegin"
	case itemArrayEnd:
		return "itemArrayEnd"
	case itemVarName:
		return "itemVarName"
	case itemParamString:
		return "itemParamString"
	case itemBegMdHtml:
		return "itemBegMdHtml"
	case itemMdHtmlBlock:
		return "itemMdHtmlBlock"
	case itemMdHtmlBlockLine:
		return "itemMdHtmlBlockLine"
	case itemSeparator:
		return "itemSeparator"
	case itemEndOfBlock:
		return "itemEndOfBlock"
	case itemLinkSimple:
		return "itemLinkSimple"
	case itemFigStack:
		return "itemFigStack"
	case itemLinkCaption:
		return "itemLinkCaption"
	case itemYouTubeEmbed:
		return "itemYouTubeEmbed"
	case itemLatestPosts:
		return "itemLatestPosts"
	case itemArchivePosts:
		return "itemArchivePosts"
	case itemTagPosts:
		return "itemTagPosts"
	case itemSingleTaggedPosts:
		return "itemSingleTaggedPosts"
	}
	return fmt.Sprintf("TokenType %d undef", i)
}
