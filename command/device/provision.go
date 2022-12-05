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
	"context"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/arduino/arduino-cloud-cli/arduino"
	"github.com/arduino/arduino-cloud-cli/internal/binary"
	"github.com/arduino/arduino-cloud-cli/internal/serial"
	"github.com/arduino/go-paths-helper"
	iotclient "github.com/arduino/iot-client-go"
	"github.com/sirupsen/logrus"
)

// downloadProvisioningFile downloads and returns the absolute path
// of the provisioning binary corresponding to the passed fqbn.
func downloadProvisioningFile(ctx context.Context, fqbn string) (string, error) {
	index, err := binary.LoadIndex(ctx)
	if err != nil {
		return "", err
	}
	bin := index.FindProvisionBin(fqbn)
	if bin == nil {
		return "", fmt.Errorf("provisioning binary for board %s not found", fqbn)
	}
	bytes, err := binary.Download(ctx, bin)
	if err != nil {
		return "", fmt.Errorf("downloading provisioning binary: %w", err)
	}

	// Save provision binary always in the same temporary folder to
	// avoid wasting user's storage.
	filename := filepath.Base(bin.URL)
	path := paths.TempDir().Join("cloud-cli").Join(filename)
	path.Parent().MkdirAll()
	if err = path.WriteFile(bytes); err != nil {
		return "", fmt.Errorf("writing provisioning binary: %w", err)
	}
	p, err := path.Abs()
	if err != nil {
		return "", fmt.Errorf("cannot retrieve absolute path of downloaded binary: %w", err)
	}
	return p.String(), nil
}

type certificateCreator interface {
	CertificateCreate(ctx context.Context, id, csr string) (*iotclient.ArduinoCompressedv2, error)
}

// provision is responsible for running the provisioning
// procedures for boards with crypto-chip.
type provision struct {
	arduino.Commander
	cert  certificateCreator
	ser   *serial.Serial
	board *board
	id    string
}

// run provisioning procedure for boards with crypto-chip.
func (p provision) run(ctx context.Context) error {
	bin, err := downloadProvisioningFile(ctx, p.board.fqbn)
	if err != nil {
		return err
	}

	// Try to upload the provisioning sketch
	logrus.Infof("%s\n", "Uploading provisioning sketch on the board")
	if err = sleepCtx(ctx, 500*time.Millisecond); err != nil {
		return err
	}
	errMsg := "Error while uploading the provisioning sketch"
	err = retry(ctx, 5, time.Millisecond*1000, errMsg, func() error {
		//serialutils.Reset(dev.port, true, nil)
		return p.UploadBin(ctx, p.board.fqbn, bin, p.board.address, p.board.protocol)
	})
	if err != nil {
		return err
	}

	// Try to connect to board through the serial port
	logrus.Infof("%s\n", "Connecting to the board through serial port")
	if err = sleepCtx(ctx, 1500*time.Millisecond); err != nil {
		return err
	}
	p.ser = serial.NewSerial()
	errMsg = "Error while connecting to the board"
	err = retry(ctx, 5, time.Millisecond*1000, errMsg, func() error {
		return p.ser.Connect(p.board.address)
	})
	if err != nil {
		return err
	}
	defer p.ser.Close()
	logrus.Infof("%s\n\n", "Connected to the board")

	// Wait some time before using the serial port
	if err = sleepCtx(ctx, 2000*time.Millisecond); err != nil {
		return err
	}

	// Send configuration commands to the board
	if err = p.configBoard(ctx); err != nil {
		return err
	}

	logrus.Infof("%s\n\n", "Device provisioning successful")
	return nil
}

func (p provision) configBoard(ctx context.Context) error {
	logrus.Info("Receiving the certificate")
	csr, err := p.ser.SendReceive(ctx, serial.CSR, []byte(p.id))
	if err != nil {
		return err
	}
	cert, err := p.cert.CertificateCreate(ctx, p.id, string(csr))
	if err != nil {
		return err
	}

	logrus.Info("Requesting begin storage")
	if err = p.ser.Send(ctx, serial.BeginStorage, nil); err != nil {
		return err
	}

	s := strconv.Itoa(cert.NotBefore.Year())
	logrus.Info("Sending year: ", s)
	if err = p.ser.Send(ctx, serial.SetYear, []byte(s)); err != nil {
		return err
	}

	s = fmt.Sprintf("%02d", int(cert.NotBefore.Month()))
	logrus.Info("Sending month: ", s)
	if err = p.ser.Send(ctx, serial.SetMonth, []byte(s)); err != nil {
		return err
	}

	s = fmt.Sprintf("%02d", cert.NotBefore.Day())
	logrus.Info("Sending day: ", s)
	if err = p.ser.Send(ctx, serial.SetDay, []byte(s)); err != nil {
		return err
	}

	s = fmt.Sprintf("%02d", cert.NotBefore.Hour())
	logrus.Info("Sending hour: ", s)
	if err = p.ser.Send(ctx, serial.SetHour, []byte(s)); err != nil {
		return err
	}

	s = strconv.Itoa(31)
	logrus.Info("Sending validity: ", s)
	if err = p.ser.Send(ctx, serial.SetValidity, []byte(s)); err != nil {
		return err
	}

	logrus.Info("Sending certificate serial")
	b, err := hex.DecodeString(cert.Serial)
	if err != nil {
		return fmt.Errorf("decoding certificate serial: %w", err)
	}
	if err = p.ser.Send(ctx, serial.SetCertSerial, b); err != nil {
		return err
	}

	logrus.Info("Sending certificate authority key")
	b, err = hex.DecodeString(cert.AuthorityKeyIdentifier)
	if err != nil {
		return fmt.Errorf("decoding certificate authority key id: %w", err)
	}
	if err = p.ser.Send(ctx, serial.SetAuthKey, b); err != nil {
		return err
	}

	logrus.Info("Sending certificate signature")
	b, err = hex.DecodeString(cert.SignatureAsn1X + cert.SignatureAsn1Y)
	if err != nil {
		err = fmt.Errorf("decoding certificate signature: %w", err)
		return err
	}
	if err = p.ser.Send(ctx, serial.SetSignature, b); err != nil {
		return err
	}

	if err := sleepCtx(ctx, 1*time.Second); err != nil {
		return err
	}

	logrus.Info("Requesting end storage")
	if err = p.ser.Send(ctx, serial.EndStorage, nil); err != nil {
		return err
	}

	if err := sleepCtx(ctx, 2*time.Second); err != nil {
		return err
	}

	logrus.Info("Requesting certificate reconstruction")
	if err = p.ser.Send(ctx, serial.ReconstructCert, nil); err != nil {
		return err
	}

	return nil
}

func retry(ctx context.Context, tries int, sleep time.Duration, errMsg string, fun func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var err error
	for n := 0; n < tries; n++ {
		err = fun()
		if err == nil {
			break
		}
		logrus.Warningf("%s: %s: %s", errMsg, err.Error(), "\nTrying again...")
		if err := sleepCtx(ctx, sleep); err != nil {
			return err
		}
	}
	return err
}

func sleepCtx(ctx context.Context, tm time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(tm):
		return nil
	}
}
