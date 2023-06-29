package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

/*
#define _GNU_SOURCE
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h> // for open
#include <unistd.h> // for close

__attribute__((constructor)) void enter_namespace(void) {
	char *mycontainer_pid;
	mycontainer_pid = getenv("mycontainer_pid");
	if (!mycontainer_pid) {
		return;
	}
	char *mycontainer_cmd;
	mycontainer_cmd = getenv("mycontainer_cmd");
	if (!mycontainer_cmd) {
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", mycontainer_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);
		setns(fd,0);
		close(fd);
	}
	int res = system(mycontainer_cmd);
	exit(0);
	return;

}
*/
import "C"

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			mydocker run -ti [imageName] [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		//提供run后面的-name指定的容器名字参数
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container command")
		}
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		//get image name
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]

		tty := context.Bool("ti")
		detach := context.Bool("d")
		if tty && detach {
			return fmt.Errorf("ti and d parameter can not both provided")
		}
		//cmd := context.Args().Get(0)
		resourceConf := &ResourceConfig{
			memoryMax: context.String("m"),
		}
		volume := context.String("v")
		containerName := context.String("name")
		Run(tty, cmdArray, resourceConf, volume, containerName, imageName)
		//Run(tty, cmdArray)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		err := RunContainerInitProcess()
		fmt.Print(err)
		return err
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the containers",
	Action: func(context *cli.Context) error {
		//ListContainers()
		//找到存储容器信息的路径/var/run/mydocker
		dirURL := fmt.Sprintf(DefaultInfoLocation, "")
		dirURL = dirURL[:len(dirURL)-1]
		//读取该文件夹下的所有文件
		files, err := ioutil.ReadDir(dirURL)
		if err != nil {
			log.Errorf("Read dir %s error %v", dirURL, err)
			return nil
		}
		var containers []*ContainerInfo
		for _, file := range files {
			tmpContainer, err := getContainerInfo(file.Name())
			if err != nil {
				log.Errorf("Get container info error %v", err)
				continue
			}
			containers = append(containers, tmpContainer)
		}
		//使用tabwrite.NewWrite在控制台打印出容器信息
		//tabwrite是引用text/tabwriter类库，用于在控制台打印对齐的表哥
		w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
		fmt.Fprintf(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
		for _, item := range containers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				item.Id,
				item.Name,
				item.Pid,
				item.Status,
				item.Command,
				item.CreatedTime)
		}
		//刷新标准输出流缓存区，将容器列表打印出来
		if err := w.Flush(); err != nil {
			log.Errorf("Flush error %v", err)
			return nil
		}
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			log.Errorf("Please input your container name")
		}
		containerName := context.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into cotainer",
	Action: func(context *cli.Context) error {
		log.Info("ExecCommand Run")
		//如果没有设置pid，那么ns切换将不能实现，需要先设置然后再继续
		if os.Getenv("mycontainer_pid") == "" {
			if len(context.Args()) < 2 {
				log.Errorf("Missing container name or command")
				return fmt.Errorf("Missing container name or command")
			}
			containerName := context.Args().Get(0)
			var commandArray []string
			for _, arg := range context.Args().Tail() {
				commandArray = append(commandArray, arg)
			}
			pid, err := getContainerPidByName(containerName)
			if err != nil {
				log.Errorf("get container pid by name failed%v", err)
				return nil
			}
			//totalCmdstr := strings.Join(context.Args(), " ")
			cmd := exec.Command("/proc/self/exe", "exec")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmdStr := strings.Join(commandArray, " ")
			os.Setenv("mycontainer_pid", pid)
			os.Setenv("mycontainer_cmd", cmdStr)
			if err := cmd.Run(); err != nil {
				log.Errorf("Exec container %s error%v", containerName, err)
			}

			return nil
		}

		os.Unsetenv("mycontainer_pid")
		os.Unsetenv("mycontainer_cmd")
		return nil
	},
}

//期望调用的是"main stop 容器名"这种方式，需要在开始的时候检测是否输入了容器名
var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}
		containerName := context.Args().Get(0)
		stopContainer(containerName)
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove unused containers",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container Name")
		}
		containerName := context.Args().Get(0)
		removeContainer(containerName)
		return nil
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 2 {
			return fmt.Errorf("Missing container name and image name")
		}
		containerName := context.Args().Get(0)
		imageName := context.Args().Get(1)
		commitContainer(containerName, imageName)
		return nil
	},
}
