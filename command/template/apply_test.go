package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateLIsting(t *testing.T) {
	tmpDir := t.TempDir()
	path := ".testdata/template_ok-rp-with-binaries.tino"
	file, err := extractBinary(&path, tmpDir)
	assert.Nil(t, err)
	assert.NotNil(t, file)
}
