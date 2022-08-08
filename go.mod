module github.com/nemoremold/perftests

go 1.16

require (
	github.com/chaos-mesh/chaos-mesh/api/v1alpha1 v0.0.0-20220226050744-799408773657
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/prometheus/client_golang v1.13.0
	github.com/prometheus/client_model v0.2.0
	github.com/spf13/pflag v1.0.5
	go.etcd.io/etcd/client/v3 v3.5.1
	golang.org/x/net v0.0.0-20220805013720-a33c5aa5df48
	golang.org/x/sys v0.0.0-20220804214406-8e32c043e418 // indirect
	google.golang.org/genproto v0.0.0-20220805133916-01dd62135a58 // indirect
	google.golang.org/grpc v1.48.0
	k8s.io/api v0.24.0
	k8s.io/apimachinery v0.24.0
	k8s.io/client-go v0.24.0
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.60.1
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	sigs.k8s.io/controller-runtime v0.12.0
)
