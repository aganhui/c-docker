os=centos
docker run -d --name tmp_container_$os $os sh
docker export -o busybox.tar tmp_container_$os
mkdir -p ./build/image/
mv busybox.tar ./build/image/
docker stop tmp_container_$os && docker rm tmp_container_$os 
