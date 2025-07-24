package pubsub

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"payment/internal/db"
	"payment/internal/models"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
)

type PubSub struct {
	Client        *pubsub.Client
	PaymentsTopic *pubsub.Topic
	OrderTopic    *pubsub.Topic
	OrderSub      *pubsub.Subscription
	DB            *db.DB
}

func NewClient(ctx context.Context, projectID string) (*pubsub.Client, error) {
	return pubsub.NewClient(ctx, projectID)
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

func EnsureTopic(ctx context.Context, client *pubsub.Client, topicName string) (*pubsub.Topic, error) {
	topic := client.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		topic, err = client.CreateTopic(ctx, topicName)
		if err != nil {
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
		sub, err = client.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{Topic: topic})
		if err != nil {
			return nil, err
		}
		log.Printf("Created Pub/Sub subscription: %s", subName)
	}
	return sub, nil
}

func (ps *PubSub) EnsureTopicAndSubscription(ctx context.Context) error {
	topicName := os.Getenv("PUBSUB_ORDER_TOPIC")
	if topicName == "" {
		topicName = "orders"
	}
	subName := os.Getenv("PUBSUB_ORDER_SUBSCRIPTION")
	if subName == "" {
		subName = "orders-sub"
	}
	paymentsTopicName := os.Getenv("PUBSUB_PAYMENT_TOPIC")
	if paymentsTopicName == "" {
		paymentsTopicName = "payment"
	}
	ps.PaymentsTopic = ps.Client.Topic(paymentsTopicName)
	ps.OrderTopic = ps.Client.Topic(topicName)
	ps.OrderSub = ps.Client.Subscription(subName)
	var err error
	ps.PaymentsTopic, err = EnsureTopic(ctx, ps.Client, paymentsTopicName)
	if err != nil {
		return err
	}
	ps.OrderTopic, err = EnsureTopic(ctx, ps.Client, topicName)
	if err != nil {
		return err
	}
	ps.OrderSub, err = EnsureSubscription(ctx, ps.Client, subName, topicName)
	return err
}

func (ps *PubSub) ListenForOrderEvents(ctx context.Context) {
	err := ps.OrderSub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var orderEvent struct {
			OrderID string `json:"id"`
			Amount  int    `json:"amount"`
		}
		if err := json.Unmarshal(msg.Data, &orderEvent); err != nil {
			log.Printf("Invalid order event: %v", err)
			msg.Nack()
			return
		}
		log.Printf("Received order event: %+v", orderEvent)
		time.Sleep(2 * time.Second)
		payment := models.Payment{
			TransactionID: uuid.NewString(),
			OrderID:       orderEvent.OrderID,
			Status:        "paid",
			Amount:        orderEvent.Amount,
			CreatedAt:     time.Now(),
		}
		if err := ps.DB.InsertPayment(payment); err != nil {
			log.Printf("Failed to insert payment: %v", err)
		} else {
			log.Printf("Payment processed and stored: %+v", payment)
		}
		paymentEvent, err := json.Marshal(payment)
		if err != nil {
			log.Printf("Failed to marshal payment event: %v", err)
		} else {
			result := ps.PaymentsTopic.Publish(ctx, &pubsub.Message{Data: paymentEvent})
			_, pubErr := result.Get(ctx)
			if pubErr != nil {
				log.Printf("Failed to publish payment event: %v", pubErr)
			} else {
				log.Printf("Published payment event for order %s", payment.OrderID)
			}
		}
		msg.Ack()
	})
	if err != nil {
		log.Fatalf("Error receiving messages: %v", err)
	}
}
