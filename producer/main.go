package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
)

type CrawlTask struct {
	Url string `json:"url"`
}

var (
	conn       *amqp.Connection
	ch         *amqp.Channel
	routingKey string = "task"
	queue      *RedisQueue
)

type RedisQueue struct {
	client   *redis.Client
	queueKey string
}

func NewRedisQueue(client *redis.Client, queueKey string) *RedisQueue {
	return &RedisQueue{
		client:   client,
		queueKey: queueKey,
	}
}

func (rq *RedisQueue) Enqueue(value string) error {
	_, err := rq.client.SAdd(rq.queueKey, value).Result()
	return err
}

func (rq *RedisQueue) Dequeue() (string, error) {
	value, err := rq.client.SPop(rq.queueKey).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("queue is empty")
	} else if err != nil {
		return "", err
	}

	return value, nil
}

func (rq *RedisQueue) QueueSize() (int64, error) {
	return rq.client.SCard(rq.queueKey).Result()
}

func (rq *RedisQueue) IsMember(value string) (bool, error) {
	isMember, err := rq.client.SIsMember(rq.queueKey, value).Result()
	return isMember, err
}
func init() {
	var err error
	queue = NewRedisQueue(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}), "toVisit")

	conn, err = amqp.Dial("amqp://myuser:mypassword@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		"task-exchange", // exchange name
		"direct",        // exchange type (can be direct / fanout / topic / headers)
		true,            // durable
		false,           // auto-delete
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare a direct exchange")

	// Add the first url
	queue.Enqueue("https://scrapeme.live/shop/")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func SendToTaskQueue(url string) {
	message, err := json.Marshal(CrawlTask{Url: url})
	if err != nil {
		log.Println(err)

		//todo: I could handle a dead letter queue or I do not pop it out of visited ie pushing it back to the to visit
		queue.Enqueue(url)
		return
	}
	ctx, canc := context.WithTimeout(context.Background(), 5*time.Second)
	defer canc()
	err = ch.PublishWithContext(ctx,
		"task-exchange", // exchange
		routingKey,      // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		})
	failOnError(err, "Failed to publish a message")
	if err != nil {
		log.Println(err)
		queue.Enqueue(url)
	}
	//todo: I could handle a dead letter queue or I do not pop it out of visited ie pushing it back to the to visit
}

func ReadToVisit(activityChannel chan string, done chan bool) {
	for {
		select {
		case <-done:
			return
		default:
			url, err := queue.Dequeue()
			if err != nil {
				log.Println(err)
				continue
			}
			activityChannel <- url
		}
	}
}

func main() {
	defer conn.Close()
	defer ch.Close()
	timeOut := time.Second * 30
	activityChannel := make(chan string)
	done := make(chan bool)
	timer := time.NewTimer(timeOut)
	go ReadToVisit(activityChannel, done)
	for {
		select {
		case <-timer.C:
			log.Println("Timer expired shutting down")
			close(done)
			close(activityChannel)
			return
		case url := <-activityChannel:
			go SendToTaskQueue(url)
			timer.Reset(timeOut)
			log.Println("Sent task to rabbitmq queue, URL: ", url)
		default:
			continue
		}
	}
}
