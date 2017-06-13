package hmapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type when_constructing_new_hmapi_client struct {
	suite.Suite
}

func (t *when_constructing_new_hmapi_client) Test_with_empty_config_returns_expected_construction() {
	c := NewClient(&ClientConfig{}).(*client)

	assert.Equal(t.T(), 80, c.config.Port)
	assert.Equal(t.T(), HTTP, c.config.Scheme)
	assert.Equal(t.T(), "http://localhost:80", c.baseuri)
	assert.IsType(t.T(), new(AuthNone), c.config.Auth)
}

func TestRunClientTestSuites(t *testing.T) {
	suite.Run(t, new(when_constructing_new_hmapi_client))
}
