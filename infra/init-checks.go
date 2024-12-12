package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func checkRabbit(done chan<- bool, timeInSec int, l *zerolog.Logger) {
	logger := l.With().Str("checker", "checkRabbit").Logger()

	RABBITMQ_URL, _ := os.LookupEnv("RABBITMQ_URL")

	go func() {
		count := 0
		for {
			count += 1
			conn, err := amqp.Dial(RABBITMQ_URL)
			if err == nil {
				conn.Close()
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
	logger.Debug().Msg("unable to establish rabbitmq connection before timeout!")
	done <- true
}

func checkRedis(done chan<- bool, timeInSec int, l *zerolog.Logger) {
	logger := l.With().Str("checker", "checkRedis").Logger()

	REDIS_URL, _ := os.LookupEnv("REDIS_URL")

	go func() {
		count := 0
		for {
			count += 1

			// connect to redis
			conn := redis.NewClient(&redis.Options{
				Addr: REDIS_URL,
			})
			err := conn.Ping(context.Background()).Err()
			if err == nil {
				logger.Debug().Msg(fmt.Sprintf("succefully connected to redis x%d", count))
				conn.Close()
				done <- true
				return
			}

			logger.Err(err).Msgf("failed to connect to redis x%d", count)
			time.Sleep(3 * time.Second)
		}
	}()

	// stop after a specified amount of time
	<-time.After(time.Duration(timeInSec) * time.Second)
	logger.Debug().Msg("unable to establish rabbitmq connection before timeout!")
	done <- true
}

type Checks string

const (
	RABBIT Checks = "rabbit"
	REDIS  Checks = "redis"
)

func getChecks() (map[string]bool, int) {
	validArgsSet := make(map[string]bool)
	validArgsSet[string(RABBIT)] = true
	validArgsSet[string(REDIS)] = true

	checksCount := 2

	var exclude string
	flag.StringVar(&exclude, "exclude", "", "Comma-separated list of checks to exclude: options ['rabbit', 'redis']")
	flag.Parse()

	argsInput := strings.Split(exclude, ",")
	for _, arg := range argsInput {
		if validArgsSet[strings.ToLower(arg)] {
			checksCount -= 1
			validArgsSet[strings.ToLower(arg)] = false
		}
	}

	return validArgsSet, checksCount
}

func main() {
	argSet, checks := getChecks()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).With().Str("infra", "init-checks:main").Timestamp().Logger()

	done := make(chan bool, checks)

	if argSet[string(RABBIT)] {
		go checkRabbit(done, 180, &logger)
	}

	if argSet[string(REDIS)] {
		go checkRedis(done, 180, &logger)
	}

	for a := 1; a <= checks; a++ {
		<-done
	}

	logger.Debug().Msgf("%d checks completed", checks)
}
