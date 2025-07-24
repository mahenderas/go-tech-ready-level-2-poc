package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"orders/internal/db"
	"orders/internal/models"

	"cloud.google.com/go/pubsub"
)

// PubSub struct holds references to Pub/Sub client, topics, subscriptions, and DB.
type PubSub struct {
	Client       *pubsub.Client
	PaymentTopic *pubsub.Topic
	OrdersTopic  *pubsub.Topic
	PaymentSub   *pubsub.Subscription
	DB           *db.DB
}

func NewClient(ctx context.Context, projectID string) (*pubsub.Client, error) {
	return pubsub.NewClient(ctx, projectID)
}

// EnsureTopic checks if a Pub/Sub topic exists, and creates it if not.
func EnsureTopic(ctx context.Context, client *pubsub.Client, topicName string) (*pubsub.Topic, error) {
	topic := client.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		topic, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			if strings.Contains(err.Error(), "AlreadyExists") {
				log.Printf("Topic %s already exists (race condition)", topicName)
				return topic, nil
			}
			return nil, err
		}
		log.Printf("Created Pub/Sub topic: %s", topicName)
	}
	return topic, nil
}

// EnsureSubscription checks if a Pub/Sub subscription exists, and creates it if not.
func EnsureSubscription(ctx context.Context, client *pubsub.Client, subName, topicName string) (*pubsub.Subscription, error) {
	sub := client.Subscription(subName)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		topic := client.Topic(topicName)
		// Ensure topic exists before creating subscription
		if _, err := EnsureTopic(ctx, client, topicName); err != nil {
			return nil, err
		}
		sub, err = client.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{Topic: topic})
		if err != nil {
			if strings.Contains(err.Error(), "AlreadyExists") {
				log.Printf("Subscription %s already exists (race condition)", subName)
				return sub, nil
			}
			if strings.Contains(err.Error(), "NotFound") {
				return nil, fmt.Errorf("subscription topic does not exist: %w", err)
			}
			return nil, err
		}
		log.Printf("Created Pub/Sub subscription: %s", subName)
	}
	return sub, nil
}

func SetupPubSub(ctx context.Context, projectID string, dbInstance *db.DB) (*PubSub, error) {
	client, err := NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	ps := &PubSub{
		Client: client,
		DB:     dbInstance,
	}

	if err := ps.EnsureTopicAndSubscription(ctx); err != nil {
		return nil, err
	}

	return ps, nil
}

func (ps *PubSub) EnsureTopicAndSubscription(ctx context.Context) error {
	topicName := os.Getenv("PUBSUB_PAYMENT_TOPIC")
	if topicName == "" {
		topicName = "payment"
	}
	subName := os.Getenv("PUBSUB_PAYMENT_SUBSCRIPTION")
	if subName == "" {
		subName = "payment-sub"
	}
	ordersTopicName := os.Getenv("PUBSUB_ORDERS_TOPIC")
	if ordersTopicName == "" {
		ordersTopicName = "orders"
	}
	ps.PaymentTopic = ps.Client.Topic(topicName)
	ps.OrdersTopic = ps.Client.Topic(ordersTopicName)
	ps.PaymentSub = ps.Client.Subscription(subName)
	var err error
	ps.PaymentTopic, err = EnsureTopic(ctx, ps.Client, topicName)
	if err != nil {
		return err
	}
	ps.OrdersTopic, err = EnsureTopic(ctx, ps.Client, ordersTopicName)
	if err != nil {
		return err
	}
	ps.PaymentSub, err = EnsureSubscription(ctx, ps.Client, subName, topicName)
	return err
}

func (ps *PubSub) ListenForPaymentEvents(ctx context.Context) {
	err := ps.PaymentSub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var paymentEvent struct {
			OrderID string `json:"order_id"`
			Status  string `json:"status"`
			Amount  int    `json:"amount"`
		}
		if err := json.Unmarshal(msg.Data, &paymentEvent); err != nil {
			log.Printf("Invalid payment event: %v", err)
			msg.Nack()
			return
		}
		log.Printf("Received payment event: %+v", paymentEvent)
		// Update order status in DB based on payment event
		if err := ps.DB.UpdateOrderStatus(paymentEvent.OrderID, paymentEvent.Status); err != nil {
			log.Printf("Failed to update order status: %v", err)
			msg.Nack()
			return
		}
		msg.Ack()
	})
	if err != nil {
		log.Printf("Error receiving messages: %v", err)
	}
}

// add a method to publish order events to Pub/Sub
func PublishOrderEvent(ctx context.Context, ps *PubSub, order models.Order) {
	orderEvent := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount int    `json:"amount"`
	}{
		ID:     order.ID,
		Status: order.Status,
		Amount: order.Amount,
	}

	data, err := json.Marshal(orderEvent)
	if err != nil {
		log.Printf("Failed to marshal order event: %v", err)
		return
	}

	msg := &pubsub.Message{
		Data: data,
	}

	result := ps.OrdersTopic.Publish(ctx, msg)
	if _, err := result.Get(ctx); err != nil {
		log.Printf("Failed to publish order event: %v", err)
	} else {
		log.Printf("Published order event for order ID: %s", order.ID)
	}
}
