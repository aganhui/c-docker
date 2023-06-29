/*
*未来要更新的目标，设置新的全局配置
 */
package main

// 全局使用的变量
var (
	// MntUrl = "/root/mnt/%s" // 包含镜像的位置
	MntUrl = "./mnt/%s" // 包含镜像的位置
	// RootUrl = "/root"        // 挂载根目录的位置
	RootUrl      = "./image"        // image repo
	ImageUrl     = "./image/%s"     // image repo
	ContainerUrl = "./container/%s" // 挂载根目录的位置
	WorkUrl      = "./.worker/%s"   // overlay工作文件夹
	// WriteLayerUrl = "/root/writeLayer/%s" // 可写层位置
	WriteLayerUrl = "./writeLayer/%s" // 可写层位置
)

// 下列两个变量存放在globalDefaultInfoLocation指向的目录里
var (
	globalLogName    = "log.txt"     // 日志的规范文件名
	globalConfigName = "config.json" // 配置信息的规范文件名
)

const globalDefaultInfoLocation = "./config/%s/" // 保存和容器有关信息的模式字符串，实际容器信息保存在config/containerName李里

// 下列涉及的文件实际位置在globalRootURl指向的目录里
var (
	imageTarName = "busybox.tar" // 镜像文件的名称
	imageName    = "busybox"     // 镜像文件解压的目录名
)

var globalExeLocation = ""
