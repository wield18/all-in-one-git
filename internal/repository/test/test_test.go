package test

import (
	"fmt"
	"testing"
)

type abc struct {
	def def
}

type def struct {
	abc string
}

func TestUpdateDef(t *testing.T) {
	aDef := def{abc: "gaojian"}
	aAbc := &abc{def: aDef}
	updateDef(aAbc)
	fmt.Println(aDef.abc)
}

func updateDef(aAbc *abc) {
	aAbc.def.abc = "dwdw"
}
