package main

/*
此部分代码用于完善容器文件系统，包括创建只读层、创建挂载点和创建可写层
*/
/*
TODO 目前所有的容器都会挂载到同一个可写层，其实可以设置成每个容器挂在不同的可读可写层上
*/
import (
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

//将busybox.tar解压到busybox目录下，作为容器的只读层

func NewWorkSpace(rootURL string, mntURL string, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateWorkLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)
	CreateVolumeLayer(rootURL, mntURL, volume)
}

/**
TODO 解耦合
*/
func CreateVolumeLayer(rootURL string, mntURL string, volume string) {
	if volume != "" {
		var volumeURLs []string
		volumeURLs = strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			parentUrl := volumeURLs[0]
			exist, err := PathExists(parentUrl)
			if err != nil {
				log.Infof("Fail to judge whether dir %s exists. %v", parentUrl, err)
			}
			if !exist {
				if err := os.Mkdir(parentUrl, 0777); err != nil {
					log.Infof("Mkdir parent dir %s error.%v", parentUrl, err)
				}
			}
			containerUrl := volumeURLs[1]
			containerVolumeURL := path.Join(mntURL, containerUrl)
			if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
				log.Infof("Mkdir container dir %s error.%v", containerVolumeURL, err)
			}
			//println("lowerdir=" + parentUrl + ",upperdir=" + containerVolumeURL + ",workdir=" + path.Join(rootURL, ".worker"))
			//dirs := "lowerdir=" + parentUrl + ",upperdir=" + containerVolumeURL + ",workdir=" + path.Join(rootURL, ".worker")
			cmd := exec.Command("mount", "-o", "bind", parentUrl, containerVolumeURL)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Errorf("%v", err)
			}

			log.Infof("%q", volumeURLs)
		} else {
			log.Infof("Volume parameter input is not correct")
		}

	}

}

// 这些函数在错误处理方面实现的不是很好
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := path.Join(rootURL, "busybox")
	busyboxTarURL := path.Join(rootURL, "busybox.tar")
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", busyboxURL, err)
		}
	}
}

func CreateWorkLayer(rootURL string) {
	writeURL := path.Join(rootURL, ".worker")
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", writeURL, err)
	}
}

func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", writeURL, err)
	}
}

func CreateMountPoint(rootURL string, mntURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", mntURL, err)
	}
	//dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	println("lowerdir=" + path.Join(rootURL, "busybox") + ",upperdir=" + path.Join(rootURL, "writeLayer") + ",workdir=" + path.Join(rootURL, ".worker"))
	dirs := "lowerdir=" + path.Join(rootURL, "busybox") + ",upperdir=" + path.Join(rootURL, "writeLayer") + ",workdir=" + path.Join(rootURL, ".worker")

	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {

		log.Errorf("%v", err)
	}
}

//Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	DeleteVolumeLayer(rootURL, mntURL, volume)
	DeleteMountPoint(rootURL, mntURL)
	DeleteWriteLayer(rootURL)
	DeleteWorkerLayer(rootURL)
}

func DeleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
		os.Exit(-1)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
		os.Exit(-1)
	}
}

func DeleteWriteLayer(rootURL string) {
	writeURL := path.Join(rootURL, "writeLayer")
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
	}
}

func DeleteWorkerLayer(rootURL string) {
	writeURL := path.Join(rootURL, ".worker")
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
	}
}

//其实不是删除，只是卸载了这个目录
func DeleteVolumeLayer(rootURL string, mntURL string, volume string) {
	if volume != "" {
		var volumeURLs []string
		volumeURLs = strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			parentUrl := volumeURLs[0]
			_, err := PathExists(parentUrl)
			if err != nil {
				log.Infof("Fail to judge whether dir %s exists. %v", parentUrl, err)
			}
			containerUrl := volumeURLs[1]
			containerVolumeURL := path.Join(mntURL, containerUrl)
			cmd := exec.Command("umount", containerVolumeURL)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				log.Infof("Umount mountpoint failed.%v", err)
			}
		} else {
			log.Infof("Volume parameter input is not correct")
		}

	}

}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
