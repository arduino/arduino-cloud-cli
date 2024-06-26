package otaapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgressBar_notCompletePct(t *testing.T) {
	firmwareSize := int64(25665 * 2)
	bar := formatStateData("fetch", "25665", &firmwareSize, false)
	assert.Equal(t, "[=====     ] 50.00%", bar)
}

func TestProgressBar_ifFlashState_goTo100Pct(t *testing.T) {
	firmwareSize := int64(25665 * 2)
	bar := formatStateData("fetch", "25665", &firmwareSize, true) // If in flash status, go to 100%
	assert.Equal(t, "[==========] 100.00%", bar)
}

func TestProgressBar_ifFlashStateAndUnknown_goTo100Pct(t *testing.T) {
	firmwareSize := int64(25665 * 2)
	bar := formatStateData("fetch", "Unknown", &firmwareSize, true) // If in flash status, go to 100%
	assert.Equal(t, "[==========] 100.00%", bar)
}
