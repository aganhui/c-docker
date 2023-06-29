package container

/*
此部分代码用于完善容器文件系统，包括创建只读层、创建挂载点和创建可写层
*/
/*
TODO 目前所有的容器都会挂载到同一个可写层，其实可以设置成每个容器挂在不同的可读可写层上
*/
import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"c-docker/config"
)

// 将busybox.tar解压到busybox目录下，作为容器的只读层

func NewWorkSpace(volume string, imageName string, containerName string) {
	CreateReadOnlyLayer(imageName)
	CreateWorkLayer(containerName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName, imageName)
	CreateVolumeLayer(volume, containerName)
}

/*
*
TODO 解耦合
*/
func CreateVolumeLayer(volume string, containerName string) {
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
				if err := os.MkdirAll(parentUrl, 0777); err != nil {
					log.Infof("MkdirAll parent dir %s error.%v", parentUrl, err)
				}
			}
			containerUrl := volumeURLs[1]
			mntURL := fmt.Sprintf(config.MntUrl, containerName)
			containerVolumeURL := path.Join(mntURL, containerUrl)
			if err := os.MkdirAll(containerVolumeURL, 0777); err != nil {
				log.Infof("MkdirAll container dir %s error.%v", containerVolumeURL, err)
			}
			// println("lowerdir=" + parentUrl + ",upperdir=" + containerVolumeURL + ",workdir=" + path.Join(rootURL, ".worker"))
			// dirs := "lowerdir=" + parentUrl + ",upperdir=" + containerVolumeURL + ",workdir=" + path.Join(rootURL, ".worker")
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
	unTarFolderUrl := fmt.Sprintf(config.ImageUrl, imageName)
	log.Infof("tar folder: %s", unTarFolderUrl)
	imageUrl := config.RootUrl + "/" + imageName + ".tar"
	exist, err := PathExists(unTarFolderUrl)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", unTarFolderUrl, err)
	}
	if exist == false {
		if err := os.MkdirAll(unTarFolderUrl, 0777); err != nil {
			log.Errorf("MkdirAll dir %s error. %v", unTarFolderUrl, err)
		}
		log.Infof("image: %s; untar folder: %s", imageUrl, unTarFolderUrl)
		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", unTarFolderUrl, err)
		}
	}
}

func CreateWorkLayer(containerName string) {
	writeURL := fmt.Sprintf(config.WorkUrl, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Errorf("MkdirAll dir %s error. %v", writeURL, err)
	}
}

func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(config.WriteLayerUrl, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Errorf("MkdirAll dir %s error. %v", writeURL, err)
	}
}

func CreateMountPoint(containerName string, imageName string) {
	mntUrl := fmt.Sprintf(config.MntUrl, containerName)
	// log.Infof("root url: %s; ygh testing: %s, mnt url: %s", RootUrl, rootUrl, mntUrl)
	if err := os.MkdirAll(mntUrl, 0777); err != nil {
		log.Errorf("MkdirAll dir %s error. %v", mntUrl, err)
	}

	imageUrl := fmt.Sprintf(config.ImageUrl, imageName)
	writeLayer := fmt.Sprintf(config.WriteLayerUrl, containerName)
	workUrl := fmt.Sprintf(config.WorkUrl, containerName)
	dirs := "lowerdir=" + imageUrl + ",upperdir=" + writeLayer + ",workdir=" + workUrl
	println(dirs)

	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount %v", err)
	}
}

// Delete the AUFS filesystem while container exit
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
	mntURL := fmt.Sprintf(config.MntUrl, containerName)
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
	writeURL := fmt.Sprintf(config.WriteLayerUrl, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
	}
}

// 其实不是删除，只是卸载了这个目录
func DeleteVolumeLayer(volumeURLs []string, containerName string) {
	mntURL := fmt.Sprintf(config.MntUrl, containerName)
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
