package network

const globalDefaultNetwork = "./network"

type ContainerInfo struct {
	Pid         string   `json:"pid"`         // 容器init进程在宿主机上的PID
	Id          string   `json:"id"`          // 容器id
	Name        string   `json:"name"`        // 容器名
	Command     string   `json:"command"`     // 容器内init进程的运行命令
	CreatedTime string   `json:"createTime"`  // 创建时间
	Status      string   `json:"status"`      // 容器的状态
	Volume      string   `json:"volume"`      // 容器的数据卷
	PortMapping []string `json:"portmapping"` // 端口映射
}
