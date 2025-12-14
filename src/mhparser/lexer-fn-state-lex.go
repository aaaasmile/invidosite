package mhparser

import (
	"fmt"
	"strings"
)

type FnStatLex struct {
	fnName   string
	params   []ParamItem
	varName  string
	isAssign bool
	isArray  bool
}

func NewFnStatLex() *FnStatLex {
	return &FnStatLex{
		params: make([]ParamItem, 0),
	}
}

func (fnstlex *FnStatLex) ItemModVariableAsStatement() error {
	fnstlex.isAssign = true
	return nil
}

func (fnstlex *FnStatLex) ItemArrayBeginStatement(item Token) error {
	pp := ParamItem{}
	pp.VariableName = fnstlex.varName
	pp.IsVariable = fnstlex.varName != ""
	pp.IsArray = true
	pp.ArrayValue = []string{}
	fnstlex.isArray = true
	fnstlex.params = append(fnstlex.params, pp)

	return nil
}

func (fnstlex *FnStatLex) ItemArrayEndStatement(item Token) error {
	return nil
}

func (fnstlex *FnStatLex) ItemStringValueAssignStatement(item Token) error {
	if fnstlex.isArray {
		if len(fnstlex.params) != 1 {
			return fmt.Errorf("array value not initialized")
		}
		if !fnstlex.params[0].IsArray || fnstlex.params[0].ArrayValue == nil {
			return fmt.Errorf("param array value not initialized as an array")
		}
		fnstlex.params[0].ArrayValue = append(fnstlex.params[0].ArrayValue, item.Value)
		return nil
	}
	pp := ParamItem{}
	pp.VariableName = fnstlex.varName
	pp.IsVariable = fnstlex.varName != ""
	pp.Value = item.Value
	fnstlex.params = append(fnstlex.params, pp)
	return nil
}

func (fnstlex *FnStatLex) AddParamForVariableAssign() {
	pp := ParamItem{}
	pp.VariableName = fnstlex.varName
	pp.IsArray = fnstlex.isArray
	pp.IsVariable = fnstlex.varName != ""
	pp.IsUnset = true
	fnstlex.params = append(fnstlex.params, pp)
}

func (fnstlex *FnStatLex) ItemVariableAsStatement(item Token, ll *L, sn *ScriptGrammar) error {
	pp, err := fnstlex.paramAsVariable(item, ll, sn)
	if err != nil {
		return err
	}
	fnstlex.params = append(fnstlex.params, *pp)
	return nil
}

func (fnstlex *FnStatLex) paramAsVariable(item Token, ll *L, sn *ScriptGrammar) (*ParamItem, error) {
	pp := ParamItem{}
	pp.IsVariable = true
	pp.VariableName = item.Value
	var err error
	if pp.Label, err = fnstlex.nextLabelInParam(ll); err != nil {
		if sn.Debug {
			fmt.Println("*** [paramAsVariable] Error in fill ", err)
		}
		return nil, err
	}
	return &pp, nil
}

func (fnstlex *FnStatLex) ItemStringParamAsStatement(ll *L, val string) error {
	var err error
	pp := ParamItem{}
	pp.Value = val
	if pp.Label, err = fnstlex.nextLabelInParam(ll); err != nil {
		return err
	}
	fnstlex.params = append(fnstlex.params, pp)
	return nil
}

func (fn *FnStatLex) nextLabelInParam(l *L) (string, error) {
	res := ""
	//fmt.Printf("*** fn value %v\n", []byte(fn.FnName))
	//fmt.Printf("*** fn value %v\n", fn)
	for _, v := range l.descrFns {
		if strings.Compare(fn.fnName, v.KeyName) == 0 {
			if len(fn.params) >= v.NumParam {
				if !v.VariableArgs {
					return "", fmt.Errorf("paramater in %s are %d instead of %d", fn.fnName, len(fn.params)+1, v.NumParam)
				} else {
					ix := len(fn.params)
					res = fmt.Sprintf("3dots-%d", ix)
				}
			} else {
				ix := len(fn.params)
				res = v.Labels[ix]
				if res == "" {
					return "", fmt.Errorf("nextLabelInParam: statement not configuerd: %v", fn)
				}
			}
		}
	}

	if res == "" {
		return "", fmt.Errorf("[nextLabelInParam]: function not supported: %s", fn.fnName)
	}
	return res, nil
}
