package main

import (
	"fmt"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"
)

func commitContainer(imageName string) {
	imageTar := path.Join(globalCommitURL, imageName+".tar")
	fmt.Print("%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", globalMntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", globalMntURL, err)
	}
}
