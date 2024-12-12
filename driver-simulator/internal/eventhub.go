package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventHub struct {
	routeService *RouteService
	mongoCliente *mongo.Client
	channelDriverMoved chan *DriverMovedEvent
	freightWriter *kafka.Writer
	simulatorWriter *kafka.Writer
}

func NewEventHub(routeService *RouteService, mongoClient *mongo.Client, channelDriverMoved chan *DriverMovedEvent, freightWriter *kafka.Writer, simulatorWriter *kafka.Writer) *EventHub {
	return &EventHub{
		routeService: routeService,
		mongoCliente: mongoClient,
		channelDriverMoved: channelDriverMoved,
		freightWriter: freightWriter,
		simulatorWriter: simulatorWriter,
	}
}

func (eh *EventHub) HandleEvent(message []byte) error {
	var baseEvent struct {
		EventName string `json:"event"`
	}
	err := json.Unmarshal(message, &baseEvent) 
	if err != nil {
		return fmt.Errorf("error unmarshalling event: %w", err)
	}

	switch baseEvent.EventName {
		case "RouteCreated":
			var event  RouteCreatedEvent
			err := json.Unmarshal(message, &event)
			if err != nil {
				return fmt.Errorf("error unmarshalling event: %w", err)
			}
			return eh.handleRouteCreated(event)
		case "DeliveryStarted":
			var event  DeliveryStartedEvent
			err := json.Unmarshal(message, &event)
			if err != nil {
				return fmt.Errorf("error unmarshalling event: %w", err)
			}
		default:
			return errors.New("unknown event")
	}
	return nil
}

func (eh EventHub) handleRouteCreated(event RouteCreatedEvent) error {
	freightCalculatedEvent, err := RouteCreatedHandler(&event, eh.routeService)
	if err != nil {
		return err
	}
	value, err := json.Marshal(freightCalculatedEvent)
	if err != nil {
		return fmt.Errorf("error marshalling event: %w", err)
	}
	err = eh.freightWriter.WriteMessages(context.Background(), kafka.Message{
		Key: []byte(freightCalculatedEvent.RouteID),
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}	
	return nil
}

func (eh EventHub) handleDeliveryStarted(event DeliveryStartedEvent) error {
	err := DeliveryStartedHandler(&event, eh.routeService, eh.channelDriverMoved)
	if err != nil {
		return err
	}
	go eh.sendDirections()

	return nil
}

func (eh *EventHub) sendDirections() {
	for {
		select {
			case movedEvent := <- eh.channelDriverMoved:
			value, err := json.Marshal(movedEvent)
			if err != nil {
				return
			}
			err = eh.simulatorWriter.WriteMessages(context.Background(), kafka.Message{
				Key: []byte(movedEvent.RouteID),
				Value: value,
			})
			if err != nil {
				return
			}
			case <- time.After(500 * time.Millisecond):
				return
		}
	}
}