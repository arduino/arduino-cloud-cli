package device

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arduino/arduino-cloud-cli/arduino/cli"
	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	iotclient "github.com/arduino/iot-client-go"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

const (
	deveuiUploadAttempts = 3
	deveuiUploadWait     = 1000

	serialEUIAttempts = 4
	serialEUIWait     = 2000
	serialEUITimeout  = 3500
	serialEUIBaudrate = 9600

	// dev-eui is an IEEE EUI64 address, so it must have length of 8 bytes.
	// It's retrieved as hexadecimal string, thus 16 chars are expected
	deveuiLength = 16
)

// DeviceLoraInfo contains the most interesting
// parameters of an Arduino IoT Cloud LoRa device.
type DeviceLoraInfo struct {
	DeviceInfo
	AppEUI string `json:"app-eui"`
	AppKey string `json:"app-key"`
	EUI    string `json:"eui"`
}

// CreateLoRaParams contains the parameters needed
// to provision a LoRa device.
type CreateLoraParams struct {
	CreateParams
	FrequencyPlan string
}

// CreateLora command is used to provision a new LoRa arduino device
// and to add it to Arduino IoT Cloud.
func CreateLora(params *CreateLoraParams) (*DeviceLoraInfo, error) {
	comm, err := cli.NewCommander()
	if err != nil {
		return nil, err
	}

	ports, err := comm.BoardList()
	if err != nil {
		return nil, err
	}
	board := boardFromPorts(ports, &params.CreateParams)
	if board == nil {
		err = errors.New("no board found")
		return nil, err
	}

	bin, err := deveuiBinary(board.fqbn)
	if err != nil {
		return nil, fmt.Errorf("fqbn not supported for LoRa provisioning: %w", err)
	}

	logrus.Infof("%s", "Uploading deveui sketch on the LoRa board")
	errMsg := "Error while uploading the LoRa provisioning binary"
	err = retry(deveuiUploadAttempts, deveuiUploadWait*time.Millisecond, errMsg, func() error {
		return comm.UploadBin(board.fqbn, bin, board.port)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload LoRa provisioning binary: %w", err)
	}

	eui, err := extractEUI(board.port)
	if err != nil {
		return nil, err
	}

	conf, err := config.Retrieve()
	if err != nil {
		return nil, err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return nil, err
	}

	logrus.Info("Creating a new device on the cloud")
	dev, err := iotClient.DeviceLoraCreate(params.Name, board.serial, board.dType, eui, params.FrequencyPlan)
	if err != nil {
		return nil, err
	}

	devInfo, err := getDeviceLoraInfo(iotClient, dev)
	if err != nil {
		iotClient.DeviceDelete(dev.DeviceId)
		err = fmt.Errorf("%s: %w", "cannot provision LoRa device", err)
		return nil, err
	}
	return devInfo, nil
}

// deveuiBinary gets the absolute path of the deveui binary corresponding to the
// provisioned board's fqbn. It is contained in the local binaries folder.
func deveuiBinary(fqbn string) (string, error) {
	// Use local binaries until they are uploaded online
	bin := filepath.Join("./binaries/", "getdeveui."+strings.ReplaceAll(fqbn, ":", ".")+".bin")
	bin, err := filepath.Abs(bin)
	if err != nil {
		return "", fmt.Errorf("getting the deveui binary: %w", err)
	}
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		err = fmt.Errorf("%s: %w", "deveui binary not found", err)
		return "", err
	}
	return bin, nil
}

// extractEUI extracts the EUI from the provisioned lora board
func extractEUI(port string) (string, error) {
	var ser serial.Port

	logrus.Infof("%s\n", "Connecting to the board through serial port")
	errMsg := "Error while connecting to the board"
	err := retry(serialEUIAttempts, serialEUIWait*time.Millisecond, errMsg, func() error {
		var err error
		ser, err = serial.Open(port, &serial.Mode{BaudRate: serialEUIBaudrate})
		return err
	})
	if err != nil {
		return "", fmt.Errorf("failed to extract deveui from the board: %w", err)
	}

	err = ser.SetReadTimeout(serialEUITimeout * time.Millisecond)
	if err != nil {
		return "", fmt.Errorf("setting serial read timeout: %w", err)
	}

	buff := make([]byte, deveuiLength)
	n, err := ser.Read(buff)
	if err != nil {
		return "", fmt.Errorf("reading from serial: %w", err)
	}

	if n < deveuiLength {
		return "", errors.New("cannot read eui from the device")
	}
	eui := string(buff)
	return eui, nil
}

func getDeviceLoraInfo(iotClient iot.Client, loraDev *iotclient.ArduinoLoradevicev1) (*DeviceLoraInfo, error) {
	dev, err := iotClient.DeviceShow(loraDev.DeviceId)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve device from the cloud: %w", err)
	}

	devInfo := &DeviceLoraInfo{
		DeviceInfo: DeviceInfo{
			Name:   dev.Name,
			ID:     dev.Id,
			Board:  dev.Type,
			Serial: dev.Serial,
			FQBN:   dev.Fqbn,
		},
		AppEUI: loraDev.AppEui,
		AppKey: loraDev.AppKey,
		EUI:    loraDev.Eui,
	}
	return devInfo, nil
}
