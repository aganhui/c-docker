package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func RunContainerInitProcess() error {

	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command error, cmdArray is nil")
	}
	//defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV | syscall.MS_PRIVATE | syscall.MS_REC
	//syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	setUpMount()
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}
	log.Infof("Find path%s", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil

	/*
		cmdArray := readUserCommand()
		if cmdArray == nil || len(cmdArray) == 0 {
			return fmt.Errorf("Run container get user command error, cmdArray is nil")
		}

		//setUpMount()

		path, err := exec.LookPath(cmdArray[0])
		if err != nil {
			log.Errorf("Exec loop path error %v", err)
			return err
		}
		log.Infof("Find path %s", path)
		if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
			log.Errorf(err.Error())
		}
		return nil*/
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

/**
Init 挂载点
*/
func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error %v", err)
		return
	}
	log.Infof("Current location is %s", pwd)
	pivotRoot(pwd)

	//mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Errorf("error %v", err)
	}

	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME|syscall.MS_PRIVATE|syscall.MS_REC, "mode=755")
}

func pivotRoot(root string) error {
	/**
	  为了使当前root的老 root 和新 root 不在同一个文件系统下，我们把root重新mount了一次
	  bind mount是把相同的内容换了一个挂载点的挂载方法
	*/
	//syscall.Unshare(syscall.CLONE_NEWNS)
	/*
		如果不重新挂载一下根目录，会出现无法pivot_root的情况，这是因为默认systemd把根目录以Shared的方式挂载
	*/
	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		print("root bind error")
		fmt.Print(err)
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		print("bind error")
		fmt.Print(err)
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}
	// 创建 rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		fmt.Print(err)
		return err
	}
	// pivot_root 到新的rootfs, 现在老的 old_root 是挂载在rootfs/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	//print("path" + pivotDir)
	//print("root" + root)
	//syscall.Chroot(root)
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		print("failed")
		fmt.Print(err)
		return fmt.Errorf("pivot_root %v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		fmt.Print(err)
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	// 删除临时文件夹
	//defer os.Remove(pivotDir)
	return os.Remove(pivotDir)
	//return nil
}
