module github.com/sammyne/k8s-playground/annotations

go 1.15

require (
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.19.0
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v0.19.0
)

replace k8s.io/client-go => github.com/kubernetes/client-go v0.17.9

replace k8s.io/apimachinery => github.com/kubernetes/apimachinery v0.17.9

replace k8s.io/api => github.com/kubernetes/api v0.17.9
