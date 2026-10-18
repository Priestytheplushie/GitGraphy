package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Priestytheplushie/GitGraphy/git"
)

var compareStatOnly bool

var compareCmd = &cobra.Command{
	Use:   "compare [base] <target>",
	Short: "Compare branches or revisions like a GitHub PR preview",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.Open(".")
		if err != nil {
			return fmt.Errorf("not a git repository (or any parent directory)")
		}

		base := ""
		target := ""

		if len(args) == 1 {
			if strings.Contains(args[0], "...") {
				parts := strings.Split(args[0], "...")
				base = parts[0]
				target = parts[1]
			} else if strings.Contains(args[0], "..") {
				parts := strings.Split(args[0], "..")
				base = parts[0]
				target = parts[1]
			} else {
				target = args[0]
			}
		} else {
			base = args[0]
			target = args[1]
		}

		res, err := repo.Compare(base, target)
		if err != nil {
			return err
		}

		renderCompare(res, compareStatOnly)
		return nil
	},
}

func init() {
	compareCmd.Flags().BoolVarP(&compareStatOnly, "stat", "s", false, "Show file impact summary only without code diffs")
	rootCmd.AddCommand(compareCmd)
}

func renderCompare(c *git.CompareResult, statOnly bool) {
	cyanBold := color.New(color.FgCyan, color.Bold).SprintFunc()
	faint := color.New(color.Faint).SprintFunc()
	addBold := color.New(color.FgHiGreen, color.Bold).SprintFunc()
	delBold := color.New(color.FgHiRed, color.Bold).SprintFunc()
	nodeColor := color.New(color.FgHiMagenta).SprintFunc()
	hashColor := color.New(color.FgYellow).SprintFunc()
	authorColor := color.New(color.FgBlue).SprintFunc()

	fmt.Printf(
		"Comparing: %s...%s\n",
		cyanBold(c.BaseRef),
		cyanBold(c.TargetRef),
	)
	fmt.Printf(
		"Commits Ahead: %s  •  Commits Behind: %s\n\n",
		color.New(color.FgHiGreen, color.Bold).Sprint(c.AheadCount),
		color.New(color.FgHiRed, color.Bold).Sprint(c.BehindCount),
	)

	if len(c.CommitsAhead) > 0 {
		fmt.Printf("%s\n", faint(fmt.Sprintf("Commits (%d)", len(c.CommitsAhead))))
		for _, cm := range c.CommitsAhead {
			fmt.Printf(
				"  %s %s %s  %s  %s  %s\n",
				nodeColor("●"),
				hashColor(cm.ShortHash),
				cm.Subject,
				faint("•"),
				authorColor(cm.AuthorName),
				faint("("+formatRelativeTime(cm.When)+")"),
			)
		}
		fmt.Println()
	}

	fmt.Printf(
		"%s • %s %s\n",
		faint(fmt.Sprintf("Files Changed (%d)", len(c.Stats))),
		addBold(fmt.Sprintf("+%d", c.TotalAdd)),
		delBold(fmt.Sprintf("-%d", c.TotalDel)),
	)

	maxNameLen := 0
	for _, s := range c.Stats {
		if len(s.Name) > maxNameLen {
			maxNameLen = len(s.Name)
		}
	}

	for _, s := range c.Stats {
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

	if statOnly || len(c.PatchText) == 0 {
		return
	}

	renderDiff(c.PatchText)
}