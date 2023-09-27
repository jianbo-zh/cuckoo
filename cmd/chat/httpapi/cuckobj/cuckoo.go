package cuckobj

import "github.com/jianbo-zh/dchat/cuckoo"

var global *cuckoo.Cuckoo

func SetCuckoo(cuckoo *cuckoo.Cuckoo) {
	global = cuckoo
}

func GetCuckoo() *cuckoo.Cuckoo {
	if global == nil {
		panic("global not init")
	}
	return global
}
