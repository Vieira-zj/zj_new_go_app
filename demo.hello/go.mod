module demo.hello

go 1.16

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/emicklei/proto v1.9.0
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-xorm/xorm v0.7.9
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.5.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jmoiron/sqlx v1.3.3
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.9.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	go.etcd.io/etcd v3.3.25+incompatible
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20210403161142-5e06dd20ab57 // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	golang.org/x/tools v0.1.0
	google.golang.org/genproto v0.0.0-20210406143921-e86de6bf7a46 // indirect
	google.golang.org/grpc v1.37.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.19.0
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v0.19.0
	k8s.io/utils v0.0.0-20210305010621-2afb4311ab10 // indirect
)

replace google.golang.org/grpc v1.37.0 => google.golang.org/grpc v1.26.0
