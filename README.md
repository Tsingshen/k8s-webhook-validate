# k8s-webhook-validate

## requirement cert-manager

## steps:
安装ca
```
use cert-manager generate tls
生成ca时要使用一个ca-tls
openssl genrsa -out ca.key 2048
openssl req -x509 --newkey rsa:2048 -new -nodes -key ca.key -days 3650 -reqexts v3_req -extensions v3_ca -out ca.crt
kubectl create secert tls ca-key-pair --cert=ca.crt --key=ca.key --namespace=default
kubectl apply -f install-on-k8s/issuer.yaml
```
生成证书
```
cd ./install-on-k8s
kubectl apply -f issuer.yaml
kubectl apply -f cert.yaml

kubectl get secert deploy-validate-tls -oyaml
```
- 使用上面生成的 tls secert.deploy-validate-tls ca.crt 做为 validat-webhook.yaml caBundle 的值
- 将 secret.deploy-validate-tls tls.crt,tls,key 分别 base64 反解，放在config/tls/ 目录下
