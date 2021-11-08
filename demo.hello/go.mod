module demo.hello

go 1.16

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/avast/retry-go/v3 v3.1.1
	github.com/aws/aws-sdk-go v1.38.51
	github.com/bep/debounce v1.2.0
	github.com/docker/docker v20.10.2+incompatible
	github.com/emicklei/proto v1.9.0
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/fatih/color v1.7.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gin-gonic/gin v1.7.4
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-xorm/xorm v0.7.9
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.5.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/jmoiron/sqlx v1.3.3
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.9.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/fx v1.14.2
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	golang.org/x/tools v0.1.2
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.1.3
	gorm.io/driver/sqlite v1.2.3
	gorm.io/gorm v1.22.2
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/kubectl v0.0.0
	k8s.io/kubernetes v1.22.1
	sigs.k8s.io/controller-runtime v0.10.1
)

replace (
	// google.golang.org/grpc v1.37.0 => google.golang.org/grpc v1.26.0
	k8s.io/api v0.0.0 => k8s.io/api v0.22.1
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.22.1
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.22.1
	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.22.1
	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.22.1
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.22.1
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.22.1
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.22.1
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.22.1
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.22.1
	k8s.io/component-helpers v0.0.0 => k8s.io/component-helpers v0.22.1
	k8s.io/controller-manager v0.0.0 => k8s.io/controller-manager v0.22.1
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.22.1
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.22.1
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.22.1
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.22.1
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.22.1
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.22.1
	k8s.io/kubectl v0.0.0 => k8s.io/kubectl v0.22.1
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.22.1
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.22.1
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.22.1
	k8s.io/mount-utils v0.0.0 => k8s.io/mount-utils v0.22.1
	k8s.io/pod-security-admission v0.0.0 => k8s.io/pod-security-admission v0.22.1
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.22.1
)
