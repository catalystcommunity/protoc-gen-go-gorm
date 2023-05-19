package test

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type PluginSuite struct {
	suite.Suite
}

func TestPluginSuite(t *testing.T) {
	suite.Run(t, new(PluginSuite))
}

func (s *PluginSuite) TestPlugin() {
}

func (s *PluginSuite) SetupTest() {
}
