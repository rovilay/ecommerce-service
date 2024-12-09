package main

import (
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

func checkRabbit(done chan<- bool, timeInSec int, l *zerolog.Logger) {
	logger := l.With().Str("checker", "checkRabbit").Logger()

	RABBITMQ_URL, _ := os.LookupEnv("RABBITMQ_URL")

	go func() {
		count := 0
		for {
			count += 1
			_, err := amqp.Dial(RABBITMQ_URL)
			if err == nil {
				logger.Debug().Msg(fmt.Sprintf("succefully connected to rabbitmq x%d", count))
				done <- true
				return
			}

			logger.Err(err).Msg(fmt.Sprintf("failed to connect to rabbitmq x%d", count))
			time.Sleep(3 * time.Second)
		}
	}()

	// stop after a specified amount of time
	<-time.After(time.Duration(timeInSec) * time.Second)
	logger.Debug().Msg("unable to establish rabbitmq connection before timeout")
	done <- true
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).With().Str("infra", "init-checks:main").Timestamp().Logger()

	checks := 1
	done := make(chan bool, checks)

	go checkRabbit(done, 180, &logger)

	for a := 1; a <= checks; a++ {
		<-done
	}
}
