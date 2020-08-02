package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

func main() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	rmCmd := flag.NewFlagSet("rm", flag.ExitOnError)
	lsCmd := flag.NewFlagSet("ls", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("expected 'add', 'rm', or 'ls' subcommands")
		os.Exit(10)
	}

	repoPath, err := findRepository(".")
	if err != nil {
		panic("Unable to locate git repository")
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		panic("Unable to open repository")
	}
	config, err := repo.Config()
	if err != nil {
		panic("Unable to open repository configuration")
	}

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		add(repo, config, addCmd.Args()[0])
	case "rm":
		rmCmd.Parse(os.Args[2:])
		remove(repo, config, rmCmd.Args()[0])
	case "ls":
		lsCmd.Parse(os.Args[2:])
		ls(config)
	default:
		fmt.Println("expected 'add', 'rm', or 'ls' subcommands")
		os.Exit(10)
	}

}

func findRepository(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	nextPath := filepath.Join(abs, "..")
	if len(nextPath) == 1 {
		return "", fmt.Errorf("Unable to locate repository")
	}

	gitDirectory := filepath.Join(abs, ".git")
	if _, err := os.Stat(gitDirectory); os.IsNotExist(err) {
		return findRepository(filepath.Join(abs, ".."))
	}
	return abs, nil
}

func inflateBranch(branch string) config.RefSpec {
	return config.RefSpec(fmt.Sprintf("+refs/heads/%s*:refs/remotes/origin/%s*", branch, branch))
}

func deflateBranch(refSpec config.RefSpec) string {
	branch := refSpec.String()
	branch = branch[12:]
	branch = branch[:strings.Index(branch, "*:")]

	return branch
}

func ls(config *config.Config) {
	remotes := config.Remotes
	for _, r := range remotes {
		refSpec := r.Fetch

		for _, s := range refSpec {
			branchName := deflateBranch(s)
			if len(branchName) != 0 {
				fmt.Println(branchName)
			}
		}
	}
}

func add(repo *git.Repository, cfg *config.Config, branch string) {
	remotes := cfg.Remotes
	for _, r := range remotes {
		refSpec := r.Fetch

		refSpec = append(refSpec, inflateBranch(branch))
		r.Fetch = refSpec
	}

	cfg.Remotes = remotes
	err := repo.SetConfig(cfg)
	if err != nil {
		panic(fmt.Errorf("failure to set config: %w", err))
	}
}

func remove(repo *git.Repository, cfg *config.Config, branch string) {
	remotes := cfg.Remotes
	for _, r := range remotes {
		refSpec := r.Fetch

		// newRefSpec := make([]config.RefSpec, len(refSpec))
		newRefSpec := make([]config.RefSpec, 0)
		for _, s := range refSpec {
			if deflateBranch(s) == branch {
				continue
			}
			newRefSpec = append(newRefSpec, s)
		}
		r.Fetch = newRefSpec
	}

	cfg.Remotes = remotes
	err := repo.SetConfig(cfg)
	if err != nil {
		panic(fmt.Errorf("failure to set config: %w", err))
	}
}
