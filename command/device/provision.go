package device

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/iot-cloud-cli/arduino"
	"github.com/arduino/iot-cloud-cli/internal/iot"
	"github.com/arduino/iot-cloud-cli/internal/serial"
	"github.com/sirupsen/logrus"
)

type provision struct {
	arduino.Commander
	iot.Client
	ser *serial.Serial
	dev *device
	id  string
}

type binFile struct {
	Bin      string `json:"bin"`
	Filename string `json:"filename"`
	Fqbn     string `json:"fqbn"`
	Name     string `json:"name"`
	Sha256   string `json:"sha256"`
}

func (p provision) run() error {
	bin, err := downloadProvisioningFile(p.dev.fqbn)
	if err != nil {
		return err
	}

	logrus.Infof("\n%s\n", "Uploading provisioning sketch on the device")
	time.Sleep(500 * time.Millisecond)
	// Try to upload the provisioning sketch
	errMsg := "Error while uploading the provisioning sketch: "
	err = retry(5, time.Millisecond*1000, errMsg, func() error {
		//serialutils.Reset(dev.port, true, nil)
		return p.UploadBin(p.dev.fqbn, bin, p.dev.port)
	})
	if err != nil {
		return err
	}

	logrus.Infof("\n%s\n", "Connecting to the device through serial port")
	// Try to connect to device through the serial port
	time.Sleep(1500 * time.Millisecond)
	p.ser = serial.NewSerial()
	errMsg = "Error while connecting to the device: "
	err = retry(5, time.Millisecond*1000, errMsg, func() error {
		return p.ser.Connect(p.dev.port)
	})
	if err != nil {
		return err
	}
	defer p.ser.Close()
	logrus.Infof("%s\n\n", "Connected to device")

	// Send configuration commands to the device
	err = p.configDev()
	if err != nil {
		return err
	}

	logrus.Infof("%s\n\n", "Device provisioning successful")
	return nil
}

func (p provision) configDev() error {
	logrus.Infof("Receiving the certificate")
	csr, err := p.ser.SendReceive(serial.CSR, []byte(p.id))
	if err != nil {
		return err
	}
	cert, err := p.CertificateCreate(p.id, string(csr))
	if err != nil {
		return err
	}

	logrus.Infof("Requesting begin storage")
	err = p.ser.Send(serial.BeginStorage, nil)
	if err != nil {
		return err
	}

	s := strconv.Itoa(cert.NotBefore.Year())
	logrus.Infof("Sending year: ", s)
	err = p.ser.Send(serial.SetYear, []byte(s))
	if err != nil {
		return err
	}

	s = fmt.Sprintf("%02d", int(cert.NotBefore.Month()))
	logrus.Infof("Sending month: ", s)
	err = p.ser.Send(serial.SetMonth, []byte(s))
	if err != nil {
		return err
	}

	s = fmt.Sprintf("%02d", cert.NotBefore.Day())
	logrus.Infof("Sending day: ", s)
	err = p.ser.Send(serial.SetDay, []byte(s))
	if err != nil {
		return err
	}

	s = fmt.Sprintf("%02d", cert.NotBefore.Hour())
	logrus.Infof("Sending hour: ", s)
	err = p.ser.Send(serial.SetHour, []byte(s))
	if err != nil {
		return err
	}

	s = strconv.Itoa(31)
	logrus.Infof("Sending validity: ", s)
	err = p.ser.Send(serial.SetValidity, []byte(s))
	if err != nil {
		return err
	}

	logrus.Infof("Sending certificate serial")
	b, err := hex.DecodeString(cert.Serial)
	if err != nil {
		err = fmt.Errorf("%s: %w", "decoding certificate serial", err)
		return err
	}
	err = p.ser.Send(serial.SetCertSerial, b)
	if err != nil {
		return err
	}

	logrus.Infof("Sending certificate authority key")
	b, err = hex.DecodeString(cert.AuthorityKeyIdentifier)
	if err != nil {
		err = fmt.Errorf("%s: %w", "decoding certificate authority key id", err)
		return err
	}
	err = p.ser.Send(serial.SetAuthKey, b)
	if err != nil {
		return err
	}

	logrus.Infof("Sending certificate signature")
	b, err = hex.DecodeString(cert.SignatureAsn1X + cert.SignatureAsn1Y)
	if err != nil {
		err = fmt.Errorf("%s: %w", "decoding certificate signature", err)
		return err
	}
	err = p.ser.Send(serial.SetSignature, b)
	if err != nil {
		return err
	}

	time.Sleep(time.Second)
	logrus.Infof("Requesting end storage")
	err = p.ser.Send(serial.EndStorage, nil)
	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second)
	logrus.Infof("Requesting certificate reconstruction")
	err = p.ser.Send(serial.ReconstructCert, nil)
	if err != nil {
		return err
	}

	return nil
}

func downloadProvisioningFile(fqbn string) (string, error) {
	// Use local binaries until they are uploaded online
	bin := filepath.Join("./binaries/", strings.ReplaceAll(fqbn, ":", ".")+".bin")
	bin, err := filepath.Abs(bin)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(bin); err == nil {
		return bin, nil
	}

	elf := filepath.Join("./binaries/", strings.ReplaceAll(fqbn, ":", ".")+".elf")
	elf, err = filepath.Abs(elf)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(elf); os.IsNotExist(err) {
		err = fmt.Errorf("%s: %w", "fqbn not supported", err)
		return "", err
	}
	return elf, nil

	// TODO: upload binaries on some arduino page and enable this flow
	//url := "https://api2.arduino.cc/iot/v2/binaries/provisioning?fqbn=" + fqbn
	//path, _ := filepath.Abs("./provisioning.bin")

	//cl := http.Client{
	//Timeout: time.Second * 3, // Timeout after 2 seconds
	//}

	//req, err := http.NewRequest(http.MethodGet, url, nil)
	//if err != nil {
	//err = fmt.Errorf("%s: %w", "request provisioning binary", err)
	//return "", err
	//}
	//res, err := cl.Do(req)
	//if err != nil {
	//err = fmt.Errorf("%s: %w", "request provisioning binary", err)
	//return "", err
	//}

	//if res.Body != nil {
	//defer res.Body.Close()
	//}

	//body, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//err = fmt.Errorf("%s: %w", "read provisioning request body", err)
	//return "", err
	//}

	//bin := binFile{}
	//err = json.Unmarshal(body, &bin)
	//if err != nil {
	//err = fmt.Errorf("%s: %w", "unmarshal provisioning binary", err)
	//return "", err
	//}

	//bytes, err := base64.StdEncoding.DecodeString(bin.Bin)
	//if err != nil {
	//err = fmt.Errorf("%s: %w", "decoding provisioning binary", err)
	//return "", err
	//}

	//err = ioutil.WriteFile(path, bytes, 0666)
	//if err != nil {
	//err = fmt.Errorf("%s: %w", "downloading provisioning binary", err)
	//return "", err
	//}

	//return path, nil
}

func retry(tries int, sleep time.Duration, errMsg string, fun func() error) error {
	var err error
	for n := 0; n < tries; n++ {
		err = fun()
		if err == nil {
			break
		}
		logrus.Warningf(errMsg, err.Error(), "\nTrying again...")
		time.Sleep(sleep)
	}
	return err
}
