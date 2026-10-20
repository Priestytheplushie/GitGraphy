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
	Short: "View GitLens-style visual blame with block headers and syntax highlighting",
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

	lineNumWidth := len(fmt.Sprintf("%d", b.TotalLines))
	if lineNumWidth < 3 {
		lineNumWidth = 3
	}

	highlightedLines := highlightCode(b.Path, b.Lines)

	lastHash := ""
	const targetBannerLen = 78

	for i, l := range b.Lines {
		if l.Hash != lastHash {
			lastHash = l.Hash

			if i > 0 {
				fmt.Println()
			}

			colorFunc := getAgeColor(l.Date)
			hashStr := colorFunc(l.ShortHash)
			authorStr := color.New(color.FgHiBlue, color.Bold).Sprint(l.Author)
			timeStr := faint("(" + formatRelativeTime(l.Date) + ")")
			subjStr := faint("\"" + l.Subject + "\"")

			prefix := fmt.Sprintf("── %s • %s %s • %s ", hashStr, authorStr, timeStr, subjStr)
			visibleLen := 3 + len(l.ShortHash) + 3 + len(l.Author) + 1 + len(formatRelativeTime(l.Date)) + 2 + 3 + len(l.Subject) + 2 + 1

			padLen := targetBannerLen - visibleLen
			if padLen < 4 {
				padLen = 4
			}

			rule := faint(strings.Repeat("─", padLen))
			fmt.Printf("%s%s\n", prefix, rule)
		}

		code := ""
		if i < len(highlightedLines) {
			code = highlightedLines[i]
		}
		code = strings.ReplaceAll(code, "\t", "    ")

		fmt.Printf(
			"%*d %s %s\n",
			lineNumWidth,
			l.LineNo,
			sep,
			code,
		)
	}
}

func highlightCode(path string, lines []git.BlameLine) []string {
	var fullCode strings.Builder
	for _, l := range lines {
		fullCode.WriteString(strings.TrimRight(l.Code, "\r\n"));fullCode.WriteString("\n")
	}

	lexer := lexers.Match(path)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("github-dark")
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
