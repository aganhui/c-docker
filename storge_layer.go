package main

/*
此部分代码用于完善容器文件系统，包括创建只读层、创建挂载点和创建可写层
*/
/*
TODO 目前所有的容器都会挂载到同一个可写层，其实可以设置成每个容器挂在不同的可读可写层上
*/
import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"path"
	"fmt"
)

//将busybox.tar解压到busybox目录下，作为容器的只读层

func NewWorkSpace(volume string, imageName string, containerName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName,imageName)
	CreateVolumeLayer(volume,containerName)
}

/**
TODO 解耦合
*/
func CreateVolumeLayer(volume string,containerName string) {
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
			mntURL := fmt.Sprintf(MntUrl, containerName)
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
func CreateReadOnlyLayer(imageName string) {
	unTarFolderUrl := path.Join(RootUrl, imageName)
	imageUrl := RootUrl + "/" + imageName + ".tar"
	exist, err := PathExists(unTarFolderUrl)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v",unTarFolderUrl , err)
	}
	if exist == false {
		if err := os.Mkdir(unTarFolderUrl, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", unTarFolderUrl, err)
		}
		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", unTarFolderUrl, err)
		}
	}
}

func CreateWorkLayer(rootURL string) {
	writeURL := path.Join(rootURL, ".worker")
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", writeURL, err)
	}
}

func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", writeURL, err)
	}
}

func CreateMountPoint(containerName string, imageName string) {
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	rootUrl :=fmt.Sprintf(RootUrl, containerName)
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", mntUrl, err)
	}
	//dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	println("lowerdir=" + path.Join(rootUrl, "busybox") + ",upperdir=" + path.Join(rootUrl, "writeLayer") + ",workdir=" + path.Join(rootUrl, ".worker"))
	dirs := "lowerdir=" + path.Join(rootUrl, "busybox") + ",upperdir=" + path.Join(rootUrl, "writeLayer") + ",workdir=" + path.Join(rootUrl, ".worker")

	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {

		log.Errorf("%v", err)
	}
}

//Delete the AUFS filesystem while container exit
func DeleteWorkSpace(volume string, containerName string) {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteVolumeLayer(volumeURLs, containerName)
		}
	}
	DeleteMountPoint(containerName)
	DeleteWriteLayer(containerName)
}

func DeleteMountPoint(containerName string) {
	mntURL := fmt.Sprintf(MntUrl, containerName)
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

func DeleteWriteLayer(containerName string) {
	writeURL := path.Join(containerName, "writeLayer")
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
	}
}

//其实不是删除，只是卸载了这个目录
func DeleteVolumeLayer(volumeURLs []string,containerName string) {
	mntURL := fmt.Sprintf(MntUrl, containerName)
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