package client

import (
	"testing"

	"github.com/platform-mesh/golang-commons/directive/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewOpenFGAClient(t *testing.T) {
	client, err := NewOpenFGAClient(&mocks.OpenFGAServiceClient{})
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
