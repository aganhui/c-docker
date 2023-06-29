package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	log "github.com/sirupsen/logrus"

	"c-docker/config"
)

/*
此页包含了容器日志展示需要的内容
*/

func logContainer(containerName string) {
	dirURL := fmt.Sprintf(config.GlobalDefaultInfoLocation, containerName)
	logFileLocation := path.Join(dirURL, config.GlobalLogName)
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		log.Errorf("Log container open file %s error %v", logFileLocation, err)
		return
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Log container read file %s error %v", logFileLocation, err)
		return
	}
	fmt.Fprintf(os.Stdout, string(content))
}
