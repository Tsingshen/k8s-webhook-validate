# k8s-webhook-validate

## requirement cert-manager

## steps:
```
cd ./install-on-k8s
kubectl apply -f issuer.yaml
kubectl apply -f cert.yaml

kubectl get secert deploy-validate-tls -oyaml
```
- 使用上面生成的 tls secert.deploy-validate-tls ca.crt 做为 validat-webhook.yaml caBundle 的值
- 将 secret.deploy-validate-tls tls.crt,tls,key 分别 base64 反解，放在config/tls/ 目录下
