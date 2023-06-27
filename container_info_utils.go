package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

/*
此页代码包含了容器引擎信息相关的内容
*/
//定义了容器引擎中可以用到的结构体
type ContainerInfo struct {
	Pid         string `json:"pid"`        //容器init进程在宿主机上的PID
	Id          string `json:"id"`         //容器id
	Name        string `json:"name"`       //容器名
	Command     string `json:"command"`    //容器内init进程的运行命令
	CreatedTime string `json:"createTime"` //创建时间
	Status      string `json:"status"`     //容器的状态
	Volume      string `json:"volume"`     //容器数据卷
}

//一些常量的定义,在程序内部使用，不能被随意更改
var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = globalDefaultInfoLocation
	ConfigName          string = globalConfigName
)

func recordContainerInfo(containerPID int, commandArray []string, containerName string, id string,volume string) (string, error) {
	//首先生成10位的数字的容器ID
	//id := randStringBytes(10)
	//以当前时间为容器创建时间
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	//如果用户不指定容器名，那么就以容器id当作容器名
	//if containerName == "" {
	//	containerName = id
	//}
	//生成容器信息的结构体实例
	containerInfo := &ContainerInfo{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      RUNNING,
		Name:        containerName,
		Volume:      volume,
	}
	//将容器信息的对象json序列化成字符串
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)
	//拼凑一下存储容器信息的路径
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	log.Info("createJsonPath: " + dirUrl)
	//如果该路径不存在，就级联地全部创建
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.Errorf("Mkdir error %s error %v", dirUrl, err)
		return "", err
	}
	//fileName := dirUrl + "/" + ConfigName
	fileName := path.Join(dirUrl, ConfigName)
	//创建最终的配置文件——config.json文件
	file, err := os.Create(fileName)

	if err != nil {
		log.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	//将json化后的数据写入到文件中
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return "", err
	}
	defer file.Close()
	return containerName, nil
}

func deleteContainerInfo(containerName string) {
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove dir %s error %v", dirURL, err)
	}
}

/*
一个随机生成器，可以随机表示n位的数字
*/
func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func getContainerInfo(containerName string) (*ContainerInfo, error) {
	//containerName := file.Name()
	configFileDir := fmt.Sprintf(DefaultInfoLocation, containerName)
	configFileDir = path.Join(configFileDir, globalConfigName)
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		log.Errorf("Read file %s error %v", configFileDir, err)
		return nil, err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("Json unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil

}

func getContainerPidByName(containerName string) (string, error) {
	dirURL := fmt.Sprintf(globalDefaultInfoLocation, containerName)
	configFilePath := path.Join(dirURL, globalConfigName)
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	/*
		如果容器停止了，pid也不存在，所以没有再查询的必要，直接报错
	*/
	if containerInfo.Pid == "" || containerInfo.Pid == " " {
		log.Errorf("This container has been stopped.No Pid exists!")
		return "", fmt.Errorf("This container has been stopped.No Pid exists!")
	}
	return containerInfo.Pid, nil
}

//根据容器名获取对应的struct结构
func getContainerInfoByName(containerName string) (*ContainerInfo, error) {
	//dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	//configFilePath := path.Join(dirURL, globalConfigName)
	containerinfo, err := getContainerInfo(containerName)
	if err != nil {
		log.Errorf("GetContainerInfoByNameFailed%v", err)
		return nil, err
	}
	return containerinfo, nil

}
