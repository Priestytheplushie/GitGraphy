package git

import (
	"strings"
	"fmt"
	"errors"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repository struct {
	repo *gogit.Repository
}

func Open(path string) (*Repository, error) {
	r, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}
	return &Repository{repo: r}, nil
}

func (r *Repository) GetCommits(limit int) ([]Commit, error) {
	refMap, headBranch, err := r.collectRefs()
	if err != nil {
		return nil, err
	}

	head, err := r.repo.Head()
	if err != nil {
		return nil, err
	}

	cIter, err := r.repo.Log(&gogit.LogOptions{
		From: head.Hash(),
	})
	if err != nil {
		return nil, err
	}
	defer cIter.Close()

	var commits []Commit
	count := 0
	stopLimit := errors.New("limit reached")

	err = cIter.ForEach(func(c *object.Commit) error {
		if limit > 0 && count >= limit {
			return stopLimit
		}

		hashStr := c.Hash.String()
		shortHash := hashStr
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}

		subject := strings.Split(c.Message, "\n")[0]

		var parentHashes []string
		for _, ph := range c.ParentHashes {
			parentHashes = append(parentHashes, ph.String())
		}

		refs := refMap[hashStr]
		for i := range refs {
			if refs[i].Name == headBranch {
				refs[i].IsHEAD = true
			}
		}

		commits = append(commits, Commit{
			Hash:         hashStr,
			ShortHash:    shortHash,
			ParentHashes: parentHashes,
			AuthorName:   c.Author.Name,
			AuthorEmail:  c.Author.Email,
			When:         c.Author.When,
			Subject:      subject,
			Refs:         refs,
		})

		count++
		return nil
	})

	if err == stopLimit {
		err = nil
	}

	return commits, err
}

func (r *Repository) collectRefs() (map[string][]Ref, string, error) {
	refMap := make(map[string][]Ref)
	var headBranch string

	if head, err := r.repo.Head(); err == nil && head.Name().IsBranch() {
		headBranch = head.Name().Short()
	}

	iter, err := r.repo.References()
	if err != nil {
		return nil, "", err
	}
	defer iter.Close()

	_ = iter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() != plumbing.HashReference {
			return nil
		}

		name := ref.Name()
		shortName := name.Short()
		hash := ref.Hash().String()

		refItem := Ref{
			Name:     shortName,
			IsRemote: name.IsRemote(),
			IsTag:    name.IsTag(),
		}

		refMap[hash] = append(refMap[hash], refItem)
		return nil
	})

	return refMap, headBranch, nil
}

func (r *Repository) GetStatus() (*StatusResult, error) {
	w, err := r.repo.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := w.Status()
	if err != nil {
		return nil, err
	}

	res := &StatusResult{}

	head, err := r.repo.Head()
	if err == nil {
		if head.Name().IsBranch() {
			res.Branch = head.Name().Short()
		} else {
			res.Branch = "HEAD (detached)"
		}
	} else {
		res.Branch = "HEAD (empty)"
	}

	if res.Branch != "" && !head.Name().IsBranch() == false {
		r.calculateBranchSync(res, head.Hash())
	}

	for path, fileStatus := range status {
		if fileStatus.Staging != gogit.Unmodified && fileStatus.Staging != gogit.Untracked {
			res.Staged = append(res.Staged, FileChange{
				Path:   path,
				Letter: statusCodeToLetter(fileStatus.Staging),
			})
		}
		if fileStatus.Worktree == gogit.Untracked {
			res.Untracked = append(res.Untracked, FileChange{
				Path:   path,
				Letter: "U",
			})
			continue
		}

		if fileStatus.Worktree != gogit.Unmodified {
			res.Unstaged = append(res.Unstaged, FileChange{
				Path:   path,
				Letter: statusCodeToLetter(fileStatus.Worktree),
			})
		}
	}

	return res, nil
}

func statusCodeToLetter(code gogit.StatusCode) string {
	switch code {
	case gogit.Modified:
		return "M"
	case gogit.Added:
		return "A"
	case gogit.Deleted:
		return "D"
	case gogit.Renamed:
		return "R"
	default:
		return string(code)
	}
}

func (r *Repository) calculateBranchSync(res *StatusResult, localHash plumbing.Hash) {
	config, err := r.repo.Config()
	if err != nil {
		return
	}

	branchCfg := config.Branches[res.Branch]
	if branchCfg == nil || branchCfg.Remote == "" {
		return
	}

	remoteRefName := plumbing.NewRemoteReferenceName(branchCfg.Remote, res.Branch)
	remoteRef, err := r.repo.Reference(remoteRefName, true)
	if err != nil {
		return
	}

	res.HasUpstream = true
	res.UpstreamName = branchCfg.Remote + "/" + res.Branch

	upstreamHash := remoteRef.Hash()
	if localHash == upstreamHash {
		return
	}

	upstreamMap := make(map[string]bool)
	uIter, err := r.repo.Log(&gogit.LogOptions{From: upstreamHash})
	if err == nil {
		count := 0
		_ = uIter.ForEach(func(c *object.Commit) error {
			if count >= 50 {
				return fmt.Errorf("limit")
			}
			upstreamMap[c.Hash.String()] = true
			count++
			return nil
		})
		uIter.Close()
	}

	localMap := make(map[string]bool)
	lIter, err := r.repo.Log(&gogit.LogOptions{From: localHash})
	if err == nil {
		count := 0
		_ = lIter.ForEach(func(c *object.Commit) error {
			if count >= 50 {
				return fmt.Errorf("limit")
			}
			localMap[c.Hash.String()] = true
			if !upstreamMap[c.Hash.String()] {
				res.Ahead++
			}
			count++
			return nil
		})
		lIter.Close()
	}

	for hash := range upstreamMap {
		if !localMap[hash] {
			res.Behind++
		}
	}
}

func (r *Repository) GetAllCommits(limit int) ([]Commit, error) {
	refMap, headBranch, err := r.collectRefs()
	if err != nil {
		return nil, err
	}

	var tipHashes []plumbing.Hash
	refIter, err := r.repo.References()
	if err != nil {
		return nil, err
	}
	defer refIter.Close()

	_ = refIter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference && (ref.Name().IsBranch() || ref.Name().IsRemote()) {
			tipHashes = append(tipHashes, ref.Hash())
		}
		return nil
	})

	if len(tipHashes) == 0 {
		head, err := r.repo.Head()
		if err == nil {
			tipHashes = append(tipHashes, head.Hash())
		}
	}

	commitMap := make(map[string]*object.Commit)
	for _, tip := range tipHashes {
		cIter, err := r.repo.Log(&gogit.LogOptions{From: tip})
		if err != nil {
			continue
		}
		_ = cIter.ForEach(func(c *object.Commit) error {
			commitMap[c.Hash.String()] = c
			return nil
		})
		cIter.Close()
	}

	if len(commitMap) == 0 {
		return nil, nil
	}

	childCount := make(map[string]int)
	parentsMap := make(map[string][]string)

	for h, c := range commitMap {
		for _, ph := range c.ParentHashes {
			pStr := ph.String()
			if _, exists := commitMap[pStr]; exists {
				childCount[pStr]++
				parentsMap[h] = append(parentsMap[h], pStr)
			}
		}
	}

	var ready []string
	for h := range commitMap {
		if childCount[h] == 0 {
			ready = append(ready, h)
		}
	}

	var sorted []*object.Commit
	for len(ready) > 0 {
		bestIdx := 0
		for i := 1; i < len(ready); i++ {
			if commitMap[ready[i]].Author.When.After(commitMap[ready[bestIdx]].Author.When) {
				bestIdx = i
			}
		}

		currHash := ready[bestIdx]
		ready = append(ready[:bestIdx], ready[bestIdx+1:]...)
		currCommit := commitMap[currHash]
		sorted = append(sorted, currCommit)

		if limit > 0 && len(sorted) >= limit {
			break
		}
		for _, pHash := range parentsMap[currHash] {
			childCount[pHash]--
			if childCount[pHash] == 0 {
				ready = append(ready, pHash)
			}
		}
	}

	var result []Commit
	for _, c := range sorted {
		hashStr := c.Hash.String()
		shortHash := hashStr
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}

		subject := strings.Split(c.Message, "\n")[0]

		var parentHashes []string
		for _, ph := range c.ParentHashes {
			parentHashes = append(parentHashes, ph.String())
		}

		refs := refMap[hashStr]
		for i := range refs {
			if refs[i].Name == headBranch {
				refs[i].IsHEAD = true
			}
		}

		result = append(result, Commit{
			Hash:         hashStr,
			ShortHash:    shortHash,
			ParentHashes: parentHashes,
			AuthorName:   c.Author.Name,
			AuthorEmail:  c.Author.Email,
			When:         c.Author.When,
			Subject:      subject,
			Refs:         refs,
		})
	}

	return result, nil
}

func (r *Repository) GetCommitDetails(rev string) (*CommitDetails, error) {
	if rev == "" {
		rev = "HEAD"
	}

	hash, err := r.repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return nil, fmt.Errorf("revision not found: %s", rev)
	}

	commitObj, err := r.repo.CommitObject(*hash)
	if err != nil {
		return nil, err
	}

	commitTree, err := commitObj.Tree()
	if err != nil {
		return nil, err
	}

	var parentTree *object.Tree
	if commitObj.NumParents() > 0 {
		parent, err := commitObj.Parent(0)
		if err == nil {
			parentTree, _ = parent.Tree()
		}
	}
	if parentTree == nil {
		parentTree = &object.Tree{}
	}

	patch, err := parentTree.Patch(commitTree)
	if err != nil {
		return nil, err
	}

	refMap, headBranch, _ := r.collectRefs()
	hashStr := commitObj.Hash.String()
	shortHash := hashStr
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}

	refs := refMap[hashStr]
	for i := range refs {
		if refs[i].Name == headBranch {
			refs[i].IsHEAD = true
		}
	}

	var parentHashes []string
	for _, ph := range commitObj.ParentHashes {
		parentHashes = append(parentHashes, ph.String())
	}

	details := &CommitDetails{
		Commit: Commit{
			Hash:         hashStr,
			ShortHash:    shortHash,
			ParentHashes: parentHashes,
			AuthorName:   commitObj.Author.Name,
			AuthorEmail:  commitObj.Author.Email,
			When:         commitObj.Author.When,
			Subject:      strings.Split(commitObj.Message, "\n")[0],
			Refs:         refs,
		},
		Message:   strings.TrimSpace(commitObj.Message),
		PatchText: patch.String(),
	}

	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		name := ""
		if to != nil {
			name = to.Path()
		} else if from != nil {
			name = from.Path()
		}

		adds := 0
		dels := 0
		for _, chunk := range fp.Chunks() {
			switch chunk.Type() {
			case 1:
				adds += strings.Count(chunk.Content(), "\n")
			case 2:
				dels += strings.Count(chunk.Content(), "\n")
			}
		}

		isLock := isLockfile(name)

		details.Stats = append(details.Stats, FileStat{
			Name:      name,
			Additions: adds,
			Deletions: dels,
			IsBinary:  fp.IsBinary(),
			IsLock:    isLock,
		})

		details.TotalAdd += adds
		details.TotalDel += dels
	}

	return details, nil
}

func isLockfile(path string) bool {
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	switch name {
	case "go.sum", "package-lock.json", "yarn.lock", "pnpm-lock.yaml", "Cargo.lock", "composer.lock", "Gemfile.lock":
		return true
	default:
		return false
	}
}

func (r *Repository) GetBlame(filePath string) (*BlameResult, error) {
	head, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	cleanPath := filepath.ToSlash(filePath)
	cleanPath = strings.TrimPrefix(cleanPath, "./")

	res, err := gogit.Blame(commit, cleanPath)
	if err != nil {
		return nil, fmt.Errorf("could not blame file %q: %w", filePath, err)
	}

	commitCache := make(map[string]*object.Commit)
	authorSet := make(map[string]bool)

	blameRes := &BlameResult{
		Path:       cleanPath,
		TotalLines: len(res.Lines),
	}

	for i, l := range res.Lines {
		hashStr := l.Hash.String()
		shortHash := hashStr
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}

		c, exists := commitCache[hashStr]
		if !exists {
			c, _ = r.repo.CommitObject(l.Hash)
			commitCache[hashStr] = c
		}

		authorName := l.Author
		subject := ""
		if c != nil {
			if c.Author.Name != "" {
				authorName = c.Author.Name
			}
			subject = strings.Split(c.Message, "\n")[0]
		}

		authorSet[authorName] = true

		blameRes.Lines = append(blameRes.Lines, BlameLine{
			LineNo:    i + 1,
			Hash:      hashStr,
			ShortHash: shortHash,
			Author:    authorName,
			Subject:   subject,
			Date:      l.Date,
			Code:      l.Text,
		})
	}

	for a := range authorSet {
		blameRes.Authors = append(blameRes.Authors, a)
	}
	blameRes.CommitCount = len(commitCache)

	return blameRes, nil
}