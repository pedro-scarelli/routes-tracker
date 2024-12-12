package main

import (
	"context"
	"fmt"
	"log"

	"github.com/routes/simulator/internal"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main () {
	mongoStr := "mongodb://admin:admin@localhost:27017/nest?authSource=admin"
	mongoConnection, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoStr))
	if err != nil {
		panic(err)
	}	

	freightService := internal.NewFreightService()
	routeService := internal.NewRouteService(mongoConnection, freightService)

	channelDriverMoved := make(chan *internal.DriverMovedEvent)
	kafkaBroker := "localhost:9092"

	freightWriter := &kafka.Writer{
		Addr: kafka.TCP(kafkaBroker),
		Topic: "freight",
		Balancer: &kafka.LeastBytes{},
	}
	
	simulatorWriter := &kafka.Writer{
		Addr: kafka.TCP(kafkaBroker),
		Topic: "simulator",
		Balancer: &kafka.LeastBytes{},
	}

	routeReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic: "route",
		GroupID:  "simulator",
	})

	hub := internal.NewEventHub(routeService, mongoConnection, channelDriverMoved, freightWriter, simulatorWriter)

	fmt.Println("Starting simulator")
	for {
		message, err := routeReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}

		go func(msg []byte) {
			err = hub.HandleEvent(message.Value)
			if err != nil {
				log.Printf("error handling event: %v", err)
			}
		}(message.Value)
	}
}