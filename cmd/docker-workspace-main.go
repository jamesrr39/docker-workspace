package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	setupInitCmd()
	setupResumeCmd()
	setupStopCmd()
	kingpin.Parse()
}

func setupInitCmd() {
	cmd := kingpin.Command("init", "create a new docker-workspace")
	imageName := cmd.Arg("image-name", "name for the newly created docker image").Required().String()
	cmd.Action(func(ctx *kingpin.ParseContext) error {
		return runInit(*imageName)
	})
}
func setupResumeCmd() {
	cmd := kingpin.Command("resume", "resume a previously running workspace")
	imageName := cmd.Arg("image-name", "name of the docker image").Required().String()
	cmd.Action(func(ctx *kingpin.ParseContext) error {
		return runResume(*imageName)
	})
}
func setupStopCmd() {
	cmd := kingpin.Command("stop", "stop a running workspace")
	imageName := cmd.Arg("image-name", "name of the docker image").Required().String()
	cmd.Action(func(ctx *kingpin.ParseContext) error {
		return runStop(*imageName)
	})
}

type environmentFile struct {
	Name     string
	Contents []byte
}

// create the docker container, store the name of the docker container in a file
func runInit(imageName string) error {
	err := os.Mkdir(imageName, 0700)
	if err != nil {
		return err
	}

	err = os.Mkdir(filepath.Join(imageName, "workspace"), 0700)
	if err != nil {
		return err
	}

	dockerWorkspaceFileBytes, err := yaml.Marshal(DockerWorkspaceFile{imageName})
	if err != nil {
		return err
	}

	environmentFiles := []environmentFile{
		environmentFile{
			Name:     "Dockerfile",
			Contents: []byte(dockerfileContents),
		},
		environmentFile{
			Name:     "docker-workspace.yml",
			Contents: dockerWorkspaceFileBytes,
		},
	}

	for _, fileToWrite := range environmentFiles {
		err = ioutil.WriteFile(filepath.Join(imageName, fileToWrite.Name), fileToWrite.Contents, 0600)
		if err != nil {
			return err
		}
	}

	err = runCommandThroughPipes("docker", "build", "-t"+imageName, imageName)
	if err != nil {
		return err
	}

	return runResume(imageName)
}

func getContainerNameFromImageName(imageName string) string {
	return strings.Replace(imageName, "/", "__", -1)
}

func runResume(imageName string) error {
	containerName := getContainerNameFromImageName(imageName)
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	hostDisplay := os.Getenv("DISPLAY")
	err = runCommandThroughPipes(
		"docker",
		"run",
		"--name="+containerName,
		"-it",
		"-v/tmp/.X11-unix:/tmp/.X11-unix",
		"-eDISPLAY=unix"+hostDisplay,
		fmt.Sprintf("-v%s/workspace:/home/user/workspace", wd),
		imageName,
	)

	if err != nil {
		return err
	}

	return runCommitAndCleanup(imageName)
}

func runStop(imageName string) error {
	containerName := getContainerNameFromImageName(imageName)

	err := runCommandThroughPipes("docker", "stop", containerName, imageName)
	if err != nil {
		return err
	}

	return runCommitAndCleanup(imageName)
}
func runCommitAndCleanup(imageName string) error {
	containerName := getContainerNameFromImageName(imageName)

	err := runCommandThroughPipes("docker", "commit", containerName, imageName)
	if err != nil {
		return err
	}

	return runCommandThroughPipes("docker", "rm", containerName)
}

func runCommandThroughPipes(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

type DockerWorkspaceFile struct {
	ImageName string
}

const dockerfileContents = `FROM ubuntu:18.04

RUN apt-get update && apt-get install -y git sudo wget unzip

RUN adduser --disabled-password --gecos "" user && usermod -aG sudo user && echo "user\nuser\n" | passwd user

WORKDIR /home/user

USER user
`

// suggested format
// - Dockerfile
// - docker-workspace.yml
