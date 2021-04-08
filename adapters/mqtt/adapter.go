package mqtt

import (
	"crypto/tls"
	"fmt"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Adapter struct {
	host          string
	clientID      string
	username      string
	password      string
	useSSL        bool
	autoReconnect bool
	client        paho.Client
	qos           int
}

func NewAdapterWithAuth(host, clientID, username, password string) *Adapter {
	return &Adapter{
		host:          host,
		clientID:      clientID,
		username:      username,
		password:      password,
		useSSL:        true,
		autoReconnect: true,
	}
}

func (a *Adapter) Connect() (err error) {
	a.client = paho.NewClient(a.createClientOptions())
	if token := a.client.Connect(); token.Wait() && token.Error() != nil {
		err = token.Error()
	}
	return
}

func (a *Adapter) Disconnect() (err error) {
	if a.client != nil {
		a.client.Disconnect(500)
	}
	return
}

func (a *Adapter) Publish(topic string, message []byte) error {
	_, err := a.PublishWithQOS(topic, a.qos, message)
	return err
}

func (a *Adapter) PublishWithQOS(topic string, qos int, message []byte) (paho.Token, error) {
	if a.client == nil {
		return nil, fmt.Errorf("MQTT client is nil")
	}

	token := a.client.Publish(topic, byte(qos), false, message)
	return token, nil
}


func (a *Adapter) On(topic string, f func(message paho.Message)) (bool, error) {
	_, err := a.OnWithQOS(topic, a.qos, f)
	if err != nil {
		return false, err
	}
	return true, err
}

func (a *Adapter) OnWithQOS(topic string, qos int, f func(message paho.Message)) (paho.Token, error) {
	if a.client == nil {
		return nil, fmt.Errorf("MQTT client is nil")
	}
	
	token := a.client.Subscribe(topic, byte(qos), func(client paho.Client, msg paho.Message) {
		f(msg)
	})
	return token, nil
	
}

func (a *Adapter) createClientOptions() *paho.ClientOptions {
	opts := paho.NewClientOptions()
	opts.AddBroker(a.host)
	opts.SetClientID(a.clientID)
	opts.SetUsername(a.username)
	opts.SetPassword(a.password)

	if a.useSSL {
		opts.SetTLSConfig(a.newTLSConfig())
	}
	return opts
}

func (a *Adapter) newTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS11,
		MaxVersion: tls.VersionTLS12,
		ClientAuth: tls.NoClientCert,
	}
}
