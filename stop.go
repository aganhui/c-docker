package main

/*
这部分的代码用于stop正在运行的容器
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"

	"c-docker/config"
	"c-docker/container"
)

func stopContainer(containerName string) {
	pid, err := container.GetContainerPidByName(containerName)
	if err != nil {
		log.Errorf("Get contaienr pid by name %s error %v", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		log.Errorf("Conver pid from string to int error %v", err)
		return
	}
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("Stop container %s error %v", containerName, err)
		return
	}
	containerInfo, err := container.GetContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerName, err)
		return
	}
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Json marshal %s error %v", containerName, err)
		return
	}
	dirURL := fmt.Sprintf(config.GlobalDefaultInfoLocation, containerName)
	configFilePath := path.Join(dirURL, config.GlobalConfigName)
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		log.Errorf("Write file %s error", configFilePath, err)
	}
}
