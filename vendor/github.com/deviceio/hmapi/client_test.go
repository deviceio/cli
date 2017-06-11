package hmapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
}

func (t *ClientTestSuite) Test_NewClient_empty_config_expected_construction() {
	c := NewClient(&ClientConfig{}).(*client)

	assert.Equal(t.T(), 80, c.config.Port)
	assert.Equal(t.T(), HTTP, c.config.Scheme)
	assert.Equal(t.T(), "http://localhost:80", c.baseuri)
	assert.IsType(t.T(), new(AuthNone), c.config.Auth)
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
