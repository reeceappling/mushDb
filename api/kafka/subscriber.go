package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

func startKafkaSubscriber(topic string, consumerGroupId *string) { // TODO: THOROUGHLY GO THROUGH
	// 1. Create a Reader config utilizing Consumer Groups
	readerConfig := kafka.ReaderConfig{
		Brokers:  []string{"kafka:9092"}, // TODO; FIX THIS!
		Topic:    topic,                  // TODO: FIX!
		MinBytes: 10e3,                   // 10KB // TODO: ensure ok
		MaxBytes: 10e6,                   // 10MB // TODO: ensure ok
	}
	if consumerGroupId != nil {
		readerConfig.GroupID = *consumerGroupId // Setting GroupID enables Consumer Group balance
	}

	// 2. Initialize the reader
	reader := kafka.NewReader(readerConfig)
	defer reader.Close()

	// Handle graceful shutdown via Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigchan
		fmt.Println("\nShutting down consumer...")
		cancel()
	}()

	fmt.Println("Subscribed to topic. Waiting for messages...")

	// 3. Read loop
	for {
		// ReadMessage automatically blocks, fetches, and commits offsets
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled, exiting cleanly
				break
			}
			log.Printf("Error while reading message: %v", err)
			continue
		}

		// Process the message
		fmt.Printf("Message: key=%s value=%s partition=%d offset=%d\n",
			string(msg.Key), string(msg.Value), msg.Partition, msg.Offset)
	}
}
