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
	slog.InfoContext(ctx, "Deleting uid record", "uid", request.UID)
	err := repo.DeleteRecord(ctx, request.UID)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	body, err := json.Marshal(map[string]string{
		"message": "Item deleted successfully",
	})

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusInternalServerError,
		Headers:    jsonHeaders,
		Body:       string(body),
	}, err
}

func main() {
	lambda.Start(handler)
}
