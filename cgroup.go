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
// 用于传递资源限制配置的结构体，包含内存限制，CPU时间片权重，CPU核心数
// 请注意这些控制功能可能并不是每台机器都有，请酌情选取
type ResourceConfig struct {
	// CPU Limit.
	cpuShare  string
	cpuSets string
	cpuMax string
	// Memory Limit.
	memoryLimitBytes string
	memoryMax string
	// blkio Limit.
	blkioReadBps string
	blkioWriteBps string
	// net work width limit.
	netClsClassid string
	netPrioIfpriomap string
}

func NewDefaultResourceConfig() ResourceConfig {
	return ResourceConfig{
		cpuShare:  "",
		cpuSets: "",
		cpuMax: "",
		// Memory Limit.
		memoryLimitBytes: "",
		memoryMax: "",
		// blkio Limit.
		blkioReadBps: "",
		blkioWriteBps: "",
		// net work width limit.
		netClsClassid: "",
		netPrioIfpriomap: "",
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
	// ----------------------------------------------------------------
	// author: linke song.
	// cpu: cpu.shares, cpu.cfs_quota_us, cpu.cfs_period_us, cpu.max
	// ----------------------------------------------------------------
	// cpu, memory, blkio, network, device, pids, filesystem
	// ----------------------------------------------------------------
	// Resources:
	// ----------------------------------------------------------------
	// CPU:
	// cpuShare  string
	// cpuSets string
	// cpuMax string
	// Memory:
	// Memory Limit.
	// memoryLimitBytes string
	// memoryMax string
	// Blkio:
	// blkio Limit.
	// blkioReadBps string
	// blkioWriteBps string
	// // net work width limit.
	// netClsClassid string
	// netPrioIfpriomap string
	// 
	// -----------------------------------------------
	// 1. cpu.shares. -> 表示进程所分配到的最大CPU份额
	// fmt.Print(s.path)
	if res.cpuShare != "" {
		if err := ioutil.WriteFile(path.Join(s.path, "cpu.shares"), []byte(res.cpuShare), 0644); err != nil {
			fmt.Print(err)
			return fmt.Errorf("set cgroup v2 cpu share failed")
		}
	}
	// 2. cpuset.cpus. -> 指定一个任务可以运行在哪些CPU核心上
	if res.cpuSets != "" {
		if err := ioutil.WriteFile(path.Join(s.path, "cpuset.cpus"), []byte(res.cpuSets), 0644); err != nil {
			fmt.Print(err)
			return fmt.Errorf("set cgroup v2 cpu set failed")
		}
	}
	// 3. cpu.max. -> 指定一个任务最大CPU限制
	if res.cpuMax != "" {
		if err := ioutil.WriteFile(path.Join(s.path, "cpu.max"), []byte(res.cpuMax), 0644); err != nil {
			fmt.Print(err)
			return fmt.Errorf("set cgroup v2 cpu max failed")
		}
	}
	// 4. memoryLimitBytes. -> 与memory max的不同在于前者可以继承，后者可以单独设立 
	if res.memoryLimitBytes != "" {
		if err := ioutil.WriteFile(path.Join(s.path, "memory.limit_in_bytes"), []byte(res.memoryLimitBytes), 0644); err != nil {
			fmt.Print(err)
			return fmt.Errorf("set cgroup v2 memory limit failed")
		}
	}
	// 5. memoryMax. -> 内存最大分配值
	if res.memoryMax != "" {
		if err := ioutil.WriteFile(path.Join(s.path, "memory.max"), []byte(res.memoryMax), 0644); err != nil {
			fmt.Print(err)
			return fmt.Errorf("set cgroup v2 memory max failed")
		}
	}
	// 6. blkioReadBps. -> 块设备读带宽
	if res.blkioReadBps != "" {
		if err := ioutil.WriteFile(path.Join(s.path, "blkio.throttle.read_bps_device"), []byte(res.blkioReadBps), 0644); err != nil {
			fmt.Print(err)
			return fmt.Errorf("set cgroup v2 blk readbps failed")
		}
	}
	// 7. blkioWriteBps. -> 块设备写带宽
	if res.blkioWriteBps != "" {
		if err := ioutil.WriteFile(path.Join(s.path, "blkio.throttle.write_bps_device"), []byte(res.blkioWriteBps), 0644); err != nil {
			fmt.Print(err)
			return fmt.Errorf("set cgroup v2 blk writebps failed")
		}
	}
	// 8. netClsClassid. -> 网络传输带宽限制
	if res.netClsClassid != "" {
		if err := ioutil.WriteFile(path.Join(s.path, "net_cls.classid"), []byte(res.netClsClassid), 0644); err != nil {
			fmt.Print(err)
			return fmt.Errorf("set cgroup v2 net_cls classid failed")
		}
	}
	// 9. netPrioIfpriomap. -> 网络接口传输带宽限制
	if res.netPrioIfpriomap != "" {
		if err := ioutil.WriteFile(path.Join(s.path, "net_prio.ifpriomap"), []byte(res.netPrioIfpriomap), 0644); err != nil {
			fmt.Print(err)
			return fmt.Errorf("set cgroup v2 net_prio ifpriomap failed")
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
