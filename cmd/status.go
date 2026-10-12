package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Priestytheplushie/GitGraphy/git"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show a clean visual of the working tree",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.Open(".")
		if err != nil {
			return fmt.Errorf("not a git repository (or any parent directory)")
		}

		res, err := repo.GetStatus()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		renderStatus(res)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func renderStatus(s *git.StatusResult) {
	cyanBold := color.New(color.FgCyan, color.Bold).SprintFunc()
	greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()
	yellowBold := color.New(color.FgYellow, color.Bold).SprintFunc()
	faint := color.New(color.Faint).SprintFunc()

	var syncBadge string
	if !s.HasUpstream {
		syncBadge = faint("[no upstream]")
	} else if s.Ahead == 0 && s.Behind == 0 {
		syncBadge = greenBold("[✔ in sync with ") + cyanBold(s.UpstreamName) + greenBold("]")
	} else {
		var parts string
		if s.Ahead > 0 {
			parts += color.New(color.FgHiGreen, color.Bold).Sprintf("↑ %d ", s.Ahead)
		}
		if s.Behind > 0 {
			parts += color.New(color.FgHiRed, color.Bold).Sprintf("↓ %d ", s.Behind)
		}
		syncBadge = fmt.Sprintf("[%s] %s", parts, faint(fmt.Sprintf("(ahead %d, behind %d of %s)", s.Ahead, s.Behind, s.UpstreamName)))
	}

	fmt.Printf("Branch: %s %s\n\n", cyanBold(s.Branch), syncBadge)

	if s.IsClean() {
		fmt.Println(greenBold("✔ Working tree clean — nothing to commit"))
		return
	}

	if len(s.Staged) > 0 {
		fmt.Printf("%s %s\n", greenBold("Staged Changes"), faint(fmt.Sprintf("(%d)", len(s.Staged))))
		for _, f := range s.Staged {
			fmt.Printf("  %s  %s\n", renderBadge(f.Letter), f.Path)
		}
		fmt.Println()
	}

	if len(s.Unstaged) > 0 {
		fmt.Printf("%s %s\n", yellowBold("Unstaged Changes"), faint(fmt.Sprintf("(%d)", len(s.Unstaged))))
		for _, f := range s.Unstaged {
			fmt.Printf("  %s  %s\n", renderBadge(f.Letter), f.Path)
		}
		fmt.Println()
	}

	if len(s.Untracked) > 0 {
		fmt.Printf("%s %s\n", color.New(color.FgHiGreen).Sprint("Untracked Files"), faint(fmt.Sprintf("(%d)", len(s.Untracked))))
		for _, f := range s.Untracked {
			fmt.Printf("  %s  %s\n", renderBadge(f.Letter), f.Path)
		}
		fmt.Println()
	}

	fmt.Println(faint(fmt.Sprintf(
		"Summary: %d staged, %d unstaged, %d untracked",
		len(s.Staged), len(s.Unstaged), len(s.Untracked),
	)))
}

func renderBadge(letter string) string {
	switch letter {
	case "M":
		return color.New(color.FgYellow, color.Bold).Sprint("M")
	case "A":
		return color.New(color.FgGreen, color.Bold).Sprint("A")
	case "D":
		return color.New(color.FgRed, color.Bold).Sprint("D")
	case "U":
		return color.New(color.FgHiGreen, color.Bold).Sprint("U")
	case "R":
		return color.New(color.FgCyan, color.Bold).Sprint("R")
	default:
		return color.New(color.FgWhite).Sprint(letter)
	}
}