package git

import (
	"fmt"
	"log"
	"os"

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
	return string(head.Name())
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
		fmt.Printf("Error getting worktree: %v\n", err)
		os.Exit(1)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)),
		Create: !branchExists,
		Keep:   true,
	})
	if err != nil {
		fmt.Printf("Error checking out branch: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Checked out branch '%s'\n", branchName)
	return nil
}
