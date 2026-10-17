package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joelkariyalil/cmd-cv/internal/generate"
	"github.com/joelkariyalil/cmd-cv/internal/persistance"
)

const (
	ttl     time.Duration = time.Minute * 10
	otcSize               = 6 // one time code
)

//nolint:gochecknoglobals // lambda startup optimization
var (
	repo        persistance.Repo
	jsonHeaders = map[string]string{
		"Content-Type": "application/json",
	}
)

type Request struct {
	Message string `json:"message"`
}

//nolint:gochecknoinits // lambda startup optimization
func init() {
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		panic("TABLE_NAME environment variable must be set")
	}

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	dbClient := dynamodb.NewFromConfig(cfg)

	repo = persistance.NewDbTable(dbClient, tableName)
}

func handler(ctx context.Context, request Request) (events.APIGatewayV2HTTPResponse, error) {
	slog.InfoContext(ctx, "Creating new record", "message_length", len(request.Message))

	for {
		uid, err := generate.GenerateCode(otcSize)
		if err != nil {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Headers:    jsonHeaders,
				Body:       `{"message":"Failed to generate UID"}`,
			}, err
		}

		err = repo.CreateRecord(ctx, uid, request.Message, ttl)
		if err == nil {
			body, err := json.Marshal(map[string]string{
				"uid":     uid,
				"message": "Item created successfully",
			})
			if err != nil {
				return events.APIGatewayV2HTTPResponse{
					StatusCode: http.StatusInternalServerError,
				}, err
			}

			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusCreated,
				Headers:    jsonHeaders,
				Body:       string(body),
			}, nil
		}

		if errors.Is(err, persistance.ErrRecordAlreadyExists) {
			slog.InfoContext(ctx, "Generated duplicate UID, retrying", "uid", uid)
			continue
		}

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    jsonHeaders,
			Body:       `{"message":"Failed to create record"}`,
		}, err
	}
}

func main() {
	lambda.Start(handler)
}
