package main

import (
	"fmt"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"

	"c-docker/config"
)

func commitContainer(containerName string, imageName string) {
	mntURL := fmt.Sprintf(config.MntUrl, containerName)
	mntURL += "/"

	imageTar := path.Join(config.RootUrl, imageName+".tar")

	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}
