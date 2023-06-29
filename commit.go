package main

import (
	"fmt"
	"path"
	"os/exec"
	log "github.com/sirupsen/logrus"
)

func commitContainer(containerName string, imageName string){
	mntURL := fmt.Sprintf(MntUrl, containerName)
	mntURL += "/"

	imageTar := path.Join(RootUrl, imageName+".tar")

	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}