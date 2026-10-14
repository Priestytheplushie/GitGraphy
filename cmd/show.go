package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Priestytheplushie/GitGraphy/git"
)

var statOnly bool

var showCmd = &cobra.Command{
	Use:   "show [hash]",
	Short: "Inspect a commit with summary and diff",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.Open(".")
		if err != nil {
			return fmt.Errorf("not a git repository (or any parent directory)")
		}

		rev := ""
		if len(args) > 0 {
			rev = args[0]
		}

		details, err := repo.GetCommitDetails(rev)
		if err != nil {
			return err
		}

		renderShow(details, statOnly)
		return nil
	},
}

func init() {
	showCmd.Flags().BoolVarP(&statOnly, "stat", "s", false, "Show file summary only without code diffs")
	rootCmd.AddCommand(showCmd)
}

func renderShow(d *git.CommitDetails, statOnly bool) {
	hashColor := color.New(color.FgYellow, color.Bold).SprintFunc()
	authorColor := color.New(color.FgBlue).SprintFunc()
	faint := color.New(color.Faint).SprintFunc()
	addBold := color.New(color.FgHiGreen, color.Bold).SprintFunc()
	delBold := color.New(color.FgHiRed, color.Bold).SprintFunc()
	headColor := color.New(color.FgCyan, color.Bold).SprintFunc()
	branchColor := color.New(color.FgGreen).SprintFunc()
	remoteColor := color.New(color.FgRed).SprintFunc()
	tagColor := color.New(color.FgMagenta).SprintFunc()

	var refDecorations []string
	for _, ref := range d.Commit.Refs {
		switch {
		case ref.IsHEAD:
			refDecorations = append(refDecorations, headColor("HEAD -> ")+branchColor(ref.Name))
		case ref.IsRemote:
			refDecorations = append(refDecorations, remoteColor(ref.Name))
		case ref.IsTag:
			refDecorations = append(refDecorations, tagColor("tag: "+ref.Name))
		default:
			refDecorations = append(refDecorations, branchColor(ref.Name))
		}
	}

	refString := ""
	if len(refDecorations) > 0 {
		refString = fmt.Sprintf(" (%s)", strings.Join(refDecorations, ", "))
	}

	fmt.Printf("commit %s%s\n", hashColor(d.Commit.Hash), refString)
	fmt.Printf("Author: %s %s\n", authorColor(d.Commit.AuthorName), faint("<"+d.Commit.AuthorEmail+">"))
	fmt.Printf("Date:   %s %s\n\n", d.Commit.When.Format("Mon Jan 02 15:04:05 2006 -0700"), faint("("+formatRelativeTime(d.Commit.When)+")"))

	for _, line := range strings.Split(d.Message, "\n") {
		fmt.Printf("    %s\n", line)
	}
	fmt.Println()

	fmt.Printf(
		"%s • %s %s\n",
		faint(fmt.Sprintf("Files Changed (%d)", len(d.Stats))),
		addBold(fmt.Sprintf("+%d", d.TotalAdd)),
		delBold(fmt.Sprintf("-%d", d.TotalDel)),
	)

	maxNameLen := 0
	for _, s := range d.Stats {
		if len(s.Name) > maxNameLen {
			maxNameLen = len(s.Name)
		}
	}

	for _, s := range d.Stats {
		namePadding := strings.Repeat(" ", maxNameLen-len(s.Name)+2)
		counts := fmt.Sprintf("+%-3d -%-3d", s.Additions, s.Deletions)
		bar := renderImpactBar(s.Additions, s.Deletions)

		note := ""
		if s.IsLock {
			note = faint(" [lockfile diff collapsed]")
		} else if s.IsBinary {
			note = faint(" [binary file]")
		}

		fmt.Printf("  %s%s%s  %s%s\n", s.Name, namePadding, counts, bar, note)
	}
	fmt.Println()

	if statOnly || len(d.PatchText) == 0 {
		return
	}

	renderDiff(d.PatchText)
}

func renderImpactBar(adds, dels int) string {
	total := adds + dels
	if total == 0 {
		return color.New(color.Faint).Sprint("░░░░░")
	}

	const maxBlocks = 5
	addBlocks := int(float64(adds) / float64(total) * maxBlocks)
	delBlocks := int(float64(dels) / float64(total) * maxBlocks)

	if adds > 0 && addBlocks == 0 {
		addBlocks = 1
	}
	if dels > 0 && delBlocks == 0 {
		delBlocks = 1
	}
	for addBlocks+delBlocks > maxBlocks {
		if addBlocks > delBlocks {
			addBlocks--
		} else {
			delBlocks--
		}
	}
	emptyBlocks := maxBlocks - (addBlocks + delBlocks)

	green := color.New(color.FgHiGreen).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	faint := color.New(color.Faint).SprintFunc()

	return green(strings.Repeat("█", addBlocks)) +
		red(strings.Repeat("█", delBlocks)) +
		faint(strings.Repeat("░", emptyBlocks))
}

func renderDiff(patch string) {
	lines := strings.Split(patch, "\n")
	hunkColor := color.New(color.FgCyan, color.Faint).SprintFunc()
	addColor := color.New(color.FgGreen).SprintFunc()
	delColor := color.New(color.FgRed).SprintFunc()
	faint := color.New(color.Faint).SprintFunc()
	border := color.New(color.Faint).Sprint(strings.Repeat("─", 65))

	skippingLockfile := false

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			fileName := ""
			if len(parts) >= 4 {
				fileName = strings.TrimPrefix(parts[3], "b/")
			}

			if isLockfile(fileName) {
				skippingLockfile = true
				continue
			}
			skippingLockfile = false

			fmt.Println(border)
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint(fileName), faint("(diff)"))
			fmt.Println(border)
			continue
		}

		if skippingLockfile {
			continue
		}

		if strings.HasPrefix(line, `\ No newline`) {
			continue
		}

		if strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			continue
		}

		switch {
		case strings.HasPrefix(line, "@@"):
			fmt.Println(hunkColor(line))
		case strings.HasPrefix(line, "+"):
			fmt.Println(addColor(line))
		case strings.HasPrefix(line, "-"):
			fmt.Println(delColor(line))
		default:
			fmt.Println(line)
		}
	}
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
