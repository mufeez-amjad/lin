package git

import (
	"fmt"
	"log"
	"os"

	git "github.com/libgit2/git2go/v34"
)

func getRepo() (*git.Repository, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Error getting current working directory: %v\n", err)
	}

	return git.OpenRepository(cwd)
}

func GetCurrentBranch() string {
	r, err := getRepo()
	if err != nil {
		log.Fatal(err)
	}

	head, _ := r.Head()
	return string(head.Name())
}

func doesBranchExist(repo *git.Repository, branchName string) (bool, error) {
	refs, err := repo.NewReferenceIterator()
	if err != nil {
		return false, err
	}

	branchExists := false
	for ref, err := refs.Next(); ref != nil; {
		if err != nil {
			return false, err
		}

		if !ref.IsBranch() {
			continue
		}

		refName, err := ref.Branch().Name()
		if err != nil {
			return false, err
		}

		if refName == branchName {
			branchExists = true
			break
		}
	}

	return branchExists, nil
}

func CheckoutBranch(branchName string) error {
	r, err := getRepo()
	if err != nil {
		log.Fatal(err)
	}
	defer r.Free()

	branch, err := r.LookupBranch(branchName, git.BranchAll)
	if err != nil {
		return fmt.Errorf("Error checking branch existence: %v\n", err)
	}
	defer branch.Free()

	if branch != nil {
		err = branch.Owner().CheckoutHead(&git.CheckoutOptions{
			Strategy: git.CheckoutSafe,
		})
		if err != nil {
			return fmt.Errorf("Failed to checkout existing branch: %s", err)
		}
	} else {
		head, err := r.Head()
		defer head.Free()
		commit, err := r.LookupCommit(head.Target())
		if err != nil {
			// Handle the error
		}
		defer commit.Free()

		branch, err = r.CreateBranch(branchName, commit, false)
		if err != nil {
			return fmt.Errorf("Failed to create branch: %s", err)
		}

		err = branch.Owner().CheckoutHead(&git.CheckoutOptions{
			Strategy: git.CheckoutSafe,
		})
		if err != nil {
			return fmt.Errorf("Failed to checkout new branch: %s", err)
		}
	}

	fmt.Printf("Checked out branch '%s'\n", branchName)
	return nil
}
