package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"

	"github.com/go-git/go-git/v5"
)

var pythonRepoDir string = "/tmp/pungi-test/repos"

func verifyDependencies() error {
	if err := exec.Command("poetry", "--version").Run(); err != nil {
		return err
	}
	return nil
}

type PoetryRepo struct {
	GitURL    string
	LocalPath string
}

func (repo PoetryRepo) Create() error {
	// clone the repo
	_, err := git.PlainClone(repo.LocalPath, false, &git.CloneOptions{
		URL: repo.GitURL,
	})
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		return err
	}

	// install the repo's dependencies
	cwd, err := os.Getwd()
	if err != nil {
		return errors.New("problem getting current working directory")
	}
	cwd = path.Clean(cwd)
	if os.Chdir(repo.LocalPath) != nil {
		return errors.New("could not go to python repo")
	}
	defer os.Chdir(cwd)
	if err := exec.Command("poetry", "install", "--sync").Run(); err != nil {
		return err
	}

	return nil
}

type PoetryScript struct {
	Argv      []string
	LocalRepo PoetryRepo
}

func (script PoetryScript) Run() (string, error) {
	// move into python repo
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.New("problem getting current working directory")
	}
	cwd = path.Clean(cwd)
	if os.Chdir(script.LocalRepo.LocalPath) != nil {
		return "", errors.New("could not go to python repo")
	}
	defer os.Chdir(cwd)

	var poetryRunPrefix = [2]string{"run", "python"}
	var argv []string = append(poetryRunPrefix[:], script.Argv...)
	script_output, err := exec.Command("poetry", argv...).CombinedOutput()
	return string(script_output), err
}

func main() {
	if err := os.MkdirAll(pythonRepoDir, fs.FileMode(0777)); err != nil {
		panic(err)
	}
	defer os.RemoveAll(pythonRepoDir)
	if err := verifyDependencies(); err != nil {
		panic(err)
	}

	repo := PoetryRepo{
		GitURL:    "https://gitea.com/kybouw/pungi-test-public.git",
		LocalPath: path.Join(pythonRepoDir, "pungi-test-public"),
	}
	if err := repo.Create(); err != nil {
		panic(err)
	}
	argv := []string{"pungi_test_public/hello.py"}
	script := PoetryScript{argv, repo}

	script_output, err := script.Run()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(script_output)
}
