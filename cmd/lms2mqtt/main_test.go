package main

import (
	"errors"
	"fmt"
	"github.com/cyrilix/mqtt-tools/mqttTooling"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"os"
	"testing"
)

func Test_Cli(t *testing.T) {
	mock := &appMock{
		subscriptions: make([]Subscription, 0),
		hasStarted:    false,
		isConnected:   true,
	}
	var oldNewApp = newApplication
	defer func() { newApplication = oldNewApp }()

	username := "user"
	password := "password"
	broker := "tcp://mqtt.example.com:1883"
	server := "192.168.0.1:9000"
	topic := "test"
	os.Args = []string{
		"./lms2mqtt",
		fmt.Sprintf("-mqtt-broker=%v", broker),
		fmt.Sprintf("-mqtt-username=%v", username),
		fmt.Sprintf("-mqtt-password=%v", password),
		fmt.Sprintf("-address=%v", server),
		fmt.Sprintf("-mqtt-topic=%v", topic),
		"-debug",
	}
	newApplication = func(mcp *mqttTooling.MqttCliParameters, top, serverAddress string) (RunInterruptable, error) {
		if serverAddress != server {
			t.Errorf("bad server address: %v, wants %v", serverAddress, server)
		}
		if broker != mcp.Broker {
			t.Errorf("bad mqtt broker: %v, wants %v", mcp.Broker, broker)
		}
		if top != topic {
			t.Errorf("bad mqtt topic: %v, wants %v", top, topic)
		}
		if username != mcp.Username {
			t.Errorf("bad mqtt user: %v, wants %v", mcp.Username, username)
		}
		if password != mcp.Password {
			t.Errorf("bad mqtt password: %v, wants %v", mcp.Password, password)
		}
		return mock, nil
	}

	main()

	if !mock.hasStarted {
		t.Errorf("application hasn't started has expected")
	}
	if !mock.stopCalled {
		t.Errorf("application hasn't stopped has expected")
	}
}

type Subscription struct {
	topic    string
	callback MQTT.MessageHandler
}

type appMock struct {
	t             testing.T
	isConnected   bool
	subscriptions []Subscription
	hasStarted    bool
	stopCalled    bool
}

func (a *appMock) Connect() error {
	a.isConnected = true
	return nil
}
func (a *appMock) Run() error {
	if !a.isConnected {
		a.t.Error("try to run but application isn't connected")
		return errors.New("try to run but application isn't connected")
	}
	a.hasStarted = true
	return nil
}

func (a *appMock) Subscribe(topic string, callback MQTT.MessageHandler) error {
	if !a.isConnected {
		a.t.Error("try to subscribe but application isn't connected")
		return errors.New("try to subscribe but application isn't connected")
	}
	a.subscriptions = append(a.subscriptions, Subscription{topic, callback})
	return nil
}

func (a *appMock) Stop() {
	a.stopCalled = true
	a.isConnected = false
}
