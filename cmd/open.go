package cmd

import (
	"fmt"
	"lin_cli/internal/git"
	"lin_cli/internal/linear"
	"lin_cli/internal/util"
	"log"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(openCmd)
}

var openCmd = &cobra.Command{
	Use: "open [<issue identifier>]",
	Short: "Opens the URL to the provided issue. If no issue is provided, opens the " +
		"issue attached to the current working branch",
	Run: func(cmd *cobra.Command, args []string) {
		org, needRefresh, err := linear.LoadOrg()
		if err != nil {
			log.Fatalf("Failed to read cache file: %s", err)
		}
		if needRefresh {
			org, err = linear.GetOrganization(linear.GetClient())
			if err != nil {
				log.Fatalf("Failed to fetch organization: %s", err)
			}
		}

		issues, needRefresh, err := linear.LoadIssues(linear.GetClient())
		if err != nil {
			log.Fatalf("Failed to read cache file: %s", err)
		}
		if needRefresh {
			issues, err = linear.GetIssues(linear.GetClient())
			if err != nil {
				log.Fatalf("Failed to fetch issues: %s", err)
			}
		}

		var teamKeys []string

		for _, team := range org.Teams {
			teamKeys = append(teamKeys, team.Key)
		}

		var issueId string
		if len(args) > 0 {
			issueId = strings.Trim(args[0], " ")
		}

		// Fetch issue identifier
		if issueId == "" {
			branchName := git.GetCurrentBranch()
			parseBranchNameForIssue(branchName, teamKeys)
			if branchName == "main" || branchName == "master" {
				fmt.Println("You are not on an issue branch")
				return
			} else {
				var ok bool
				issueId, ok = parseBranchNameForIssue(branchName, teamKeys)
				if !ok {
					log.Fatalf("Failed to parse issue identifier from branch name: %s", branchName)
				}
			}
		}

		for _, issue := range issues {
			if issue.Identifier == issueId {
				fmt.Printf("Opening %s...\n", issue.Identifier)
				util.OpenURL(issue.Url)
				return
			}
		}

		fmt.Printf("No issue found with identifier %s\n", issueId)
	},
}

func parseBranchNameForIssue(branchName string, teamKeys []string) (issueId string, ok bool) {
	pattern := "((?:" + strings.Join(teamKeys, "|") + ")-\\d+)"
	var re = regexp.MustCompile(fmt.Sprintf("(?i)%s", pattern))
	if len(re.FindStringIndex(branchName)) > 0 {
		issueId, ok = strings.ToUpper(re.FindString(branchName)), true
	}
	return
}
