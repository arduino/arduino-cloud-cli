package ota

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arduino/arduino-cloud-cli/internal/iot/mocks"
	"github.com/stretchr/testify/mock"
)

func TestRun(t *testing.T) {
	mockClient := &mocks.Client{}
	mockDeviceOTA := func(id string, file *os.File, expireMins int) error {
		time.Sleep(3 * time.Second)
		if strings.Split(id, "-")[0] == "fail" {
			return errors.New("err")
		}
		return nil
	}
	mockClient.On("DeviceOTA", mock.Anything, mock.Anything, mock.Anything).Return(mockDeviceOTA, nil)

	err := run(mockClient, []string{"dont-fail", "fail-1", "dont-fail", "fail-2"}, nil, 0)
	if err == nil {
		t.Error("should return error")
	}
	fmt.Println(err.Error())
	failed := strings.Split(err.Error(), ",")
	if len(failed) != 2 {
		fmt.Println(len(failed), failed)
		t.Error("two updates should have failed")
	}
}
