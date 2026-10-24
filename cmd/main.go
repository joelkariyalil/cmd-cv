package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joelkariyalil/cmd-cv/internal/generate"
	"github.com/joelkariyalil/cmd-cv/internal/model"
	"github.com/joelkariyalil/cmd-cv/internal/persistance"
)

const (
	TTL                time.Duration = time.Minute * 10
	OtcSize                          = 6 // one time code
	MaxConflictRetries               = 5
)

//nolint:gochecknoglobals // lambda startup optimization
var (
	repo        persistance.Repo
	jsonHeaders = map[string]string{
		"Content-Type": "application/json",
	}
)

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

func createHandler(ctx context.Context, req events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	request := model.CreateRequest{}
	err := json.Unmarshal([]byte(req.Body), &request)
	if err != nil {
		slog.WarnContext(ctx, "Invalid create request payload", "error", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    jsonHeaders,
			Body:       `{"message":"Invalid request payload"}`,
		}
	}

	slog.InfoContext(ctx, "Creating new record", "message_length", len(request.Message))

	for i := range MaxConflictRetries {
		uid, err := generate.GenerateCode(OtcSize)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to generate UID", "error", err)
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Headers:    jsonHeaders,
				Body:       `{"message":"Failed to generate UID"}`,
			}
		}

		err = repo.CreateRecord(ctx, uid, request.Message, TTL)
		if err == nil {
			body, err := json.Marshal(map[string]string{
				"uid":     uid,
				"message": "Item created successfully",
			})
			if err != nil {
				slog.ErrorContext(ctx, "Failed to serialize create response", "error", err)
				return events.APIGatewayV2HTTPResponse{
					StatusCode: http.StatusInternalServerError,
					Headers:    jsonHeaders,
					Body:       `{"message":"Failed to serialize response"}`,
				}
			}

			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusCreated,
				Headers:    jsonHeaders,
				Body:       string(body),
			}
		}

		if errors.Is(err, persistance.ErrRecordAlreadyExists) {
			slog.InfoContext(ctx, "Generated duplicate UID, retrying",
				"uid", uid,
				"attempt", i+1,
			)
			continue
		}

		slog.ErrorContext(ctx, "Failed to create record",
			"uid", uid,
			"error", err,
		)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    jsonHeaders,
			Body:       `{"message":"Failed to create record"}`,
		}
	}

	slog.ErrorContext(ctx, "Exhausted UID conflict retries",
		"max_attempts", MaxConflictRetries,
	)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusInternalServerError,
		Headers:    jsonHeaders,
		Body:       `{"message":"Failed to create record"}`,
	}
}

func deleteHandler(ctx context.Context, req events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	request := model.DeleteRequest{}
	err := json.Unmarshal([]byte(req.Body), &request)
	if err != nil {
		slog.WarnContext(ctx, "Invalid delete request payload", "error", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    jsonHeaders,
			Body:       `{"message":"Invalid request payload"}`,
		}
	}

	slog.InfoContext(ctx, "Deleting UID record", "uid", request.UID)

	err = repo.DeleteRecord(ctx, request.UID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete record",
			"uid", request.UID,
			"error", err,
		)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    jsonHeaders,
			Body:       `{"message":"Failed to delete record"}`,
		}
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers:    jsonHeaders,
		Body:       `{"message":"Item deleted successfully"}`,
	}
}

func retrieveHandler(ctx context.Context, req events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	id := req.PathParameters["id"]

	if id == "" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    jsonHeaders,
			Body:       `{"message":"Missing required parameter: id"}`,
		}
	}

	slog.InfoContext(ctx, "Retrieving data for ID record", "id", id)

	rec, err := repo.ReadRecord(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve record",
			"id", id,
			"error", err,
		)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    jsonHeaders,
			Body:       `{"message":"Failed to retrieve record"}`,
		}
	}

	if rec == nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusNotFound,
			Headers:    jsonHeaders,
			Body:       `{"message":"Record not found"}`,
		}
	}

	body, err := json.Marshal(rec)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to serialize retrieved record",
			"id", id,
			"error", err,
		)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    jsonHeaders,
			Body:       `{"message":"Failed to serialize record"}`,
		}
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers:    jsonHeaders,
		Body:       string(body),
	}
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	switch {
	case req.RawPath == "/create" && req.RequestContext.HTTP.Method == http.MethodPost:
		return createHandler(ctx, req)

	case req.RawPath == "/delete" && req.RequestContext.HTTP.Method == http.MethodDelete:
		return deleteHandler(ctx, req)

	case strings.HasPrefix(req.RawPath, "/retrieve/") && req.RequestContext.HTTP.Method == http.MethodGet:
		return retrieveHandler(ctx, req)

	default:
		return events.APIGatewayV2HTTPResponse{
			Headers:    jsonHeaders,
			StatusCode: http.StatusNotFound,
			Body:       `{"message":"Route not found"}`,
		}
	}
}

func main() {
	lambda.Start(handler)
}
