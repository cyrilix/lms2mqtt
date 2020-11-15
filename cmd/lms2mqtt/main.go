package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cyrilix/lms2mqtt/squeeze"
	"github.com/cyrilix/mqtt-tools/mqttTooling"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	defaultClientId = "lms2mqtt"
)

type RunInterruptable interface {
	Run() error
	Subscribe(topic string, callback MQTT.MessageHandler) error
	Stop()
}

type application struct {
	client  MQTT.Client
	params  *mqttTooling.MqttCliParameters
	topic   string
	address string
}

var newApplication = func(mcp *mqttTooling.MqttCliParameters, topic, serverAddress string) (RunInterruptable, error) {
	app := &application{
		params:  mcp,
		topic:   topic,
		address: serverAddress,
	}
	err := app.connect()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to mqtt bus: %v", err)
	}
	return app, nil
}

func (a *application) connect() error {
	if a.client != nil && a.client.IsConnected() {
		return fmt.Errorf("connection already exists")
	}
	client, err := mqttTooling.Connect(a.params)
	if err != nil {
		return fmt.Errorf("unable to connect to mqtt bus: %v", err)
	}
	a.client = client
	return nil
}

func (a *application) Stop() {
	if a.client != nil && a.client.IsConnected() {
		log.Info("Stop mqtt connection")
		a.client.Disconnect(50)
	}
}

func (a *application) Subscribe(topic string, onMessage MQTT.MessageHandler) error {
	t := a.client.Subscribe(topic, byte(a.params.Qos), onMessage)
	t.Wait()
	return t.Error()
}

func (a *application) Run() error {

	s := squeeze.New(a.address)
	go func() {
		err := s.Listen()
		if err != nil {
			log.Panicf("unable to connect to %v instance: %v", a.address, err)
		}
	}()
	defer func() {
		if err := s.Close(); err != nil {
			log.Warnf("unable to close channel: %v", err)
		}
	}()

	chanTrack := s.NotifyTrackChange()
	for {
		for t := range chanTrack {
			go a.publishTrack(a.topic, t)
		}
	}
}

func (a *application) publishTrack(topic string, t *squeeze.Track) {
	content, err := json.Marshal(*t)
	if err != nil {
		log.Errorf("unable to marshall message %#v: %v", *t, err)
		return
	}
	a.client.Publish(topic, byte(a.params.Qos), a.params.Retain, content)
}

func main() {
	var topic, address string
	var debug bool

	parameters := mqttTooling.MqttCliParameters{ClientId: defaultClientId}
	flag.StringVar(&topic, "mqtt-topic", "", "The topic name to/from which to publish/subscribe")
	flag.StringVar(&address, "address", "127.0.0.1:9090", "The squeezebox server address")
	flag.BoolVar(&debug, "debug", false, "Display debug logs")

	mqttTooling.InitMqttFlagSet(&parameters)
	flag.Parse()
	if len(os.Args) <= 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	configureLogs(debug)

	app, err := newApplication(&parameters, topic, address)
	if err != nil {
		log.Fatalf("unable to start application: %v", err)
	}
	defer app.Stop()

	err = app.Subscribe(topic, onMessage)
	if err != nil {
		log.Fatalf("unable to subscribe to topic %v: %v", topic, err)
	}

	err = app.Run()
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}

}

func configureLogs(debug bool) {
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
		PadLevelText:           true,
	})
	log.SetOutput(os.Stdout)
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetReportCaller(false)
}

var onMessage = func(client MQTT.Client, message MQTT.Message) {

}
