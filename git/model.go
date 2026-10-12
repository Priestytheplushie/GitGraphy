package git

import "time"

type Ref struct {
	Name     string
	IsRemote bool
	IsTag    bool
	IsHEAD   bool
}

type Commit struct {
	Hash         string
	ShortHash    string
	ParentHashes []string
	AuthorName   string
	AuthorEmail  string
	When         time.Time
	Subject      string
	Refs         []Ref
}