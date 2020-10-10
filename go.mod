module github.com/sammyne/k8s-playground

go 1.15

require (
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v0.19.0
)

replace k8s.io/client-go => github.com/kubernetes/client-go v0.19.0

replace k8s.io/apimachinery => github.com/kubernetes/apimachinery v0.19.0
