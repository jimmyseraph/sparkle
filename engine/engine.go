package engine

import "fmt"

type TestFeature struct {
	Name       string
	BeforeAll  func(assertion *Assertion)
	AfterAll   func(assertion *Assertion)
	BeforeEach func(assertion *Assertion)
	AfterEach  func(assertion *Assertion)
	TestCases  []TestCase
}

type TestCase struct {
	Name         string
	Tag          []string
	Ignore       bool
	Parameterize func() [][]interface{}
	Case         func(assertion *Assertion, args ...interface{})
}

const (
	FEATURE_START = "Feature Start"
	FEATURE_END   = "Feature End"
	CASE_START    = "Case Start"
	CASE_END      = "Case End"
	STEP          = "Step"
	RESULT        = "Result"
	ASSERT        = "Assert"
)

func (t *TestFeature) RunFeature(parent *Assertion, logger Logger, c chan *Assertion, tags ...string) {
	node := NewAssertion(t.Name, TEST_FEATURE, parent, logger)
	logger.Log(FEATURE_START, "Start running feature %s", t.Name)
	if t.BeforeAll != nil {
		logger.Log(STEP, "Running BeforeAll")
		t.BeforeAll(node)
	}
	if x := recover(); x != nil {
		node.fail()
		logger.Log(RESULT, "BeforeAll failed on feature %s", t.Name)
		return
	}

	for _, testCase := range t.TestCases {
		if !testCase.Ignore && testCase.Case != nil {
			if tags != nil {
				isMatch := false
				for _, tag := range tags {
					if isMatch {
						break
					}
					for _, caseTag := range testCase.Tag {
						if tag == caseTag {
							isMatch = true
							break
						}
					}
				}
				if !isMatch {
					testNode := NewAssertion(testCase.Name, TEST_CASE, node, logger)
					testNode.result = IGNORE
					continue
				}
			}

			if testCase.Parameterize != nil {
				parameters := testCase.Parameterize()
				for i, param := range parameters {
					caseName := fmt.Sprintf("%s[%d]", testCase.Name, i)
					testNode := NewAssertion(caseName, TEST_CASE, node, logger)

					if t.BeforeEach != nil {
						testNode.AddDetail(STEP, "Running BeforeEach before testcase %s", caseName)
						t.BeforeEach(testNode)
						if x := recover(); x != nil || testNode.result == FAIL {
							testNode.fail()
							testNode.AddDetail(RESULT, "BeforeEach failed before testcase %s", caseName)
							c <- testNode
							continue
						}
					}

					testCase.runCase(caseName, testNode, param...)
					if x := recover(); x != nil {
						testNode.fail()
						testNode.AddDetail(RESULT, "testcase %s failed", caseName)
					}

					if t.AfterEach != nil {
						testNode.AddDetail(STEP, "Running AfterEach before testcase %s", caseName)
						t.AfterEach(testNode)
						if x := recover(); x != nil || testNode.result == FAIL {
							testNode.fail()
							testNode.AddDetail(RESULT, "AfterEach failed before testcase %s", caseName)
							c <- testNode
							continue
						}
					}
					c <- testNode
				}
			} else {
				testNode := NewAssertion(testCase.Name, TEST_CASE, node, logger)
				if t.BeforeEach != nil {
					testNode.AddDetail(STEP, "Running BeforeEach before testcase %s", testCase.Name)
					t.BeforeEach(testNode)
					if x := recover(); x != nil || testNode.result == FAIL {
						testNode.fail()
						testNode.AddDetail(RESULT, "BeforeEach failed before testcase %s", testCase.Name)
						c <- testNode
						continue
					}
				}

				testCase.runCase(testCase.Name, testNode)
				if x := recover(); x != nil {
					testNode.fail()
					testNode.AddDetail(RESULT, "testcase %s failed", testCase.Name)
				}

				if t.AfterEach != nil {
					testNode.AddDetail(STEP, "Running AfterEach before testcase %s", testCase.Name)
					t.AfterEach(testNode)
					if x := recover(); x != nil || testNode.result == FAIL {
						testNode.fail()
						testNode.AddDetail(RESULT, "AfterEach failed before testcase %s", testCase.Name)
						c <- testNode
						continue
					}
				}
				c <- testNode
			}
		} else {
			testNode := NewAssertion(testCase.Name, TEST_CASE, node, logger)
			testNode.result = IGNORE
			c <- testNode
		}
	}
	if t.AfterAll != nil {
		logger.Log(STEP, "Running AfterAll")
		t.AfterAll(node)
	}
	if x := recover(); x != nil {
		node.fail()
		logger.Log(RESULT, "AfterAll failed on feature %s", t.Name)
	}
	defer logger.Log(FEATURE_END, "End running feature %s", t.Name)
}

func (t *TestFeature) RunTestCase(testCases []*TestCase, parent *Assertion, c chan *Assertion, logger Logger) {

	node := NewAssertion(t.Name, TEST_FEATURE, parent, logger)
	logger.Log(FEATURE_START, "Start running feature %s", t.Name)
	if t.BeforeAll != nil {
		logger.Log(STEP, "Running BeforeAll")
		t.BeforeAll(node)
	}
	if x := recover(); x != nil {
		node.fail()
		logger.Log(RESULT, "BeforeAll failed on feature %s", t.Name)
		return
	}

	for _, testCase := range testCases {
		if !testCase.Ignore && testCase.Case != nil {
			if testCase.Parameterize != nil {
				parameters := testCase.Parameterize()
				for i, param := range parameters {
					caseName := fmt.Sprintf("%s[%d]", testCase.Name, i)
					testNode := NewAssertion(caseName, TEST_CASE, node, logger)

					if t.BeforeEach != nil {
						testNode.AddDetail(STEP, "Running BeforeEach before testcase %s", caseName)
						t.BeforeEach(testNode)
						if x := recover(); x != nil || testNode.result == FAIL {
							testNode.fail()
							testNode.AddDetail(RESULT, "BeforeEach failed before testcase %s", caseName)
							c <- testNode
							continue
						}
					}

					testCase.runCase(caseName, testNode, param...)
					if x := recover(); x != nil {
						testNode.fail()
						testNode.AddDetail(RESULT, "testcase %s failed", caseName)
					}

					if t.AfterEach != nil {
						testNode.AddDetail(STEP, "Running AfterEach before testcase %s", caseName)
						t.AfterEach(testNode)
						if x := recover(); x != nil || testNode.result == FAIL {
							testNode.fail()
							testNode.AddDetail(RESULT, "AfterEach failed before testcase %s", caseName)
							c <- testNode
							continue
						}
					}
					c <- testNode
				}
			} else {
				testNode := NewAssertion(testCase.Name, TEST_CASE, node, logger)
				if t.BeforeEach != nil {
					testNode.AddDetail(STEP, "Running BeforeEach before testcase %s", testCase.Name)
					t.BeforeEach(testNode)
					if x := recover(); x != nil || testNode.result == FAIL {
						testNode.fail()
						testNode.AddDetail(RESULT, "BeforeEach failed before testcase %s", testCase.Name)
						c <- testNode
						continue
					}
				}

				testCase.runCase(testCase.Name, testNode)
				if x := recover(); x != nil {
					testNode.fail()
					testNode.AddDetail(RESULT, "testcase %s failed", testCase.Name)
				}

				if t.AfterEach != nil {
					testNode.AddDetail(STEP, "Running AfterEach before testcase %s", testCase.Name)
					t.AfterEach(testNode)
					if x := recover(); x != nil || testNode.result == FAIL {
						testNode.fail()
						testNode.AddDetail(RESULT, "AfterEach failed before testcase %s", testCase.Name)
						c <- testNode
						continue
					}
				}

				c <- testNode
			}
		} else {
			testNode := NewAssertion(testCase.Name, TEST_CASE, node, logger)
			testNode.result = IGNORE
			c <- testNode
		}
	}

	if t.AfterAll != nil {
		logger.Log(STEP, "Running AfterAll")
		t.AfterAll(node)
	}
	if x := recover(); x != nil {
		node.fail()
		logger.Log(RESULT, "AfterAll failed on feature %s", t.Name)
	}
	defer logger.Log(FEATURE_END, "End running feature %s", t.Name)

}

func (t *TestCase) runCase(name string, testNode *Assertion, params ...interface{}) {
	testNode.AddDetail(CASE_START, "Start running case %s", name)
	t.Case(testNode, params...)
	testNode.AddDetail(CASE_END, "End running case %s", name)
}

type Logger interface {
	Log(logType string, message string, args ...interface{})
}

type Result int

const (
	NOTRUN Result = iota
	PASS
	FAIL
	IGNORE
)

func (r Result) String() string {
	switch r {
	case NOTRUN:
		return "Not Run"
	case PASS:
		return "Pass"
	case FAIL:
		return "Fail"
	case IGNORE:
		return "Ignore"
	default:
		return "unkonwn"
	}
}

type NodeType int

const (
	TEST_SUITE NodeType = iota
	TEST_FEATURE
	TEST_CASE
)

func (n NodeType) String() string {
	switch n {
	case TEST_SUITE:
		return "Test Suite"
	case TEST_FEATURE:
		return "Test Feature"
	case TEST_CASE:
		return "Test Case"
	default:
		return "unknown"
	}
}
