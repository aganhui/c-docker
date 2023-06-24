/*
*未来要更新的目标，设置新的全局配置
 */
package main

//全局使用的变量
var globalMntURL = "./mnt"          //挂载根目录的位置
var globalRootURL = "./"            //包含镜像的位置
var globalCommitURL = globalRootURL //提交保存镜像的位置

//下列两个变量存放在globalDefaultInfoLocation指向的目录里
var globalLogName = "log.txt"        //日志的规范文件名
var globalConfigName = "config.json" //配置信息的规范文件名

const globalDefaultInfoLocation = "./config/%s/" //保存和容器有关信息的模式字符串，实际容器信息保存在config/containerName李里

//下列涉及的文件实际位置在globalRootURl指向的目录里
var imageTarName = "busybox.tar" //镜像文件的名称
var imageName = "busybox"        //镜像文件解压的目录名
