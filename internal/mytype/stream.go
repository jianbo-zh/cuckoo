package mytype

import "time"

const (
	DialTimeout              time.Duration = 3 * time.Second
	PbioReaderMaxSizeNormal  int           = 4 * 1024   // 4KB
	PbioReaderMaxSizeMessage int           = 500 * 1024 // 200KB
)
