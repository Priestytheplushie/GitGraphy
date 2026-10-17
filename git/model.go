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

type FileChange struct {
	Path   string
	Letter string
}

type StatusResult struct {
	Branch       string
	HasUpstream  bool
	UpstreamName string
	Ahead        int
	Behind       int
	Staged       []FileChange
	Unstaged     []FileChange
	Untracked    []FileChange
}

func (s *StatusResult) IsClean() bool {
	return len(s.Staged) == 0 && len(s.Unstaged) == 0 && len(s.Untracked) == 0
}

type FileStat struct {
	Name      string
	Additions int
	Deletions int
	IsBinary  bool
	IsLock    bool
}

type CommitDetails struct {
	Commit    Commit
	Message   string
	Stats     []FileStat
	TotalAdd  int
	TotalDel  int
	PatchText string
}

type BlameLine struct {
	LineNo    int
	Hash      string
	ShortHash string
	Author    string
	Subject   string
	Date      time.Time
	Code      string
}

type BlameResult struct {
	Path        string
	TotalLines  int
	Authors     []string
	CommitCount int
	Lines       []BlameLine
}