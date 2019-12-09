module github.com/applinskinner/redsync

go 1.13

require (
	github.com/go-redis/redis v6.15.6+incompatible
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/hashicorp/go-multierror v1.0.0
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/stvp/tempredis v0.0.0-20181119212430-b82af8480203
)

// TODO: Remove this once this issue is addressed, or redigo no longer points to
//       v2.0.0+incompatible, above: https://github.com/gomodule/redigo/issues/366
replace github.com/gomodule/redigo => github.com/gomodule/redigo v1.7.0
