package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Priestytheplushie/GitGraphy/git"
)

var blameCmd = &cobra.Command{
	Use:   "blame <file>",
	Short: "View GitLens-style visual blame with syntax highlighting and commit grouping",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.Open(".")
		if err != nil {
			return fmt.Errorf("not a git repository (or any parent directory)")
		}

		res, err := repo.GetBlame(args[0])
		if err != nil {
			return err
		}

		renderBlame(res)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blameCmd)
}

func renderBlame(b *git.BlameResult) {
	faint := color.New(color.Faint).SprintFunc()
	cyanBold := color.New(color.FgCyan, color.Bold).SprintFunc()
	sep := faint("│")

	fmt.Printf(
		"Blame: %s • %s • %s • %s\n\n",
		cyanBold(b.Path),
		faint(fmt.Sprintf("%d lines", b.TotalLines)),
		faint(fmt.Sprintf("%d commits", b.CommitCount)),
		faint(fmt.Sprintf("%d contributors", len(b.Authors))),
	)

	maxAuthorLen := 0
	for _, l := range b.Lines {
		if len(l.Author) > maxAuthorLen {
			maxAuthorLen = len(l.Author)
		}
	}
	if maxAuthorLen > 20 {
		maxAuthorLen = 20
	}
	if maxAuthorLen < 6 {
		maxAuthorLen = 6
	}

	lineNumWidth := len(fmt.Sprintf("%d", b.TotalLines))
	if lineNumWidth < 3 {
		lineNumWidth = 3
	}

	highlightedLines := highlightCode(b.Path, b.Lines)

	lastHash := ""
	const timeWidth = 9
	const subjectWidth = 26

	gutterWidth := 7 + 1 + maxAuthorLen + 1 + timeWidth + 1 + subjectWidth

	for i, l := range b.Lines {
		isNewBlock := l.Hash != lastHash
		lastHash = l.Hash

		colorFunc := getAgeColor(l.Date)

		var gutter string
		if isNewBlock {
			author := l.Author
			if len(author) > maxAuthorLen {
				author = author[:maxAuthorLen-3] + "..."
			}
			padLen := maxAuthorLen - len(author)
			if padLen < 0 {
				padLen = 0
			}
			authorPadded := author + strings.Repeat(" ", padLen)

			relTime := formatRelativeTime(l.Date)
			if len(relTime) > timeWidth {
				relTime = relTime[:timeWidth]
			}
			timePadLen := timeWidth - len(relTime)
			if timePadLen < 0 {
				timePadLen = 0
			}
			timePadded := relTime + strings.Repeat(" ", timePadLen)

			subj := l.Subject
			if len(subj) > subjectWidth {
				subj = subj[:subjectWidth-3] + "..."
			}
			subjPadLen := subjectWidth - len(subj)
			if subjPadLen < 0 {
				subjPadLen = 0
			}
			subjPadded := subj + strings.Repeat(" ", subjPadLen)

			gutter = fmt.Sprintf(
				"%s %s %s %s",
				colorFunc(l.ShortHash),
				color.New(color.FgHiBlue).Sprint(authorPadded),
				faint(timePadded),
				faint(subjPadded),
			)
		} else {
			gutter = strings.Repeat(" ", gutterWidth)
		}

		code := ""
		if i < len(highlightedLines) {
			code = highlightedLines[i]
		}

		fmt.Printf(
			"%*d %s %s %s %s\n",
			lineNumWidth,
			l.LineNo,
			sep,
			gutter,
			sep,
			code,
		)
	}
}

func highlightCode(path string, lines []git.BlameLine) []string {
	var fullCode strings.Builder
	for _, l := range lines {
		fullCode.WriteString(strings.TrimRight(l.Code, "\r\n") + "\n")
	}

	lexer := lexers.Match(path)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, fullCode.String())
	if err != nil {
		res := make([]string, len(lines))
		for i, l := range lines {
			res[i] = strings.TrimRight(l.Code, "\r\n")
		}
		return res
	}

	var buf bytes.Buffer
	_ = formatter.Format(&buf, style, iterator)

	highlighted := strings.Split(buf.String(), "\n")
	res := make([]string, len(lines))
	for i := range lines {
		if i < len(highlighted) {
			res[i] = strings.TrimRight(highlighted[i], "\r\n")
		} else {
			res[i] = strings.TrimRight(lines[i].Code, "\r\n")
		}
	}
	return res
}

func getAgeColor(t time.Time) func(a ...interface{}) string {
	d := time.Since(t)
	switch {
	case d < 24*time.Hour:
		return color.New(color.FgHiGreen, color.Bold).SprintFunc()
	case d < 7*24*time.Hour:
		return color.New(color.FgYellow, color.Bold).SprintFunc()
	case d < 30*24*time.Hour:
		return color.New(color.FgCyan, color.Bold).SprintFunc()
	default:
		return color.New(color.Faint).SprintFunc()
	}
}