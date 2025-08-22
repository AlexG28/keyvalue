# Distributed Key value store 

### Setup Instructions 

build the container with `docker build -t key-value-store .` and then run it with `docker run -p 8080:2222 key-value-store`


## Instructions for running a single instance in local K8s cluster (on ARM mac): 
- install Kubectl, Minikubes, Qemu and socket_vmnet
- start up custom network `sudo brew services start socket_vmnet`
- start up minikube within qemu virtual machine `minikube start driver=qemu2` (this will automatically use the socket_vmnet network)
- run `eval $(minikube -p minikube docker-env)` to connect minikubes to your docker 
- build docker image `docker build -t test_key_value:latest .`
- deploy `kubectl apply -f deployment.yaml`
- test deployment `kubectl get pods`. It should say 'running' under status. 
- get url of pod: `minikube service kvstore-service --url`. This is the URL to which to send Get and Set 
- so for example `curl "http://192.168.105.3:32437/Set/hello/world"`. But the IP address and port will be different (assigned by k8s)


