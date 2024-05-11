package testsuite

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestRunTestSuite(t *testing.T) {
	s := new(ExampleTestSuite)
	s.caseRunTime = make(map[string]int64, 2)
	suite.Run(t, s)
}

type ExampleTestSuite struct {
	suite.Suite
	varShoulBeFive int
	caseRunTime    map[string]int64
}

// Test Hooks

func (suite *ExampleTestSuite) SetupSuite() {
	suite.T().Log("SetupSuite")
}

func (suite *ExampleTestSuite) TearDownSuite() {
	suite.T().Log("TearDownSuite")
	for tc, rtime := range suite.caseRunTime {
		suite.T().Logf("case=%s, run_time=%d(millisecs)", tc, rtime)
	}
}

func (suite *ExampleTestSuite) SetupTest() {
	suite.T().Log("SetupTest")
	suite.varShoulBeFive = 5
}

func (suite *ExampleTestSuite) TearDownTest() {
	suite.T().Log("TearDownTest")
}

func (suite *ExampleTestSuite) BeforeTest(suiteName, testName string) {
	suite.T().Logf("BeforeTest: suite=%s, case=%s", suiteName, testName)
	suite.caseRunTime[getCaseFullName(suiteName, testName)] = time.Now().UnixMilli()
}

func (suite *ExampleTestSuite) AfterTest(suiteName, testName string) {
	suite.T().Logf("AfterTest: suite=%s, case=%s", suiteName, testName)
	start := suite.caseRunTime[getCaseFullName(suiteName, testName)]
	suite.caseRunTime[getCaseFullName(suiteName, testName)] = time.Since(time.UnixMilli(start)).Milliseconds()
}

// TestCases

func (suite *ExampleTestSuite) TestCase01() {
	suite.T().Log("TestCase01")
	time.Sleep(50 * time.Millisecond)
	suite.Equal(suite.varShoulBeFive, 5, "int not equal")
	suite.T().Log("TestCase01 done")
}

func (suite *ExampleTestSuite) TestCase02() {
	suite.T().Log("TestCase02")
	time.Sleep(30 * time.Millisecond)
	suite.Equal(suite.varShoulBeFive, 4, "int not equal")
	suite.T().Log("TestCase02 done")
}

func (suite *ExampleTestSuite) TestCase03() {
	suite.T().Log("TestCase03")
	time.Sleep(100 * time.Millisecond)
	r := suite.Require()
	// the first requirement which fails interrupts and fails the complete test
	r.Equal(suite.varShoulBeFive, 4, "int not equal, with failnow")
	suite.T().Log("TestCase03 done")
}

func (suite *ExampleTestSuite) TestCaseSkip() {
	suite.T().Skip("case to be skip")
	time.Sleep(20 * time.Millisecond)
	suite.T().Log("TestSkip")
}

func (suite *ExampleTestSuite) TestCaseParallel() {
	for i := 0; i < 5; i++ {
		idx := strconv.Itoa(i)
		suite.T().Run("parallel case"+idx, func(t *testing.T) {
			t.Parallel()
			t.Log("run parallel case" + idx)
			time.Sleep(10 * time.Millisecond)
		})
	}
}

func getCaseFullName(suiteName, testName string) string {
	return suiteName + "/" + testName
}
