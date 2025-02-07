package main

import (
	"consumer/fetcher"
	"consumer/header"
	"consumer/parser"
	"consumer/proxy"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/gocolly/colly"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	conn           *amqp.Connection
	ch             *amqp.Channel
	msgs           <-chan amqp.Delivery = make(<-chan amqp.Delivery)
	visiting       *RedisQueue
	visited        *RedisQueue
	toVisit        *RedisQueue
	MAXNumber      int64 = 500
	scrapeMeParser parser.Parser
	wg             *sync.WaitGroup = &sync.WaitGroup{}
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

type CrawlTask struct {
	Url string `json:"url"`
}

func init() {
	var err error
	conn, err = amqp.Dial("amqp://myuser:mypassword@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")

	queue, err := ch.QueueDeclare(
		"task-queue",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	ch.QueueBind(
		queue.Name,
		"task",
		"task-exchange",
		false,
		nil,
	)

	msgs, err = ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	visiting = NewRedisQueue(rdb, "visiting")
	visited = NewRedisQueue(rdb, "visited")
	toVisit = NewRedisQueue(rdb, "toVisit")

	scrapeMeParser = parser.NewScrapemeParser()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
func main() {
	defer conn.Close()
	defer ch.Close()
	for msg := range msgs {
		log.Println("Received a message", string(msg.Body))
		var task CrawlTask
		err := json.Unmarshal(msg.Body, &task)
		if err != nil {
			log.Println(err)
			continue
		}

		visitedCount, _ := visited.QueueSize()
		visitingCount, _ := visiting.QueueSize()

		if visitingCount+visitedCount >= MAXNumber {
			log.Println("Max number of tasks reached")
			wg.Wait()
			return
		}

		isBeingVisited, _ := visiting.IsMember(task.Url)
		hasBeenVisited, _ := visited.IsMember(task.Url)

		if isBeingVisited || hasBeenVisited {
			log.Println("Already visited or being visited")
			continue
		}

		visiting.Enqueue(task.Url)
		go TaskHandler(task)
		domain := extractDomain(task.Url)
		if strings.Contains(domain, "scrapeme.live") {
			log.Println("Scrapeme domain found")
			wg.Add(1)
			go handleScrapeMeDomain(wg, task.Url)
		}

	}
}

func handleScrapeMeDomain(wg *sync.WaitGroup, url string) {
	defer wg.Done()
	log.Println("Handling scrape me domain")
	headerGen := header.NewHeaderGenFunc()
	proxyGen := proxy.NewProxyGenFunc()
	if !scrapeMeParser.IsValidURL(url) {
		log.Println("Invalid url")
		return
	}
	content, err := scrapeMeParser.GetHTML(url, headerGen, proxyGen, fetcher.NewCollyFetcher())
	if err != nil {
		log.Println(err)
		return
	}
	rawContent, err := scrapeMeParser.ExtractContent(content)
	if err != nil {
		log.Println(err)
		return
	}
	content, err = json.Marshal(rawContent)
	if err != nil {
		log.Println(err)
		return
	}
	err = scrapeMeParser.StoreContent(content)
	if err != nil {
		log.Println(err)
		return
	}
}

func extractDomain(givenURL string) string {
	url, err := url.Parse(givenURL)
	if err != nil {
		log.Println(err)
		return ""
	}
	return url.Hostname()
}

func TaskHandler(task CrawlTask) {
	log.Println("Handling task")
	links, err := ExtractLinks(task)
	if err != nil {
		log.Println(err)
		return
	}
	visited.Enqueue(task.Url)
	visiting.Dequeue()
	for _, link := range links {
		visitedNewLink, _ := visited.IsMember(link)
		visitingNewLink, _ := visiting.IsMember(link)
		if visitedNewLink || visitingNewLink || strings.Contains(link, "#") || link == "" {
			continue
		}
		toVisit.Enqueue(link)
		log.Println("Enqueued link", link)
	}
}

func ExtractLinks(task CrawlTask) ([]string, error) {
	log.Println("Extracting links")
	startTime := time.Now()
	url := task.Url
	c := colly.NewCollector()
	links := make([]string, 0)

	c.OnHTML("a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		link = e.Request.AbsoluteURL(link)
		if strings.Contains(e.Attr("rel"), "nofollow") {
			return
		}
		links = append(links, link)
	})

	if err := c.Visit(url); err != nil {
		log.Println(err)
		return nil, err
	}

	elapsed := time.Since(startTime)
	log.Printf("Extracted %d links in %s", len(links), elapsed)
	return links, nil
}
