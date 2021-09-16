// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2021 ARDUINO SA (http://www.arduino.cc/)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package device

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/arduino-cloud-cli/arduino"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/arduino/arduino-cloud-cli/internal/serial"
	"github.com/sirupsen/logrus"
)

type provision struct {
	arduino.Commander
	iot.Client
	ser   *serial.Serial
	board *board
	id    string
}

type binFile struct {
	Bin      string `json:"bin"`
	Filename string `json:"filename"`
	Fqbn     string `json:"fqbn"`
	Name     string `json:"name"`
	Sha256   string `json:"sha256"`
}

func (p provision) run() error {
	bin, err := downloadProvisioningFile(p.board.fqbn)
	if err != nil {
		return err
	}

	logrus.Infof("%s\n", "Uploading provisioning sketch on the board")
	time.Sleep(500 * time.Millisecond)
	// Try to upload the provisioning sketch
	errMsg := "Error while uploading the provisioning sketch"
	err = retry(5, time.Millisecond*1000, errMsg, func() error {
		//serialutils.Reset(dev.port, true, nil)
		return p.UploadBin(p.board.fqbn, bin, p.board.port)
	})
	if err != nil {
		return err
	}

	logrus.Infof("%s\n", "Connecting to the board through serial port")
	// Try to connect to board through the serial port
	time.Sleep(1500 * time.Millisecond)
	p.ser = serial.NewSerial()
	errMsg = "Error while connecting to the board"
	err = retry(5, time.Millisecond*1000, errMsg, func() error {
		return p.ser.Connect(p.board.port)
	})
	if err != nil {
		return err
	}
	defer p.ser.Close()

	// Wait some time before using the serial port
	time.Sleep(500 * time.Millisecond)
	logrus.Infof("%s\n\n", "Connected to the board")

	// Send configuration commands to the board
	err = p.configBoard()
	if err != nil {
		return err
	}

	logrus.Infof("%s\n\n", "Device provisioning successful")
	return nil
}

func (p provision) configBoard() error {
	logrus.Info("Receiving the certificate")
	csr, err := p.ser.SendReceive(serial.CSR, []byte(p.id))
	if err != nil {
		return err
	}
	cert, err := p.CertificateCreate(p.id, string(csr))
	if err != nil {
		return err
	}

	logrus.Info("Requesting begin storage")
	err = p.ser.Send(serial.BeginStorage, nil)
	if err != nil {
		return err
	}

	s := strconv.Itoa(cert.NotBefore.Year())
	logrus.Info("Sending year: ", s)
	err = p.ser.Send(serial.SetYear, []byte(s))
	if err != nil {
		return err
	}

	s = fmt.Sprintf("%02d", int(cert.NotBefore.Month()))
	logrus.Info("Sending month: ", s)
	err = p.ser.Send(serial.SetMonth, []byte(s))
	if err != nil {
		return err
	}

	s = fmt.Sprintf("%02d", cert.NotBefore.Day())
	logrus.Info("Sending day: ", s)
	err = p.ser.Send(serial.SetDay, []byte(s))
	if err != nil {
		return err
	}

	s = fmt.Sprintf("%02d", cert.NotBefore.Hour())
	logrus.Info("Sending hour: ", s)
	err = p.ser.Send(serial.SetHour, []byte(s))
	if err != nil {
		return err
	}

	s = strconv.Itoa(31)
	logrus.Info("Sending validity: ", s)
	err = p.ser.Send(serial.SetValidity, []byte(s))
	if err != nil {
		return err
	}

	logrus.Info("Sending certificate serial")
	b, err := hex.DecodeString(cert.Serial)
	if err != nil {
		err = fmt.Errorf("%s: %w", "decoding certificate serial", err)
		return err
	}
	err = p.ser.Send(serial.SetCertSerial, b)
	if err != nil {
		return err
	}

	logrus.Info("Sending certificate authority key")
	b, err = hex.DecodeString(cert.AuthorityKeyIdentifier)
	if err != nil {
		err = fmt.Errorf("%s: %w", "decoding certificate authority key id", err)
		return err
	}
	err = p.ser.Send(serial.SetAuthKey, b)
	if err != nil {
		return err
	}

	logrus.Info("Sending certificate signature")
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
	logrus.Info("Requesting end storage")
	err = p.ser.Send(serial.EndStorage, nil)
	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second)
	logrus.Info("Requesting certificate reconstruction")
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
		logrus.Warningf("%s: %s: %s", errMsg, err.Error(), "\nTrying again...")
		time.Sleep(sleep)
	}
	return err
}
