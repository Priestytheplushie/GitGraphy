package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Priestytheplushie/GitGraphy/git"
)

var limit int

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show a clean, decorated commit history",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.Open(".")
		if err != nil {
			return fmt.Errorf("not a git repository (or any parent directory)")
		}

		commits, err := repo.GetCommits(limit)
		if err != nil {
			return fmt.Errorf("failed to read commits: %w", err)
		}

		if len(commits) == 0 {
			fmt.Println("No commits found.")
			return nil
		}

		renderLog(commits)
		return nil
	},
}

func init() {
	logCmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of commits to show")
	rootCmd.AddCommand(logCmd)
}

type rowData struct {
	nodeText      string
	hashText      string
	refText       string
	subjectText   string
	authorText    string
	timeText      string
	visibleMsgLen int
	authorLen     int
}

func renderLog(commits []git.Commit) {
	hashColor := color.New(color.FgYellow).SprintFunc()
	headColor := color.New(color.FgCyan, color.Bold).SprintFunc()
	branchColor := color.New(color.FgGreen).SprintFunc()
	remoteColor := color.New(color.FgRed).SprintFunc()
	tagColor := color.New(color.FgMagenta).SprintFunc()
	authorColor := color.New(color.FgBlue).SprintFunc()
	timeColor := color.New(color.Faint).SprintFunc()
	nodeColor := color.New(color.FgHiMagenta).SprintFunc()
	bulletColor := color.New(color.Faint).SprintFunc()

	const maxSubjectLen = 50
	maxContentCol := 40
	maxAuthorCol := 0

	rows := make([]rowData, len(commits))

	for i, c := range commits {
		var coloredRefs []string
		var plainRefs []string

		for _, ref := range c.Refs {
			switch {
			case ref.IsHEAD:
				coloredRefs = append(coloredRefs, headColor("HEAD -> ")+branchColor(ref.Name))
				plainRefs = append(plainRefs, "HEAD -> "+ref.Name)
			case ref.IsRemote:
				coloredRefs = append(coloredRefs, remoteColor(ref.Name))
				plainRefs = append(plainRefs, ref.Name)
			case ref.IsTag:
				coloredRefs = append(coloredRefs, tagColor("tag: "+ref.Name))
				plainRefs = append(plainRefs, "tag: "+ref.Name)
			default:
				coloredRefs = append(coloredRefs, branchColor(ref.Name))
				plainRefs = append(plainRefs, ref.Name)
			}
		}

		refString := ""
		plainRefString := ""
		if len(coloredRefs) > 0 {
			refString = fmt.Sprintf("(%s) ", strings.Join(coloredRefs, ", "))
			plainRefString = fmt.Sprintf("(%s) ", strings.Join(plainRefs, ", "))
		}

		subject := c.Subject
		if len(subject) > maxSubjectLen {
			subject = subject[:maxSubjectLen-3] + "..."
		}

		visibleMsgLen := len(plainRefString) + len(subject)
		if visibleMsgLen > maxContentCol {
			maxContentCol = visibleMsgLen
		}

		authorLen := len(c.AuthorName)
		if authorLen > maxAuthorCol {
			maxAuthorCol = authorLen
		}

		rows[i] = rowData{
			nodeText:      nodeColor("●"),
			hashText:      hashColor(c.ShortHash),
			refText:       refString,
			subjectText:   subject,
			authorText:    authorColor(c.AuthorName),
			timeText:      timeColor("(" + formatRelativeTime(c.When) + ")"),
			visibleMsgLen: visibleMsgLen,
			authorLen:     authorLen,
		}
	}

	for _, row := range rows {
		msgPadding := strings.Repeat(" ", maxContentCol-row.visibleMsgLen)
		authorPadding := strings.Repeat(" ", maxAuthorCol-row.authorLen)

		fmt.Printf(
			"%s %s %s%s%s  %s  %s%s  %s\n",
			row.nodeText,
			row.hashText,
			row.refText,
			row.subjectText,
			msgPadding,
			bulletColor("•"),
			row.authorText,
			authorPadding,
			row.timeText,
		)
	}
}

func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case duration < 30*24*time.Hour:
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 02, 2006")
	}
}