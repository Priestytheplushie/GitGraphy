# Commands

GItGraphy includes many CLI commands which you can use to inspect your Git History. All of GitGraphy's commands are **Read Only** and cannot modify git objects. 

## log 
The `gitgraphy log` command displays a visually organized list of recent git commits. By default, it displays `20` recent commits by default, but also supports the `-n` or `--limit` flag to show more or less commits

>[!WARNING]
>Passing the `-n` flag or `--limit` flag with `0` will print **all** commits, which can cause lag or freezes

### Usage 
```bash
Show a clean, decorated commit history

Usage:
  gitgraphy log [flags]

Flags:
  -h, --help        help for log
  -n, --limit int   Number of commits to show (default 20)
```

### Usage Examples
```bash
# Print out the last 20 commits
gitgraphy log

# Print out the last 5 commits
gitgraphy log -n 5

# Print out all commits
gitgraphy log -n 0
```

## compare
The `gitgraphy compare` command lets you compare two branches and see which files have changed, their diff count, and the actual diffs. By default, it will display all modified files with their full diff, but you can pass the `-s` flag or `--stat` to only get the summary. By default, `compare` assumes the base branch is `main`, and then passing in another branch will compare against that, however you can also pass in a `base` and a `target` branch for comparisons

### Usage
```bash
Compare branches or revisions like a GitHub PR preview

Usage:
  gitgraphy compare [base] <target> [flags]

Flags:
  -h, --help   help for compare
  -s, --stat   Show file impact summary only without code diffs
``` 

### Usage Examples
```bash
# Compare main <- develop
gitgraphy compare develop

# Compare feature/auth <- bugfix
gitgraphy compare feature/auth bugfix

# Compare main <- auth with summary only
gitgraphy compare main feature/auth -s 
```

## blame
The `gitgraphy blame` command allows you to see which `Commit` and `Author` modified that speicfic line of a file. There are **no flags** for this command, you just pass the `FileName` with the command.

### Usage 
```bash
View GitLens-style visual blame with block headers and syntax highlighting

Usage:
  gitgraphy blame <file> [flags]

Flags:
  -h, --help   help for blame
  ```

### Usage Examples
```bash
# Check on main.go
gitgraphy blame main.go 
```

## graph
The `gitgraphy graph` command lets you view a visual git graph of your commit history. On the graph, you can visually see merges. Like `log`, all commit hashes, names and author, as well as the date. You can pass the `-n` or `--limit` flag to show more or less commits 

>[!WARNING]
>Passing the `-n` flag or `--limit` flag with `0` will print **all** commits, which can cause lag or freezes

### Usage
```bash
Visualize commit history as a multi-branch transit map

Usage:
  gitgraphy graph [flags]

Flags:
  -h, --help        help for graph
  -n, --limit int   Number of commits to show (0 for all) (default 30)
```

### Usage Examples
```bash
# View a graph of the last 30 commits
gitgraphy graph

# View a graph of the last 100 commits
gitgraphy graph -n 100

# View a graph with all commits
gitgraphy graph -n 0
``` 

## show 
The `gitgraphy show` command allows you to inspect a single commit, viewing extended information about it, including the `diff`. Like `gitgraphy compare`, you can pass the `-s` or `-stats` flag to only show the summary, without the diff. 

### Usage
```bash
Inspect a commit with summary and diff

Usage:
  gitgraphy show [hash] [flags]

Flags:
  -h, --help   help for show
  -s, --stat   Show file summary only without code diffs
```

### Usage Examples
```bash
# Inspect the HEAD commit 
gitgraphy show

# Inspect a specific commit (can also be shortened hash)
gitgraphy show 876ce9c59db8510fa0d767fd1d62a725a1eb2531

# Inspect HEAD with summary only
gitgraphy show -s
```

## status
The `gitgraphy status` command provides a high level status overview of the **current working tree**, letting you see ahead / behind counts and unstaged, staged, modified, and untracked files. This command has no flags

### Usage
```bash
Show a clean visual of the working tree

Usage:
  gitgraphy status [flags]

Flags:
  -h, --help   help for status
```

### Usage Examples
```bash
# View the current working tree
gitgraphy status
```