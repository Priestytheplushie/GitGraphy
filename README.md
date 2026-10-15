# GitGraphy 

A CLI for visualizing git repositories. 

## QuickStart
Build and install GitGraphy locally using `Go`: 

### Prerequisites 
* **[Go](https://go.dev/)**

### Installation 
Install GitGraphy by cloning the repository and building from source:
```bash
# Clone Repository
git clone https://github.com/Priestytheplushie/GitGraphy
cd GitGraphy

# Build 
go build -o gitgraphy .

# Install
go install .

# Run GitGraphy
gitgraphy
```

## Usage
You can use the `gitgraphy log` command to get a color coded compact log of your repository history. Passing the `-n` or `--limit` flag will limit how many commits are displayed. You can pass `-n 0` to have no limit 
```bash
gitgraphy log -n 30
```
You can use `gitgraphy graph` to create a visual graph represneation of your repository with color coded branches, passing the `-n` flag or `--limit` will limit how many commits are displayed. You can pass `-n 0` to have no limit 
```bash
gitgraphy graph -n 100
```
You can use `gitgraphy status` to get a visual overview of the current working tree, seeing ahead/behind status of the branch as well as current staged, unstaged, and untracked files
```bash
gitgraphy log
```
You can use `gitgraphy show` to view a commit and it's `diff`. You can pass in the `-s` or `-stats` flag to get a summary without a diff, or pass in a commit hash to inspect a specific commit
```bash 
# Check latest commit
gitgraphy show 

# Check specific commit with summary
gitgraphy show 7bb0e6d -s
```
## Documentation 
Documentation is **coming soon**! Stay tuned

## License
This project is licensed under the GPLv3 license, for more info, check the [LICENSE](LICENSE) file. 