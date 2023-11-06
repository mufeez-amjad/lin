package git

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func getRepo() (*git.Repository, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Error getting current working directory: %v\n", err)
	}

	r, err := git.PlainOpen(cwd)
	if err != nil {
		return nil, fmt.Errorf("Error opening repository: %v\n", err)
	}

	return r, nil
}

func GetCurrentBranch() string {
	r, err := getRepo()
	if err != nil {
		log.Fatal(err)
	}

	head, _ := r.Head()
	return string(head.Name().Short())
}

func CheckoutBranch(branchName string) error {
	r, err := getRepo()
	if err != nil {
		log.Fatal(err)
	}

	refs, err := r.Branches()
	if err != nil {
		return fmt.Errorf("Error getting branches: %v\n", err)
	}

	branchExists := false
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().Short() == branchName {
			branchExists = true
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error checking branch existence: %v\n", err)
	}

	// Checkout the branch
	w, err := r.Worktree()
	if err != nil {
		log.Fatalf("Error getting worktree: %s", err)
	}

	// Retrieve the base commit to reset changes to
	commit, err := r.Head()

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: !branchExists,
		Keep:   true,
	})
	if err != nil {
		log.Fatalf("Error checking out branch: %s", err)
	}

	err = w.Reset(&git.ResetOptions{
		Commit: commit.Hash(),
		Mode:   git.SoftReset,
	})
	if err != nil {
		log.Fatalf("Error checking out branch: %s", err)
	}

	fmt.Printf("Checked out branch '%s'\n", branchName)
	return nil
}

func FindBranches(issueIdentifier string) ([]*plumbing.Reference, error) {
	r, err := getRepo()
	if err != nil {
		return nil, err
	}

	refs, err := r.Branches()
	if err != nil {
		return nil, err
	}
	defer refs.Close()

	pattern := fmt.Sprintf("(?mi)%s[^\\d]", issueIdentifier)
	re := regexp.MustCompile(pattern)

	var branches []*plumbing.Reference
	for {
		ref, err := refs.Next()
		if err != nil {
			break
		}

		branchName := ref.Name().Short()
		if len(re.FindStringIndex(branchName)) > 0 {
			branches = append(branches, ref)
		}
	}

	return branches, nil
}
