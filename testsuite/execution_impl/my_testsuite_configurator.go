package execution_impl

import (
	"github.com/galenmarchetti/kurtosis-onboarding-test/testsuite/testsuite_impl"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type MyTestsuiteConfigurator struct {}

func NewMyTestsuiteConfigurator() *MyTestsuiteConfigurator {
	return &MyTestsuiteConfigurator{}
}

func (t MyTestsuiteConfigurator) SetLogLevel(logLevelStr string) error {
	level, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing loglevel string '%v'", logLevelStr)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	return nil
}

func (t MyTestsuiteConfigurator) ParseParamsAndCreateSuite(paramsJsonStr string) (testsuite.TestSuite, error) {
	suite := &testsuite_impl.MyTestsuite{}
	return suite, nil
}