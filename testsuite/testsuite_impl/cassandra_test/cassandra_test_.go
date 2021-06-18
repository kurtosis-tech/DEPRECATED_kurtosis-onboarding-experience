package cassandra_test

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/galenmarchetti/kurtosis-onboarding-test/testsuite/services_impl/cassandra_service"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	cassandraServiceID services.ServiceID = "cassandra-1"

	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxPolls = 15
)

type CassandraTest struct {
	CassandraServiceImage string
}

func NewCassandraTest(image string) *CassandraTest {
	return &CassandraTest{CassandraServiceImage: image}
}

func (test CassandraTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(30).WithRunTimeoutSeconds(30)
}

func (test CassandraTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	logrus.Infof("Setting up cassandra test.")
	/*
		NEW USER ONBOARDING:
		- Add services multiple times using the below logic in order to have more than one service.
	*/
	configFactory := cassandra_service.NewCassandraServiceConfigFactory(test.CassandraServiceImage)
	_, hostPortBindings, availabilityChecker, err := networkCtx.AddService(cassandraServiceID, configFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the service")
	}
	if err := availabilityChecker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxPolls); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the service to become available")
	}
	logrus.Infof("Added service with host port bindings: %+v", hostPortBindings)
	return networkCtx, nil
}

func (test CassandraTest) Run(uncastedNetwork networks.Network) error {
	logrus.Infof("Running cassandra test.")
	// Necessary because Go doesn't have generics
	castedNetwork := uncastedNetwork.(*networks.NetworkContext)

	uncastedService, err := castedNetwork.GetService(cassandraServiceID)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the datastore service")
	}

	// Necessary again due to no Go generics
	castedService := uncastedService.(*cassandra_service.CassandraService)
	logrus.Infof("Service is available: %v", castedService.IsAvailable())

	time.Sleep(10 * time.Second)
	// Define object used to represent local Cassandra cluster
	cluster := gocql.NewCluster(castedService.GetIPAddress())
	cluster.Consistency = gocql.One
	cluster.ProtoVersion = 4

	// Define object used to send queries to local Cassandra cluster
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Create a keyspace "test" to use for testing purposes
	err = session.Query(`CREATE KEYSPACE test WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 }`).Exec()
	if err != nil {
		panic(err)
	}

	// Create a table "tweet" to use for testing purposes
	err = session.Query(`CREATE TABLE test.tweet(timeline text, id timeuuid PRIMARY KEY, text text)`).Exec()
	if err != nil {
		panic(err)
	}

	// Insert a tweet into "tweet" table
	err = session.Query(`INSERT INTO test.tweet (timeline, id, text) VALUES (?, ?, ?)`, "me", gocql.TimeUUID(), "hello world").Exec()
	if err != nil {
		panic(err)
	}

	// Read a tweet from "tweet" table
	var (
		id   gocql.UUID
		text string
	)
	err = session.Query(`SELECT id, text FROM test.tweet WHERE timeline = ? LIMIT 1 ALLOW FILTERING`, "me").Scan(&id, &text)
	if err != nil {
		panic(err)
	}

	// Print the contents of the tweet
	fmt.Printf("Tweet: %v, %v\n", id, text)
	// Print results of test condition
	fmt.Printf("Test passed?: %v\n", text == "hello world")

	return nil
}