# performance-testing-operator-demo

The Performance Testing Operator is a demo for the Performance Test Interface Spec. It's purpose is to server as an example for other performance testing vendor implementations. It is currently build upon k6.



## Demo

Clone Demo application repo
```
git clone https://github.com/performancetestinterface/podinfo.git
```

Deploy demo application
```
kubectl apply -k podinfo/deploy/bases/frontend
```

Clone Operator repo
```
https://github.com/performancetestinterface/performance-testing-operator.git
```

Install CRDs
```
kubectl apply -f performance-testing-operator/config/crd/bases
```

Install RBAC
```
kubectl apply -k performance-testing-operator/config/rbac
```

Install Operator
```
kubectl apply -k performance-testing-operator/config/manager
```

Install samples
```
kubectl apply -f performance-testing-operator/config/samples
```


