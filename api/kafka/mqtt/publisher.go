package mqtt

import (
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type PublisherClient struct {
	internal mqtt.Client
	timeout  *time.Duration
}

const (
	atMostOnce  byte = 0
	atLeastOnce byte = 1
	exactlyOnce byte = 2
)

func (client PublisherClient) Publish(topic string, qos byte, keepMostRecentForTopicOnBroker bool, msg string) error {
	token := client.internal.Publish(topic, qos, keepMostRecentForTopicOnBroker, msg)
	var delivered = false
	if client.timeout != nil {
		delivered = token.WaitTimeout(*client.timeout)
	} else {
		delivered = token.Wait()
	}
	if !delivered {
		// TODO: handle not delivered
		return fmt.Errorf("failed to deliver message during timeframe")
	}
	return token.Error()
}
func NewPublisher(broker string, clientId string, timeout *time.Duration) (PublisherClient, error) {
	out := PublisherClient{
		timeout: timeout,
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientId)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	out.internal = mqtt.NewClient(opts)
	token := out.internal.Connect()
	if token.Wait() && token.Error() != nil { // TODO: wait duration?
		return out, token.Error()
	}
	return out, nil
}

const (
	broker   = "tcp://localhost:1883"     // TODO: FIX!
	clientId = "mqtt-publisher-client-id" // TODO: FIX!
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected to MQTT Broker") // TODO: what here?
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err) // TODO: what here?
}

type SensorReading struct {
	grouping string
	// TODO: does this need sensor name or will that be the topic name? sensors.grouping.sensorName?
	value float64
}

func main() {
	cid := "my-publisher-client-id" // TODO; fix
	client, err := NewPublisher(broker, cid, utils.Pointer(3*time.Second))
	if err != nil {
		panic(err)
	}
	//airTempSensorQoS, soilTempSensorQoS, rhSensorQoS, soilMoistureSensorQoS, lightSensorQoS, o2SensorQoS,co2SensorQoS := atLeastOnce,atLeastOnce,atLeastOnce,atLeastOnce,atLeastOnce,atLeastOnce,atLeastOnce
	//foggerQoS,fanQoS, lightQoS := atLeastOnce,atLeastOnce,atLeastOnce

	err = client.Publish("topicName1", atLeastOnce, false, "hello world 1")
	if err != nil {
		panic("1: " + err.Error())
	}
	err = client.Publish("topicName2", atMostOnce, false, "hello world 2")
	if err != nil {
		panic("2: " + err.Error())
	}
	err = client.Publish("topicName3", exactlyOnce, false, "hello world 3")
	if err != nil {
		panic("3: " + err.Error())
	}
}
