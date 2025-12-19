package mhparser

import (
	"errors"
	"fmt"
	"invido-site/src/idl"
	"invido-site/src/mhparser/trans"
	"log"
	"strings"
	"unicode"
)

func lexStateMdHtmlOnLine(l *L) StateFunc {
	for {
		if nextSt, ok := lexMatchFnKey(l); ok {
			return nextSt
		}
		switch r := l.next(); {
		case r == EOFRune:
			l.emit(itemMdHtmlBlock)
			return nil
		case r == '\r' || r == '\n':
			l.rewind()
			l.emit(itemMdHtmlBlockLine)
			l.inc_line(r)
			l.next()
			l.ignore()
		}
	}
}
func lexStateAfterString(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case r == EOFRune || r == '\r' || r == '\n':
			return l.errorf("[lexStateAfterString] Expected next param or close curl")
		case r == ',':
			l.emit(itemSeparator)
			return lexStateSingleBeforeBegStr
		case r == ']':
			l.emit(itemEndOfBlock)
			return lexStateMdHtmlOnLine
		case r == '\'':
			l.emit(itemText)
		case unicode.IsSpace(r):
			l.ignore()
		default:
			return l.errorf("[lexStateAfterString] Malformed end of parameter: %s", l.source[l.start:l.position])
		}
	}
}

func lexStateSingleBeforeEndSquare(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case r == EOFRune || r == '\r' || r == '\n':
			return l.errorf("[lexStateSingleBeforeEndSquare] Expected close curl")
		case r == ']':
			l.emit(itemEndOfBlock)
			return lexStateMdHtmlOnLine
		case unicode.IsSpace(r):
			l.ignore()
		default:
			return l.errorf("[lexStateSingleBeforeEndSquare] Malformed end of parameter: %s", l.source[l.start:l.position])
		}
	}
}

func lexStateMultiAfterString(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case r == EOFRune || r == '\r' || r == '\n':
			l.ignore()
			l.inc_line(r)
		case r == ',':
			l.emit(itemSeparator)
			return lexStateMultiBeforeBegStr
		case r == ']':
			l.emit(itemEndOfBlock)
			return lexStateMdHtmlOnLine
		case r == '\'':
			l.emit(itemText)
		case unicode.IsSpace(r):
			l.ignore()
		default:
			return l.errorf("[lexStateMultiAfterString] Malformed end of parameter: %s", l.source[l.start:l.position])
		}
	}
}

func lexStateInParamString(l *L) StateFunc {
	ll := 0
	for {
		rleos := l.peek()
		//fmt.Println("***> ", rleos, ll)
		if ll == 0 && rleos == '\'' {
			l.emit(itemEmptyString)
			return lexStateAfterString
		}

		switch r := l.next(); {
		case r == EOFRune || r == '\r' || r == '\n':
			return l.errorf("[lexStateInParamString] expected end of string")
		case r == '\'':
			l.rewind()
			l.emit(itemParamString)
			return lexStateAfterString
		default:
			ll += 1
		}
	}
}

func lexStateMultiInParamString(l *L) StateFunc {
	ll := 0
	for {
		rleos := l.peek()
		//fmt.Println("***> ", rleos, ll)
		if ll == 0 && rleos == '\'' {
			l.emit(itemEmptyString)
			return lexStateMultiAfterString
		}

		switch r := l.next(); {
		case r == EOFRune || r == '\r' || r == '\n':
			return l.errorf("[lexStateMultiInParamString] expected end of string inside a single line")
		case r == '\'':
			l.rewind()
			l.emit(itemParamString)
			return lexStateMultiAfterString
		default:
			ll += 1
		}
	}
}

func lexStateSingleBeforeBegStr(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case unicode.IsSpace(r):
			l.ignore()
		case r == '\r' || r == '\n':
			return l.errorf("[lexStateSingleBeforeBegStr] Expected next param string")
		case r == '\'':
			l.emit(itemText)
			return lexStateInParamString
		default:
			return l.errorf("[lexStateSingleBeforeBegStr] Expected ( but got %s (Line %d)", l.source[l.start:l.position], l.scriptLine)
		}
	}
}

func lexStateMultiBeforeBegStr(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case unicode.IsSpace(r):
			l.ignore()
		case r == '\r' || r == '\n':
			l.ignore()
			l.inc_line(r)
		case r == '\'':
			l.emit(itemText)
			return lexStateMultiInParamString
		default:
			return l.errorf("[lexStateMultiBeforeBegStr] Expected ( but got %s (Line %d)", l.source[l.start:l.position], l.scriptLine)
		}
	}
}

func lexStateMdHtmlBeforeStm(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case r == EOFRune:
			return nil
		case r == '\r' || r == '\n':
			l.rewind()
			l.emit(itemBegMdHtml)
			l.inc_line(r)
			l.next()
			l.ignore()
			return lexStateMdHtmlOnLine
		case r == '-':
			// nothing
		default:
			return l.errorf("[lexStateMdHtmlBeforeStm] Unexpected char in data separator %s ", l.source[l.start:l.position])
		}
	}
}

func lexMatchMdHtmlKey(l *L) (StateFunc, bool) {
	khtml := "---"
	if strings.HasPrefix(l.source[l.position:], khtml) {
		return lexStateMdHtmlBeforeStm, true
	}
	return nil, false
}

/////////// ---- Grammar

type MdHtmlGram struct {
	Nodes             []trans.IMdhtmlLineNode
	_curr_Node        trans.IMdhtmlLineNode
	isMdHtmlCtx       bool
	debug             bool
	templDir          string
	fig_stack_counter int
	mapLinks          *idl.MapPagePostsLinks
}

func NewMdHtmlGr(templDir string, maplinks *idl.MapPagePostsLinks, debug bool) *MdHtmlGram {
	item := MdHtmlGram{
		Nodes:    make([]trans.IMdhtmlLineNode, 0),
		debug:    debug,
		templDir: templDir,
		mapLinks: maplinks,
	}
	return &item
}

func (mh *MdHtmlGram) processItem(item Token) (bool, error) {
	if item.Type == itemBegMdHtml {
		mh.isMdHtmlCtx = true
		return true, nil
	}
	if !mh.isMdHtmlCtx {
		return false, nil
	}
	switch {
	case item.Type == itemMdHtmlBlockLine:
		if err := mh.blockHtmlPart(item.Value); err != nil {
			return false, err
		}
	case item.Type == itemMdHtmlBlock:
		if err := mh.blockHtmlPart(item.Value); err != nil {
			return false, err
		}
	case item.Type == itemLinkSimple:
		mh._curr_Node = trans.NewLinkSimpleNode(item.Value)
	case item.Type == itemLinkCaption:
		mh._curr_Node = trans.NewLinkCaptionNode(item.Value)
	case item.Type == itemYouTubeEmbed:
		mh._curr_Node = trans.NewYouTubeNode(item.Value)
	case item.Type == itemFigStack:
		mh._curr_Node = trans.NewFigStackNode(item.Value, mh.fig_stack_counter)
		mh.fig_stack_counter += 1
	case item.Type == itemLatestPosts:
		mh._curr_Node = trans.NewLatestPostsNode(item.Value, mh.mapLinks)
	case item.Type == itemArchivePosts:
		mh._curr_Node = trans.NewArchivePostsNode(item.Value, mh.mapLinks)
	case item.Type == itemTagPosts:
		mh._curr_Node = trans.NewTagPostsNode(item.Value, mh.mapLinks)
	case item.Type == itemSingleTaggedPosts:
		mh._curr_Node = trans.NewSingleTaggedPostsNode(item.Value, mh.mapLinks)
	case item.Type == itemText:
		// ignore
	case item.Type == itemSeparator:
		// ignore
	case item.Type == itemParamString:
		if err := mh.addParameterString(item.Value); err != nil {
			return false, err
		}
	case item.Type == itemEndOfBlock:
		if err := mh.endOfBlock(item.Value); err != nil {
			return false, err
		}
	case item.Type == itemEOF:
		return false, nil
	case item.Type == itemError:
		return false, errors.New(item.Value)
	default:
		return false, fmt.Errorf("[MdHtmlGram] unsupported statement parser %q", item)
	}
	return true, nil
}

func (mh *MdHtmlGram) blockHtmlPart(val string) error {
	tt, ok := mh._curr_Node.(trans.IMdhtmlTransfNode)
	if ok {
		if err := tt.AddblockHtml(val); err != nil {
			return err
		}
		mh._curr_Node = trans.NewMdhtLineNode("undef")
	} else {
		mh.Nodes = append(mh.Nodes, trans.NewMdhtLineNode(val))
	}
	return nil
}

func (mh *MdHtmlGram) addParameterString(valPar string) error {
	tt, ok := mh._curr_Node.(trans.IMdhtmlTransfNode)
	if ok {
		return tt.AddParamString(valPar)
	}
	return fmt.Errorf("addParameterString is not supported in IMdhtmlLineNode interface (node=%v)", mh._curr_Node)
}

func (mh *MdHtmlGram) endOfBlock(valPar string) error {
	if valPar != "]" {
		return fmt.Errorf("[endOfBlock] end of block not  recognized")
	}
	mh.Nodes = append(mh.Nodes, mh._curr_Node)
	return nil
}

func (mh *MdHtmlGram) storeMdHtmlStatement(nrmPrg *NormPrg, scrGr *ScriptGrammar) error {
	if mh.debug {
		log.Println("[storeMdHtmlStatement] nodes len ", len(mh.Nodes))
	}

	stName := "mdhtml"
	fnStMdHtml := FnStatement{
		IsInternal: true,
		FnName:     stName,
		Type:       TtHtmlVerbatim,
		Params:     make([]ParamItem, 1),
	}
	linesParam := &fnStMdHtml.Params[0]
	linesParam.Label = "Lines"
	linesParam.IsArray = true
	linesParam.ArrayValue = make([]string, 0)
	for _, node := range mh.Nodes {
		tt, ok := node.(trans.IMdhtmlTransfNode)
		if ok {
			if err := tt.Transform(scrGr.TemplDir); err != nil {
				return err
			}
			linesParam.ArrayValue = append(linesParam.ArrayValue, tt.Block())
			if tt.HasJsonBlock() {
				fnStmJson := FnStatement{
					IsInternal: true,
					FnName:     "JsonBlock",
					Type:       TtJsonBlock,
					Params:     make([]ParamItem, 1),
				}
				jParam := &fnStmJson.Params[0]
				jParam.Label = tt.JsonBlockType()
				jParam.Value = tt.JsonBlock()

				nrmPrg.FnsList = append(nrmPrg.FnsList, fnStmJson)
				nrm_st_name, err := nrmPrg.statementInNormMap(stName, scrGr, len(nrmPrg.FnsList)-1)
				if err != nil {
					return err
				}
				if mh.debug {
					log.Println("[storeMdHtmlStatement] norm name", nrm_st_name)
				}
				//fmt.Println("*** stored json statement in norm")
			}
		} else {
			linesParam.ArrayValue = append(linesParam.ArrayValue, node.Block())
		}
	}

	nrmPrg.FnsList = append(nrmPrg.FnsList, fnStMdHtml)
	nrm_st_name, err := nrmPrg.statementInNormMap(stName, scrGr, len(nrmPrg.FnsList)-1)
	if err != nil {
		return err
	}
	if mh.debug {
		log.Println("[storeMdHtmlStatement] norm name", nrm_st_name)
	}
	return nil
}

func lexMatchFnKey(l *L) (StateFunc, bool) {
	for _, v := range l.descrFns {
		k := fmt.Sprintf("[%s", v.KeyName)
		if strings.HasPrefix(l.source[l.position:], k) { // make sure to parse the longest keyword first
			l.position += len(k)
			l.emitCustFn(v.ItemTokenType, v.CustomID)
			if v.IsMultiline {
				return lexStateMultiBeforeBegStr, true
			}
			if v.NumParam == 0 {
				return lexStateSingleBeforeEndSquare, true
			}
			return lexStateSingleBeforeBegStr, true
		}
	}
	return nil, false
}
