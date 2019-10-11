package purecloud_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-logger"
	"github.com/stretchr/testify/suite"

	purecloud "github.com/gildas/go-purecloud"
)

type OrganizationSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time

	Client *purecloud.Client
}

func TestOrganizationSuite(t *testing.T) {
	suite.Run(t, new(OrganizationSuite))
}

func (suite *OrganizationSuite) TestOrganizationHasName() {
	suite.Assert().NotNil(suite.Client.Organization)
	suite.Assert().NotEmpty(suite.Client.Organization.Name)
	suite.Assert().NotEmpty(suite.Client.Organization.Name)
	suite.Assert().NotEmpty(suite.Client.Organization.Name)
	suite.Logger.Record("org", suite.Client.Organization).Infof("Organization Details")
	suite.T().Logf("Organization: %s", suite.Client.Organization.Name)
}

// Suite Tools

func (suite *OrganizationSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	logFolder := filepath.Join(".", "log")
	os.MkdirAll(logFolder, os.ModePerm)
	suite.Logger = logger.CreateWithDestination("test", fmt.Sprintf("file://%s/test-%s.log", logFolder, strings.ToLower(suite.Name)))
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))

	var (
		region       = core.GetEnvAsString("PURECLOUD_REGION", "")
		clientID     = core.GetEnvAsString("PURECLOUD_CLIENTID", "")
		secret       = core.GetEnvAsString("PURECLOUD_CLIENTSECRET", "")
		deploymentID = core.GetEnvAsString("PURECLOUD_DEPLOYMENTID", "")
	)

	suite.Client = purecloud.New(purecloud.ClientOptions{
		Region:       region,
		DeploymentID: deploymentID,
		ClientID:     clientID,
		ClientSecret: secret,
		Logger:       suite.Logger,
	})
	suite.Require().NotNil(suite.Client, "PureCloud Client is nil")
}

func (suite *OrganizationSuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
}

func (suite *OrganizationSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()

	// Reuse tokens as much as we can
	if !suite.Client.IsAuthorized() {
		err := suite.Client.Login()
		suite.Require().Nil(err, "Failed to login")
	}
}

func (suite *OrganizationSuite) AfterTest(suiteName, testName string) {
	duration := time.Now().Sub(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}