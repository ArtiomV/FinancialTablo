package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAssetsApi_NotNil(t *testing.T) {
	api := NewAssetsApi(nil)
	assert.NotNil(t, api)
}

func TestAssets_Singleton(t *testing.T) {
	assert.NotNil(t, Assets)
	assert.NotNil(t, Assets.assets)
}
