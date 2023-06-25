os=centos
docker run -d --name tmp_container_$os $os sh
docker export -o busybox.tar tmp_container_$os
mv busybox.tar ./build/
docker stop tmp_container_$os && docker rm tmp_container_$os 
