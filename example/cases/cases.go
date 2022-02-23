package cases

import (
	"time"

	"github.com/jimmyseraph/sparkle/engine"
	"github.com/jimmyseraph/sparkle/handler"
	"github.com/jimmyseraph/sparkle/logger"
)

func TestDemo() *engine.TestFeature {
	return &engine.TestFeature{
		Name: "Test Feature Example",
		TestCases: []engine.TestCase{
			{
				Name: "TestCase-1",
				Case: func(assertion *engine.Assertion, args ...interface{}) {
					time.Sleep(2 * time.Second)
					expected := 2
					actual := 1 + 1
					assertion.AssertEquals(expected, actual, "int assert")
				},
			},
		},
	}
}

func Run() {
	logger := logger.NewZapLogger()
	c := make(chan *engine.Assertion, 10)
	quit := make(chan bool, 1)
	handler := handler.NewZapHandler()
	go func() {
		TestDemo().RunFeature(nil, logger, c)
		quit <- true
	}()

	engine.StartListener(handler, c, quit)
}
