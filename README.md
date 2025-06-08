# mushDb
Add a proper description

## Make sure your main node has the right tags
microk8s start
kubectl get nodes --show-labels
kubectl label nodes <node-name> mushroomNode=main
run swapoff -a
192.168.65.2
# If you want pods to be able to run on the main node:
kubectl taint node myMasterNodeName node-role.kubernetes.io/control-plane:NoSchedule-
where myMasterNodeName is from: kubectl get nodes -o wide
If that does not work, switch "control-plane" to "master" for older k8s
# Helpful commands
kubectl get nodes -o wide
kubectl get pods --all-namespaces
kubectl describe pods
helm install -f noCommit-testvalues.yaml test-chart fungi-tracker
helm uninstall test-chart
sh scripts/build.sh api
docker save mush-api > mush-api.tar
multipass transfer mush-api.tar microk8s-vm:/tmp/mush-api.tar
microk8s ctr image import /tmp/mush-api.tar
kubectl get pod,pvc
multipass mount <local path> <instance name>:<instance path>
multipass mount /tmp/testHost microk8s-vm:/tmp/testHost
multipass mount /tmp/testHost microk8s-vm:/tmp/testHost
multipass mount /tmp/testHost microk8s-vm:/tmp/testHostm
multipass exec microk8s-vm -- sudo ls /tmp/mush/testdb
multipass exec microk8s-vm -- sudo mkdir /tmp/mush/testdb
kubectl logs --namespace cloudflare-ddns cloudflare-ddns-58dfd9b747-xw9h9
# TESTING TEMPLATES
helm install -f homelab-ddns/noCommit-values.yaml test-chart homelab-ddns --dry-run --debug
helm template -s templates/deployment.yaml
# TODO:
mongosh into the db to setup users
