package types

import "reflect"

type HandlerIdentity string

func (h HandlerIdentity) String() string {
	return string(h)
}

type HandlerInfo struct {
	HandlerName     string
	HandlerLongName HandlerIdentity
	Handler         interface{}
	ParamsKinds     []reflect.Kind
	ParamTypes      []reflect.Type
	ReturnTypes     []reflect.Type
	ReturnKinds     []reflect.Kind
	NumIn           int
	NumOut          int
}

type Activity HandlerInfo

type Workflow HandlerInfo
