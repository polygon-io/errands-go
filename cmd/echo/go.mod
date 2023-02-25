module github.com/polygon-io/errands-go/echo

go 1.16

require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/polygon-io/errands-go v0.0.6
	github.com/polygon-io/errands-server v1.1.0
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/sys v0.1.0 // indirect
)

// Use the local copy of errands
replace github.com/polygon-io/errands-go => ../../
