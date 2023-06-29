package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"

	"c-docker/network"
)

func NewParentProcess(
	tty bool,
	containerName string,
	volume string,
	imageName string,
) (*exec.Cmd, *os.File) {
	// args := []string{"init", command}
	// 创建新的管道
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Print(err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		/*
			把输出重定向也设置为可配置的
		*/
		//file, _ := os.Create("output.txt")
		//cmd.Stdin = file
		//cmd.Stdout = file
		//cmd.Stderr = file
		//重定向到配置目录对应的container.log文件
		dirURL := fmt.Sprintf(globalDefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			log.Errorf("NewParentProcess mkdir %s error %v", dirURL, err)
			return nil, nil
		}
		// log.Info("pathName%s", globalLogName)
		log.Info("dirurl%s", containerName)
		stdLogFilePath := path.Join(dirURL, globalLogName)
		// log.Info("build config in dir %s", stdLogFilePath)
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			log.Errorf("NewParentProcess create file %s error %v", stdLogFile, err)
			return nil, nil
		}
		cmd.Stdin = stdLogFile
		cmd.Stdout = stdLogFile
		cmd.Stderr = stdLogFile

	}
	/*
		这里的路径以后可以做修改
		TODO
		这两个变量穿插在两个函数间，很有可能使得后面维护困难，可以做修改
	*/

	NewWorkSpace(volume, imageName, containerName)

	cmd.Dir = fmt.Sprintf(MntUrl, containerName)

	cmd.ExtraFiles = []*os.File{readPipe}
	return cmd, writePipe
}

func Run(
	tty bool,
	comArray []string,
	res *ResourceConfig,
	volume string,
	containerName string,
	imageName string,
	nw string,
	portmapping []string,
) {
	containerId := randStringBytes(10)
	if containerName == "" {
		containerName = containerId
	}
	parent, writePipe := NewParentProcess(tty, containerName, volume, imageName)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		fmt.Print("error in parent start")
		log.Error(err)
	}
	// 记录容器信息
	containerName, err := recordContainerInfo(
		parent.Process.Pid,
		comArray,
		containerName,
		containerId,
		volume,
	)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}
	// 记录容器信息结束
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
	cgroupmanager := getDefaultCgroupV2Manager()
	defer cgroupmanager.Destory()
	// 没有准备初始化函数，所以需要手动设置一下变量
	cgroupmanager.relativepath = "c-docker-cgroup"
	cgroupmanager.Set(res)
	cgroupmanager.Apply(parent.Process.Pid)
	if nw != "" {
		// config container network
		network.Init()
		containerInfo := &network.ContainerInfo{
			Id:          containerId,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portmapping,
		}
		if err := network.Connect(nw, containerInfo); err != nil {
			log.Errorf("Error Connect Network %v", err)
			return
		}
	}
	if tty {
		parent.Wait()
		os.Chdir(globalExeLocation)
		DeleteWorkSpace(volume, containerName)
		deleteContainerInfo(containerName)

	}

	os.Exit(0)
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}
