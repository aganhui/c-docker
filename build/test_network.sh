#!/bin/bash

exe=c-docker
net=net

# 获取命令行参数
arg="$1"

# 根据参数执行不同的操作
if [ "$arg" == "create" ]; then
  echo "create net..."
  sudo ./$exe network create $net --subnet 192.168.0.0/24 --driver bridge
  sudo ./$exe network list
elif [ "$arg" == "run1" ]; then
  echo "run container 1 .."
  sudo ./$exe run -ti -p 80:80 --net $net /bin/bash
  # ip addr
elif [ "$arg" == "run2" ]; then
  echo "run container 2 .."
  # 创建多个容器，测试容器之间的网络互通性
  sudo ./$exe run -ti -p 81:81 --net $net /bin/bash
  # ping <ip>
elif [ "$arg" == "remove" ]; then
  echo "remove net..."
  sudo ./$exe network remove $net
  sudo ./$exe network list
else
  echo "Usage: $0 {create|run1|run2|remove}"
fi




