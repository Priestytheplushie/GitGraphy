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
The `gitgraphy blame` command allows you to see which `Commit` and `Author` modified that speicfic line of a file 