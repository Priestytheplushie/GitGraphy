# Data Structures
GitGraphy uses the only data structre that go seems to have, Structs, or 'Structures", this document is more of a boilerplate then anything 

## Strcuts 
Our first structure is the `Ref` structure, it stores info for a git `ref`, which is bascially a pointer to a specific commit or tag. The `ref` structure is produced below
```go
type Ref struct {
	Name     string
	IsRemote bool
	IsTag    bool
	IsHEAD   bool
}
```
The second structure is the `Commit` structure, it represents a git `commit` object, it holds the hash, author, date, etc in string variables. Note that `When` is the time, and `Subject` is the commit message. 
```go
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
``` 
Our third structure is the `FileChange` structure, it tracks if files are changed, it only stores 2 strings, `Path` which is the file path, and the `Letter`, which is the badge 
```go
type FileChange struct {
	Path   string
	Letter string
}
```
The fourth structure, the `StatusResult`, which tracks the changes the **local branch** and the **remote branch**, it uses the `FileChange` structs in lists to track all file changes across `staged`, `unstaged`, and `untracked`, as well as using `int` primatives for ahead / behind
```go
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
```
The fifth data structure is the `FileStat` structure, this stores the additions and deletions, as well as if it's a binary file, so it can be ignored during `diff` displays 
```go
type FileStat struct {
	Name      string
	Additions int
	Deletions int
	IsBinary  bool
	IsLock    bool
}
```
The sixth structure is the `CommitDetails` struct, which contains a commit inside of it, and is used for extended information, such as the full commit message
```go
type CommitDetails struct {
	Commit    Commit
	Message   string
	Stats     []FileStat
	TotalAdd  int
	TotalDel  int
	PatchText string
}
```
The seventh structure is the `BlameLine`, which is used to store 1 line for the `gitgraphy blame` command's visual output, it stores the linked commit `hash` so it can attribute the author properly 
```go
type BlameLine struct {
	LineNo    int
	Hash      string
	ShortHash string
	Author    string
	Subject   string
	Date      time.Time
	Code      string
}
```
The eighth structure is the `BlameResult`, which just displays the summary for the `gitgraphy blame` command
```go
type BlameResult struct {
	Path        string
	TotalLines  int
	Authors     []string
	CommitCount int
	Lines       []BlameLine
}
```
The fianl structure is the `CompareResult`, which stores a comparison between two branches. It uses the other structs like `Commit` and `FileStat`. 
```go
type CompareResult struct {
	BaseRef            string
	TargetRef          string
	UnrelatedHistories bool
	CommitsAhead       []Commit
	AheadCount         int
	BehindCount        int
	Stats              []FileStat
	TotalAdd           int
	TotalDel           int
	PatchText          string
}
``` 

## Arrays 
The only real data storage structure we use besides primative `int`, `string`, `bool`, etc are `arrays`, or more specificly `slices` because they don't have a fixed size. 