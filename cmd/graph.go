package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Priestytheplushie/GitGraphy/git"
)

var graphLimit int

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Visualize commit history as a multi-branch visual graph",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.Open(".")
		if err != nil {
			return fmt.Errorf("not a git repository (or any parent directory)")
		}

		commits, err := repo.GetAllCommits(graphLimit)
		if err != nil {
			return fmt.Errorf("failed to read commit graph: %w", err)
		}

		if len(commits) == 0 {
			fmt.Println("No commits found.")
			return nil
		}

		renderGraph(commits)
		return nil
	},
}

func init() {
	graphCmd.Flags().IntVarP(&graphLimit, "limit", "n", 30, "Number of commits to show")
	rootCmd.AddCommand(graphCmd)
}

var laneColors = []*color.Color{
	color.New(color.FgCyan, color.Bold),
	color.New(color.FgYellow, color.Bold),
	color.New(color.FgHiMagenta, color.Bold),
	color.New(color.FgHiGreen, color.Bold),
	color.New(color.FgHiBlue, color.Bold),
	color.New(color.FgHiRed, color.Bold),
}

func getLaneColor(lane int) func(a ...interface{}) string {
	return laneColors[lane%len(laneColors)].SprintFunc()
}

func renderGraph(commits []git.Commit) {
	hashColor := color.New(color.FgYellow).SprintFunc()
	headColor := color.New(color.FgCyan, color.Bold).SprintFunc()
	branchColor := color.New(color.FgGreen).SprintFunc()
	remoteColor := color.New(color.FgRed).SprintFunc()
	tagColor := color.New(color.FgMagenta).SprintFunc()
	authorColor := color.New(color.FgBlue).SprintFunc()
	timeColor := color.New(color.Faint).SprintFunc()
	bulletColor := color.New(color.Faint).SprintFunc()

	lanes := []string{}
	maxLanesEncountered := 1

	for _, c := range commits {
		converging := []int{}
		for i, h := range lanes {
			if h == c.Hash {
				converging = append(converging, i)
			}
		}

		if len(converging) > 1 {
			target := converging[0]
			source := converging[1]

			line := renderBranchConnector(len(lanes), target, source, "merge") + padGutter(len(lanes), maxLanesEncountered)
			fmt.Println(line)

			lanes[source] = ""
			for len(lanes) > 0 && lanes[len(lanes)-1] == "" {
				lanes = lanes[:len(lanes)-1]
			}
		}

		laneIdx := -1
		for i, h := range lanes {
			if h == c.Hash {
				laneIdx = i
				break
			}
		}

		if laneIdx == -1 {
			for i, h := range lanes {
				if h == "" {
					laneIdx = i
					break
				}
			}
			if laneIdx == -1 {
				lanes = append(lanes, c.Hash)
				laneIdx = len(lanes) - 1
			} else {
				lanes[laneIdx] = c.Hash
			}
		}

		if len(lanes) > maxLanesEncountered {
			maxLanesEncountered = len(lanes)
		}

		var graphPrefix strings.Builder
		for i := 0; i < len(lanes); i++ {
			cFunc := getLaneColor(i)
			if i == laneIdx {
				graphPrefix.WriteString(cFunc("●") + " ")
			} else if lanes[i] != "" {
				graphPrefix.WriteString(cFunc("│") + " ")
			} else {
				graphPrefix.WriteString("  ")
			}
		}
		graphPrefix.WriteString(padGutter(len(lanes), maxLanesEncountered))

		isMerge := len(c.ParentHashes) > 1
		p2Lane := -1

		if len(c.ParentHashes) == 0 {
			lanes[laneIdx] = ""
		} else {
			lanes[laneIdx] = c.ParentHashes[0]
			if isMerge {
				parent2 := c.ParentHashes[1]
				for i, h := range lanes {
					if h == "" {
						p2Lane = i
						break
					}
				}
				if p2Lane == -1 {
					lanes = append(lanes, parent2)
					p2Lane = len(lanes) - 1
				} else {
					lanes[p2Lane] = parent2
				}
				if len(lanes) > maxLanesEncountered {
					maxLanesEncountered = len(lanes)
				}
			}
		}

		var refDecorations []string
		for _, ref := range c.Refs {
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
			refString = fmt.Sprintf("(%s) ", strings.Join(refDecorations, ", "))
		}

		subject := c.Subject
		if len(subject) > 42 {
			subject = subject[:39] + "..."
		}

		relTime := formatRelativeTime(c.When)

		fmt.Printf(
			"%s%s %s%s  %s  %s  %s\n",
			graphPrefix.String(),
			hashColor(c.ShortHash),
			refString,
			subject,
			bulletColor("•"),
			authorColor(c.AuthorName),
			timeColor("("+relTime+")"),
		)

		if isMerge && p2Lane != -1 {
			line := renderBranchConnector(len(lanes), laneIdx, p2Lane, "fork") + padGutter(len(lanes), maxLanesEncountered)
			fmt.Println(line)
		}
	}
}

func renderBranchConnector(totalLanes, trunkLane, branchLane int, kind string) string {
	var sb strings.Builder

	left := trunkLane
	right := branchLane
	if left > right {
		left, right = right, left
	}

	trunkColor := getLaneColor(trunkLane)
	branchColor := getLaneColor(branchLane)

	for i := 0; i < totalLanes; i++ {
		if i == left {
			if kind == "fork" {
				sb.WriteString(trunkColor("├") + branchColor("─╮") + " ")
			} else {
				sb.WriteString(trunkColor("├") + branchColor("─╯") + " ")
			}
			i = right
		} else if i < right {
			sb.WriteString(getLaneColor(i)("│") + " ")
		}
	}
	return sb.String()
}

func padGutter(currentLanes, maxLanes int) string {
	pad := ""
	if currentLanes < maxLanes {
		pad = strings.Repeat("  ", maxLanes-currentLanes)
	}
	return pad + "  "
}