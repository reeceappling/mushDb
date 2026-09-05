package mqtt

import (
	"context"
	"errors"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func NewBroker(ctx context.Context) error { // TODO: likely use a docker image with a config instead! see https://github.com/mochi-mqtt/server
	// 1. Create a new MQTT Server instance
	server := mqtt.New(&mqtt.Options{
		Capabilities: &mqtt.Capabilities{
			MaximumSessionExpiryInterval: 3600, // TODO: fix
			MaximumClientWritesPending:   3,    // TODO: fix
			Compatibilities: mqtt.Compatibilities{
				ObscureNotAuthorized: true,
			},
		},
		ClientNetWriteBufferSize: 4096,  // TODO: fix
		ClientNetReadBufferSize:  4096,  // TODO: fix
		SysTopicResendInterval:   10,    // TODO: fix
		InlineClient:             false, // TODO: fix
	})
	handleIncomingMqttItem := func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
		topicNameParts := strings.Split(pk.TopicName, "/")
		switch topicNameParts[0] {
		case "sensors":
			// sensors/location/groupID/type
			//location, groupId, typ := topicNameParts[1], topicNameParts[2], topicNameParts[3]
			msgs := []kafka.Message{{
				Topic: pk.TopicName,
				//Key:        nil, // TODO: time?
				Value: pk.Payload,
				//Time:       time.Time{}, // TODO: fix?
			}, {
				Topic: "sensors",            // TODO: send to topic with only the latest sensors info
				Key:   []byte(pk.TopicName), // TODO: time?
				Value: pk.Payload,
				//Time:       time.Time{}, // TODO: fix?
			}}
			err := writeToKafka(ctx, msgs...)
			if err != nil {
				panic("failed to write")
			}
		case "controllers": // TODO: ????
			// controllers/location/groupID/type
			//location, groupId, typ := topicNameParts[1], topicNameParts[2], topicNameParts[3]
			msgs := []kafka.Message{{
				Topic: pk.TopicName,
				//Key:        nil, // TODO: time?
				Value: pk.Payload,
				//Time:       time.Time{}, // TODO: fix?
			}, {
				Topic: "controllers",        // TODO: send to topic with only the latest sensors info
				Key:   []byte(pk.TopicName), // TODO: time?
				Value: pk.Payload,
				//Time:       time.Time{}, // TODO: fix?
			}}
			err := writeToKafka(ctx, msgs...)
			if err != nil {
				panic("failed to write")
			}
		default:
			panic("unknown topic!")
		}

		//server.Log.Info("inline client received message from subscription", "client", cl.ID, "subscriptionId", sub.Identifier, "topic", pk.TopicName, "payload", string(pk.Payload))
	}
	err := server.Subscribe("sensors/#", 1, handleIncomingMqttItem)
	if err != nil {
		panic(err)
	}
	err = server.Subscribe("controllers/#", 1, handleIncomingMqttItem)
	if err != nil {
		panic(err)
	}
	/*
		Single-level Wildcard (+): Replaces a single sub-topic level (e.g., subscribing to home/+/temperature matches home/livingroom/temperature and home/kitchen/temperature).
		Multi-level Wildcard (#): Matches any number of deeper sub-topic levels at the tail end of a string (e.g., subscribing to home/# matches all sub-topics under home).
	*/

	// 2. Allow all connections (For production, use secure authentication hooks!) // TODO: this!
	_ = server.AddHook(new(auth.AllowHook), nil) // TODO: change AllowHook!

	// 3. Create a TCP listener on the standard MQTT port (1883)
	tcpListener := listeners.NewTCP(listeners.Config{
		ID:      "t1",    // TODO: this!
		Address: ":1883", // TODO: this!
	})

	err = server.AddListener(tcpListener)
	if err != nil {
		return errors.Join(errors.New("failed to add listener"), err)
	}

	// 4. Start the broker server asynchronously
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatalf("Broker server error: %v", err)
		}
	}()
	log.Println("MQTT Broker is running on port :1883...")

	// 5. Wait for a system signal to gracefully shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer server.Close()
		<-sigChan
		log.Println("Shutting down MQTT Broker...")
	}()
	return nil
}

func writeToKafka(ctx context.Context, msgs ...kafka.Message) error { // TODO: Messages must contain topic!
	w := &kafka.Writer{
		Addr: kafka.TCP("localhost:9092", "localhost:9093", "localhost:9094"), // TODO: FIX!
		// NOTE: When Topic is not defined here, each Message must define it instead.
		Balancer: &kafka.LeastBytes{}, // TODO: ????
	}

	err := w.WriteMessages(ctx, msgs...)
	if err != nil {
		return err
	}

	return w.Close()
}
