package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joelkariyalil/cmd-cv/internal/persistance"
)

//nolint:gochecknoglobals // lambda startup optimization
var (
	repo        persistance.Repo
	jsonHeaders = map[string]string{
		"Content-Type": "application/json",
	}
)

type Request struct {
	UID string `json:"uid"`
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
	slog.InfoContext(ctx, "Retrieving data for uid record", "uid", request.UID)

	rec, err := repo.ReadRecord(ctx, request.UID)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    jsonHeaders,
			Body:       `{"message":"Failed to retrieve record"}`,
		}, err
	}

	if rec == nil {
		body, err := json.Marshal(map[string]string{
			"message": "Record not found",
		})
		if err != nil {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
			}, err
		}

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusNotFound,
			Headers:    jsonHeaders,
			Body:       string(body),
		}, nil
	}

	body, err := json.Marshal(rec)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    jsonHeaders,
			Body:       `{"message":"Failed to serialize record"}`,
		}, err
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers:    jsonHeaders,
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
