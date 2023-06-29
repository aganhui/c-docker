package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"c-docker/config"
	"c-docker/container"
)

/*
删除容器，逻辑比较简单，主要分为以下4个步骤
1.根据容器名查找容器信息
2.判断容器是否处于停止状态
3.查找容器存储信息的地址
4.移除记录容器信息的地址
*/
func removeContainer(containerName string) {
	containerInfo, err := container.GetContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerName, err)
		return
	}
	if containerInfo.Status != container.STOP {
		log.Errorf("Couldn't remove running container")
		return
	}
	dirURL := fmt.Sprintf(config.GlobalDefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove file %s error %v", dirURL, err)
		return
	}
}
