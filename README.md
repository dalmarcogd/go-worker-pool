
![Build](https://github.com/dalmarcogd/gwp/workflows/Go/badge.svg)
[![codecov](https://codecov.io/gh/dalmarcogd/gwp/branch/master/graph/badge.svg)](https://codecov.io/gh/dalmarcogd/gwp)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/beee52f22195471abea544a19ee6304a)](https://www.codacy.com/manual/dalmarco.gd/gwp?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=dalmarcogd/gwp&amp;utm_campaign=Badge_Grade)
[![Go Report Card](https://goreportcard.com/badge/github.com/dalmarcogd/gwp)](https://goreportcard.com/report/github.com/dalmarcogd/gwp)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/dalmarcogd/gwp)
[![license](https://img.shields.io/hexpm/l/apa)](https://pkg.go.dev/github.com/dalmarcogd/gwp/blob/master/LICENSE.md)
[![Release](https://img.shields.io/github/v/release/dalmarcogd/gwp.svg?style=flat-square)](https://github.com/dalmarcogd/gwp/releases)
[![Go version](https://img.shields.io/badge/go-%5E1.14-blue)](https://golang.org/dl/)

# gwp - Go Worker Pool

This package wants to offer the community and implement workers with the pure Go code for Golangers, without any other dependency just Uuid. It allows you to expose an http server to answer the response of health checks, stats, debug pprof and the main "workers". Workers for consumer queues, channel processes and other things that you think worker needs.

![image](/golang-worker.png)


## Prerequisites
Golang version >= [1.14](https://golang.org/doc/devel/release.html#go1.14)

## Features
- Setup http server to monitoring yours;
  - /stats with workers, showing statuses her, number of goroutines, number of cpus and more;
  - /health-check that look for status of workers;
  - /debug/pprof expose all endpoints of investivate golang runtime [http](https://golang.org/pkg/net/http/pprof/);
- Allow multiple concurrencies of work, handle errors and restart always worker;

## Documentation
For examples visit godoc#pkg-examples

For GoDoc reference, visit [pkg.go.dev](https://pkg.go.dev/github.com/dalmarcogd/gwp)

## Examples

#### [Simple Worker](https://github.com/dalmarcogd/test-go-worker-pool/blob/master/simpleWorker/simpleWorker.go) ###

```go
package main

import (
	"errors"
	"github.com/dalmarcogd/gwp"
	"github.com/dalmarcogd/gwp/pkg/worker"
	"log"
	"time"
)

func main() {
	if err := gwp.
		New().
		Stats().
		HealthCheck().
		DebugPprof().
		HandleError(func(w *worker.Worker, err error) {
			log.Printf("Worker [%s] error: %s", w.Name, err)
		}).
		Worker(
			"w1",
			func() error {
				<-time.After(10 * time.Second)
				return errors.New("test")
			},
			worker.WithRestartAlways()).
		Worker(
			"w2",
			func() error {
				<-time.After(30 * time.Second)
				return nil
			}).
		Worker(
			"w3",
			func() error {
				<-time.After(1 * time.Minute)
				return errors.New("test")
			}).
		Run(); err != nil {
		panic(err)
	}
}
```

#### [Simple Worker Consume Channel](https://github.com/dalmarcogd/test-gwp/blob/master/simpleWorkerChannels/simpleWorkerChannels.go) ###
```go
package main

import (
	"github.com/dalmarcogd/gwp"
	"github.com/dalmarcogd/gwp/pkg/worker"
	"log"
	"time"
)

func main() {

	ch := make(chan bool, 1)

	if err := gwp.
		New().
		Stats().
		HealthCheck().
		DebugPprof().
		HandleError(func(w *worker.Worker, err error) {
			log.Printf("Worker [%s] error: %s", w.Name, err)
		}).
		Worker(
			"w1",
			func() error {
				<-time.After(10 * time.Second)
				ch <- true
				log.Printf("Produced %t", true)
				return nil
			},
			1,
			true).
		Worker(
			"w2",
			func() error {
				for {
					select {
					case r := <-ch:
						log.Printf("Received %t", r)
					}
				}
			},
			1,
			false).
		Run(); err != nil {
		panic(err)
	}
}
```

#### [Simple Worker Consume Buffered Channel](https://github.com/dalmarcogd/test-gwp/blob/master/simpleWorkerBufferedChannels/simpleWorkerBufferedChannels.go) ###

```go
package main

import (
	"github.com/dalmarcogd/gwp"
	"github.com/dalmarcogd/gwp/pkg/worker"
	"log"
	"time"
)

func main() {

	numberOfConcurrency := 10
	ch := make(chan bool, numberOfConcurrency)

	if err := gwp.
		New().
		Stats().
		HealthCheck().
		DebugPprof().
		HandleError(func(w *worker.Worker, err error) {
			log.Printf("Worker [%s] error: %s", w.Name, err)
		}).
		Worker(
			"w1",
			func() error {
				<-time.After(10 * time.Second)
				ch <- true
				ch <- true
				ch <- true
				ch <- true
				ch <- true
				ch <- true
				ch <- true
				log.Printf("Produced %t", true)
				return nil
			},

			worker.WithRestartAlways()).
		Worker(
			"w2",
			func() error {
				for {
					select {
					case r := <-ch:
						log.Printf("Received %t", r)
					}
				}
			},
			worker.WithConcurrency(numberOfConcurrency)).
		Run(); err != nil {
		panic(err)
	}
}
```

#### [Simple Worker Consume SQS](https://github.com/dalmarcogd/test-gwp/blob/master/simpleWorkerConsumeSQS/simpleWorkerConsumeSQS.go) ###
```go
package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/dalmarcogd/gwp"
	"github.com/dalmarcogd/gwp/pkg/worker"
	"log"
	"strconv"
)

func main() {

	params := &sqs.CreateQueueInput{
		QueueName: aws.String("test-consume-sqs"), // Required
	}
	ss, _ := session.NewSession(&aws.Config{
		Endpoint: aws.String("http://localhost:9324"),
		Region:   aws.String("us-east-1"),
	})
	svc := sqs.New(ss)

	var resp, err = svc.CreateQueue(params)

	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	fmt.Println(resp)

	queueURL := aws.String("http://localhost:9324/queue/test-consume-sqs")

	for i := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
		paramsSend := &sqs.SendMessageInput{
			MessageBody: aws.String("Testing " + strconv.Itoa(i)), // Required
			QueueUrl:    queueURL,                                 // Required
		}
		respSend, err := svc.SendMessage(paramsSend)
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		fmt.Println(respSend)
	}

	if err := gwp.
		New().
		Stats().
		HealthCheck().
		DebugPprof().
		HandleError(func(w *worker.Worker, err error) {
			log.Printf("Worker [%s] error: %s", w.Name, err)
		}).
		Worker("w2", func() error {
			params := &sqs.ReceiveMessageInput{
				QueueUrl:            queueURL, // Required
				MaxNumberOfMessages: aws.Int64(10),
				VisibilityTimeout:   aws.Int64(20),
			}
			resp, err := svc.ReceiveMessage(params)

			if err != nil {
				fmt.Println(err.Error())
				return err
			}
			fmt.Println(resp.Messages)
			for _, msg := range resp.Messages {
				fmt.Println(aws.StringValue(msg.Body))
			}
			return nil
		}, worker.WithRestartAlways()).
		Run(); err != nil {
		panic(err)
	}
}
```

#### [Simple Worker Consume Rabbit](https://github.com/dalmarcogd/test-go-worker-pool/blob/master/simpleWorkerConsumeRabbit/simpleWorkerConsumeRabbit.go) ###
```go
package main

import (
	"fmt"
	"github.com/dalmarcogd/gwp"
	"github.com/dalmarcogd/gwp/pkg/worker"
	"github.com/streadway/amqp"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {

	connection, err := amqp.Dial("amqp://rabbitmq:rabbitmq@localhost:5672//")

	failOnError(err, "Error when get connection")
	defer connection.Close()

	channel, err := connection.Channel()
	failOnError(err, "Error when get channel")
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		"test-consume", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Error when declare a queue")

	for i := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
		failOnError(channel.Publish("", queue.Name, false, false, amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Body:         []byte(fmt.Sprint(i)),
		}), "fail on publishing")
	}

	if err := gwp.
		New().
		Stats().
		HealthCheck().
		DebugPprof().
		HandleError(func(w *worker.Worker, err error) {
			log.Printf("Worker [%s] error: %s", w.Name, err)
		}).
		Worker("w2", func() error {
			msgs, err := channel.Consume(queue.Name,
				"",
				true,
				false,
				false,
				false,
				nil)
			failOnError(err, "Error when create consumer")

			for msg := range msgs {
				fmt.Println(string(msg.Body))
			}
			return nil
		}, worker.WithRestartAlways()).
		Run(); err != nil {
		panic(err)
	}
}
```

#### [Simple Worker Consume Kafka](https://github.com/dalmarcogd/test-gwp/blob/master/simpleWorkerConsumeKafka/simpleWorkerConsumeKafka.go) ###
```go
package main

import (
	"context"
	"fmt"
	"github.com/dalmarcogd/gwp"
	"github.com/dalmarcogd/gwp/pkg/worker"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	topic := "teste"
	partition := 1

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	failOnError(err, "Fail when create connection")

	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, _ = conn.WriteMessages(
		kafka.Message{Value: []byte("one!")},
		kafka.Message{Value: []byte("two!")},
		kafka.Message{Value: []byte("three!")},
	)

	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	if err := gwp.
		New().
		Stats().
		HealthCheck().
		DebugPprof().
		HandleError(func(w *worker.Worker, err error) {
			log.Printf("Worker [%s] error: %s", w.Name, err)
		}).
		Worker("w2", func() error {
			batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max
			b := make([]byte, 10e3)            // 10KB max per message
			for {
				_, err := batch.Read(b)
				if err != nil {
					break
				}
				fmt.Println(string(b))
			}

			_ = batch.Close()
			return nil
		},  worker.WithRestartAlways()).
		Run(); err != nil {
		panic(err)
	}
}
```
