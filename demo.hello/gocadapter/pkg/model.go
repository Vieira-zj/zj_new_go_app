package pkg

import (
	"gorm.io/gorm"
)

type srvCoverMeta struct {
	Env       string
	Region    string
	Component string
	Branch    string
	Commit    string
}

// GocSrvCoverage .
type GocSrvCoverage struct {
	gorm.Model
	srvCoverMeta
	IsLatest    bool
	CovFilePath string
	CoverTotal  float32
}
