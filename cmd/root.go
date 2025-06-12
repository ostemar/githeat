package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	days    int
	rootCmd = &cobra.Command{
		Use:   "githeat",
		Short: "GitHeat is a tool to visualize git repository activity",
		RunE: func(cmd *cobra.Command, args []string) error {
			var repoPath string
			if len(args) > 0 {
				repoPath = args[0]
			} else {
				repoPath, _ = cmd.Flags().GetString("repo")
			}
			if repoPath == "" {
				fmt.Println("Error: repository path is required")
				os.Exit(1)
			}

			sinceDate := time.Now().AddDate(0, 0, -days)

			commitMap, err := getCommitsByDate(repoPath, sinceDate)
			if err != nil {
				return err
			}

			// Prepare a map for quick lookup
			commitCountByDate := make(map[string]int)
			weekdayTotals := make([]int, 7)
			for date, commits := range commitMap {
				commitCountByDate[date] = len(commits)
				parsedDate, err := time.Parse("2006-01-02", date)
				if err == nil {
					wd := int(parsedDate.Weekday())
					if wd == 0 {
						wd = 6 // Sunday as 6
					} else {
						wd = wd - 1 // Monday as 0
					}
					weekdayTotals[wd] += len(commits)
				}
			}

			monday := getMondaysDateForDate(sinceDate)
			today := time.Now()
			weekCount := int(today.Sub(monday).Hours()/24/7) + 1

			fmt.Println()

			// Print month labels
			monthLabels := make([]string, weekCount)
			monthPrinted := make(map[string]bool)
			for w := range weekCount {
				weekStart := monday.AddDate(0, 0, w*7)
				monthKey := weekStart.Format("2006-01")
				if weekStart.Day() <= 7 && !monthPrinted[monthKey] {
					monthLabels[w] = weekStart.Format("Jan")
					monthPrinted[monthKey] = true
				}
			}
			fmt.Printf("    ")
			for w := 0; w < weekCount; {
				if monthLabels[w] != "" {
					fmt.Printf("%4s", monthLabels[w])
					w += 2
				} else {
					fmt.Printf("  ")
					w++
				}
			}
			fmt.Println()

			// Prepare matrix: rows are days (Mon-Sun), columns are weeks
			matrix := make([][]string, 7)
			for i := range 7 {
				matrix[i] = make([]string, weekCount)
			}

			for w := range weekCount {
				for d := range 7 {
					current := monday.AddDate(0, 0, w*7+d)
					if current.After(today) {
						matrix[d][w] = " "
						continue
					}
					dateStr := current.Format("2006-01-02")
					if count, ok := commitCountByDate[dateStr]; ok && count > 0 {
						matrix[d][w] = colorize(count, false)
					} else {
						matrix[d][w] = " "
					}
				}
			}

			dayNames := []string{"Mon", "", "Wed", "", "Fri", "", "Sun"}
			for i, day := range dayNames {
				fmt.Printf("%4s ", day)
				for w := range weekCount {
					fmt.Printf("%s ", matrix[i][w])
				}
				fmt.Printf("   %2d %s", weekdayTotals[i], day)
				fmt.Println()
			}

			return nil
		},
	}
)

func getCommitsByDate(repoPath string, sinceDate time.Time) (map[string][]struct {
	SHA  string
	Tags string
}, error,
) {
	gitcmd := exec.Command("git", "log",
		"--since="+sinceDate.Format("2006-01-02"),
		"--pretty=format:%ad %h %d",
		"--date=short",
		"--first-parent")
	gitcmd.Dir = repoPath
	output, err := gitcmd.Output()
	if err != nil {
		fmt.Printf("Error executing git command: %v\n", err)
		return nil, err
	}

	commitMap := make(map[string][]struct {
		SHA  string
		Tags string
	})
	lines := strings.SplitSeq(string(output), "\n")
	for line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		date := parts[0]
		sha := parts[1]
		tags := ""
		if len(parts) > 2 {
			tags = strings.TrimSpace(parts[2])
		}
		commitMap[date] = append(commitMap[date], struct {
			SHA  string
			Tags string
		}{
			SHA:  sha,
			Tags: tags,
		})
	}
	return commitMap, nil
}

func getMondaysDateForDate(date time.Time) time.Time {
	// startDate, _ := time.Parse("2006-01-02", date)
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday as 7
	}
	monday := date.AddDate(0, 0, -weekday+1)
	return monday
}

func colorize(count int, background bool) string {
	if background {
		switch {
		case count >= 9:
			return "\033[30;48;5;22m9\033[0m" // Dark Green
		case count == 8:
			return "\033[30;48;5;28m8\033[0m" // Another shade of Dark Green
		case count == 7:
			return "\033[30;48;5;34m7\033[0m" // More Green
		case count == 6:
			return "\033[30;48;5;40m6\033[0m" // Medium Green
		case count == 5:
			return "\033[30;48;5;46m5\033[0m" // Lighter Green
		case count == 4:
			return "\033[30;48;5;82m4\033[0m" // Lighter Green
		case count == 3:
			return "\033[30;48;5;118m3\033[0m" // Even Lighter Green
		case count == 2:
			return "\033[30;48;5;154m2\033[0m" // Light Green
		case count == 1:
			return "\033[30;48;5;191m1\033[0m" // Very Light Green
		default:
			return "-" // No commits
		}
	} else {
		switch {
		case count >= 9:
			return "\033[38;5;191m9\033[0m" // Very Light Green
		case count == 8:
			return "\033[38;5;154m8\033[0m" // Light Green
		case count == 7:
			return "\033[38;5;118m7\033[0m" // Even Lighter Green
		case count == 6:
			return "\033[38;5;82m6\033[0m" // Lighter Green
		case count == 5:
			return "\033[38;5;46m5\033[0m" // Lighter Green
		case count == 4:
			return "\033[38;5;40m4\033[0m" // Medium Green
		case count == 3:
			return "\033[38;5;34m3\033[0m" // More Green
		case count == 2:
			return "\033[38;5;28m2\033[0m" // Another shade of Dark Green
		case count == 1:
			return "\033[38;5;22m1\033[0m" // Dark Green
		default:
			return "-" // No commits
		}
	}
}

func init() {
	rootCmd.Flags().StringP("repo", "r", "", "Path to the git repository")
	rootCmd.Flags().IntVarP(&days, "days", "d", 365, "Number of days to look back for commits")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
