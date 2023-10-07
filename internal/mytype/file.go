package mytype

import (
	"fmt"
	"strings"
)

type FileID struct {
	HashAlgo  string
	HashValue string
	Extension string
}

func (f *FileID) String() string {
	return fmt.Sprintf("%s_%s.%s", f.HashAlgo, f.HashValue, strings.TrimLeft(f.Extension, "."))
}
