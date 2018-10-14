package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	setupInitCmd()
	setupResumeCmd()
	kingpin.Parse()
}

func setupInitCmd() {
	initCmd := kingpin.Command("init", "create a new docker-workspace")
	imageName := initCmd.Arg("image-name", "name for the newly created docker image").Required().String()
	initCmd.Action(func(ctx *kingpin.ParseContext) error {
		return runInit(*imageName)
	})
}
func setupResumeCmd() {
	resumeCmd := kingpin.Command("resume", "resume a previously running workspace")
	imageName := resumeCmd.Arg("image-name", "name of the docker image").Required().String()
	resumeCmd.Action(func(ctx *kingpin.ParseContext) error {
		return runResume(*imageName)
	})
}

// create the docker container, store the name of the docker container in a file
func runInit(imageName string) error {
	err := runCommandThroughPipes("docker", "build", "-t"+imageName, ".")
	if err != nil {
		return err
	}

	return runResume(imageName)
}

func runResume(imageName string) error {
	containerName := strings.Replace(imageName, "/", "__", -1)
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

	err = runCommandThroughPipes("docker", "commit", containerName, imageName)
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
