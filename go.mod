module github.com/example/observatory-operator

go 1.21

require (
    sigs.k8s.io/controller-runtime v0.16.3
    k8s.io/api v0.28.4
    k8s.io/apimachinery v0.28.4
    k8s.io/client-go v0.28.4
    github.com/onsi/gomega v1.27.10
)

replace k8s.io/api => k8s.io/api v0.28.4
replace k8s.io/apimachinery => k8s.io/apimachinery v0.28.4
replace k8s.io/client-go => k8s.io/client-go v0.28.4
