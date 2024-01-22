package direct

import (
	"encoding/json"
	"fmt"
)

const debug = false
const asserted = true

func Assert(flag bool, infos ...interface{}) {
	if !flag {
		panic(fmt.Sprintf("assert fail %v", infos))
	}
}

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicErr1[R interface{}](r R, err error) R {
	if err != nil {
		panic(err)
	}
	return r
}

func Jsonify(val interface{}) string {
	marshal, err := json.Marshal(val)
	if err != nil {
		panic(fmt.Sprintf("cannot jsonify %v", val))
	}
	return string(marshal)
}
