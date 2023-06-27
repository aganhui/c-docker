package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

/***********
当前设计只支持cgroupv2，对cgroupv1暂不支持
***********/
//用于传递资源限制配置的结构体，包含内存限制，CPU时间片权重，CPU核心数
type ResourceConfig struct {
	memoryMax string
	cpuShare  string
}

func NewDefaultResourceConfig() ResourceConfig {
	return ResourceConfig{
		memoryMax: "",
		cpuShare:  "",
	}
}

func FindCgroupV2MountPoint() (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", fmt.Errorf("Failed to open mountinfo.")

	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		if strings.Contains(txt, "cgroup2") {
			//print("field", fields[4])
			return fields[4], nil
			//return "", fmt.Errorf("Currently only cgroupv2 is supported, but we didn't find it.")
		}
	}
	return "", fmt.Errorf("Currently only cgroupv2 is supported, but we didn't find it.")
}

//得到cgroup在文件系统中的绝对位置
func GetCgroupV2Path(cgrouppath string, autoCreate bool) (string, error) {
	if cgroupRoot, err := FindCgroupV2MountPoint(); err == nil {
		if _, err := os.Stat(path.Join(cgroupRoot, cgrouppath)); err == nil || (autoCreate && os.IsNotExist(err)) {
			if os.IsNotExist(err) {
				if err := os.Mkdir(path.Join(cgroupRoot, cgrouppath), 0755); err == nil {
					return path.Join(cgroupRoot, cgrouppath), nil
				} else {
					fmt.Print(err)
					return "", fmt.Errorf("error create cgroup %v", err)
				}
			} else {
				return path.Join(cgroupRoot, cgrouppath), nil
			}
		}
	} else {
		fmt.Print(err)
		return "", fmt.Errorf("cgroup path error %v", err)
	}
	return "", fmt.Errorf("cgroup path error\n")
}

type CgroupManager interface {
	Version() uint
	Set(res *ResourceConfig) error
	Apply(pid int) error
	Destory() error
}

type CgroupV2Manager struct {
	path         string //这里会把创建的subcgroup路径保存起来
	relativepath string //相对位置，在set的时候需要用到
}

func getDefaultCgroupV2Manager() CgroupV2Manager {
	return CgroupV2Manager{
		path:         "",
		relativepath: "",
	}
}

func (s *CgroupV2Manager) Version() uint {
	return 2
}

func (s *CgroupV2Manager) Set(res *ResourceConfig) error {
	if s.path == "" {
		//print(s.relativepath)
		if path, err := GetCgroupV2Path(s.relativepath, true); err == nil {
			s.path = path //如果绝对路径为空，就设置一个绝对路径
			//print(path)
		} else {
			fmt.Print(err)
			return fmt.Errorf("failed to get cgroup v2 path", err)
		}
	}
	//print("max" + res.memoryMax)
	if res.memoryMax != "" {
		//print("set" + s.path)
		if err := ioutil.WriteFile(path.Join(s.path, "memory.max"), []byte(res.memoryMax), 0644); err != nil {
			fmt.Print(err)
			return fmt.Errorf("set cgroup v2 memory max failed")
		}

	}
	return nil
}

func (s *CgroupV2Manager) Destory() error {
	//print(s.path)
	if s.path != "" {
		return os.Remove(s.path)
	}
	return nil
}

func (s *CgroupV2Manager) Apply(pid int) error {
	if s.path != "" {
		if err := ioutil.WriteFile(path.Join(s.path, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)

		} else {
			return nil
		}
	} else {
		return fmt.Errorf("subcgroup path is empty!")
	}
	return nil
}
