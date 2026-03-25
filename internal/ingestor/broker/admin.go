package broker

import (
	"fmt"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func PrepareKafkaTopic(brokerAddress string, topicName string) error {

	// Kafka communicates over TCP
	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		fmt.Printf("Failed to dial Kafka: %v", err)
		return err
	}
	defer conn.Close() // Close connection

	fmt.Println("Successfully connected to the Kafka cluster")

	// Get Kafka controller (Authority to create, delete or modify topics)
	controller, err := conn.Controller()
	if err != nil {
		fmt.Printf("Failed to get Kafka controller: %v", err)
		return err
	}

	// Open connection to Kafka controller
	controllerAddress := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	controllerConn, err := kafka.Dial("tcp", controllerAddress)
	if err != nil {
		fmt.Printf("Failed to get dial controller: %v", err)
		return err
	}
	defer controllerConn.Close()

	fmt.Printf("Successfully connected to Kafka controller at %s\n", controllerAddress)

	// Topic config for Kafka controller
	topicConfig := kafka.TopicConfig{
		Topic:             topicName,
		NumPartitions:     1,
		ReplicationFactor: 1, // Backup copies of data existing across this cluster
	}

	// Create Kafka topics with config
	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		fmt.Printf("Failed to create topic: %v", err)
		return err
	}

	fmt.Printf("Topic is created successfully! Topic '%s' is ready for data.\n", topicName)
	return nil
}
