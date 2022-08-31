package googleapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fileIdForTest = "1IHQpHTk52hHCYKnIk1CfEH2sVIMIV4l1QzybzZS6I0Q"

func TestListFilesSample(t *testing.T) {
	gdriver := NewGDriver()
	err := gdriver.ListFilesSample(context.Background())
	assert.NoError(t, err)
}

func TestGetFilesSample(t *testing.T) {
	gdriver := NewGDriver()
	err := gdriver.GetFilesSample(context.Background(), fileIdForTest)
	assert.NoError(t, err)
}

func TestAddEditPermissionForAnyone(t *testing.T) {
	gdriver := NewGDriver()
	err := gdriver.AddEditPermissionForAnyone(context.Background(), fileIdForTest)
	assert.NoError(t, err)
	t.Log("add permission finish")
}
