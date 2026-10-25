package persistance_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/joelkariyalil/cmd-cv/internal/generate"
	"github.com/joelkariyalil/cmd-cv/internal/persistance"
	"github.com/stretchr/testify/suite"
)

type dynamodbMock struct {
	PutItemMock    func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	DeleteItemMock func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	UpdateItemMock func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

func (m *dynamodbMock) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return m.PutItemMock(ctx, params, optFns...)
}

func (m *dynamodbMock) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return m.DeleteItemMock(ctx, params, optFns...)
}

func (m *dynamodbMock) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return m.UpdateItemMock(ctx, params, optFns...)
}

type repoSuite struct {
	suite.Suite
	repo *persistance.DbTable
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, &repoSuite{})
}

//nolint:errcheck //disabling it as it's a test function, and we cognizant of the values passed in
func (s *repoSuite) TestCreateRecord() {
	// given
	otcSize := 6
	expectedUID, _ := generate.GenerateCode(otcSize)
	expectedMessage := "hello-world"
	expectedTTL := time.Minute * 10
	startTime := time.Now().Unix()

	var capturedItem map[string]types.AttributeValue

	dbMock := &dynamodbMock{
		PutItemMock: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			capturedItem = params.Item
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	s.repo = persistance.NewDbTable(dbMock, "mock-table")

	// when
	err := s.repo.CreateRecord(context.Background(), expectedUID, expectedMessage, expectedTTL)

	// then
	s.Require().NoError(err)
	s.Require().Equal(expectedUID, capturedItem["PK"].(*types.AttributeValueMemberS).Value)
	s.Require().Equal(expectedMessage, capturedItem["message"].(*types.AttributeValueMemberS).Value)
	s.Require().NotNil(capturedItem["created_at"].(*types.AttributeValueMemberN).Value)

	createdAt, err := convertToIntFromStr(capturedItem["created_at"].(*types.AttributeValueMemberN).Value)
	s.Require().NoError(err)

	s.Require().NotNil(capturedItem["last_accessed_at"].(*types.AttributeValueMemberN).Value)

	lastAccessedAt, err := convertToIntFromStr(capturedItem["last_accessed_at"].(*types.AttributeValueMemberN).Value)
	s.Require().NoError(err)

	expiresAt, err := convertToIntFromStr(capturedItem["expires_at"].(*types.AttributeValueMemberN).Value)
	s.Require().NoError(err)

	s.Require().GreaterOrEqual(createdAt, startTime)
	s.Require().GreaterOrEqual(lastAccessedAt, startTime)

	expectedExpiry := startTime + int64(expectedTTL.Seconds())
	s.Require().GreaterOrEqual(expiresAt, expectedExpiry)
}

//nolint:errcheck //disabling it as it's a test function, and we cognizant of the values passed in
func (s *repoSuite) TestDeleteRecord() {
	// given
	otcSize := 6
	expectedUID, _ := generate.GenerateCode(otcSize)
	var capturedItem map[string]types.AttributeValue

	dbMock := &dynamodbMock{
		DeleteItemMock: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
			capturedItem = params.Key
			return &dynamodb.DeleteItemOutput{}, nil
		},
	}

	// when
	s.repo = persistance.NewDbTable(dbMock, "mock-table")
	err := s.repo.DeleteRecord(context.Background(), expectedUID)

	// then
	s.Require().NoError(err)
	s.Require().Equal(expectedUID, capturedItem["PK"].(*types.AttributeValueMemberS).Value)
}

func (s *repoSuite) TestReadRecord() {
	// given
	otcSize := 6
	expectedUID, _ := generate.GenerateCode(otcSize)
	expectedMessage := "hello-world"

	var capturedParams *dynamodb.UpdateItemInput

	dbMock := &dynamodbMock{
		UpdateItemMock: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			capturedParams = params

			return &dynamodb.UpdateItemOutput{
				Attributes: map[string]types.AttributeValue{
					"PK":               &types.AttributeValueMemberS{Value: expectedUID},
					"message":          &types.AttributeValueMemberS{Value: expectedMessage},
					"last_accessed_at": &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
					"expires_at":       &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
				},
			}, nil
		},
	}

	// when
	s.repo = persistance.NewDbTable(dbMock, "mock-table")
	rec, err := s.repo.ReadRecord(context.Background(), expectedUID)

	// then
	s.Require().NoError(err)
	s.Require().NotNil(rec)
	s.Require().Equal(expectedUID, rec.PK)
	s.Require().Equal(expectedMessage, rec.Message)

	s.Require().NotNil(capturedParams)
	s.Require().NotNil(capturedParams.UpdateExpression)
}

func (s *repoSuite) TestReadRecord_ReturnsNotFoundWhenDeletedOrExpired() {
	// given
	otcSize := 6
	expectedUID, _ := generate.GenerateCode(otcSize)

	dbMock := &dynamodbMock{
		UpdateItemMock: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			return nil, &types.ConditionalCheckFailedException{}
		},
	}

	s.repo = persistance.NewDbTable(dbMock, "mock-table")

	// when
	rec, err := s.repo.ReadRecord(context.Background(), expectedUID)

	// then
	s.Require().ErrorIs(err, persistance.ErrRecordNotFound)
	s.Require().Nil(rec)
}

func (s *repoSuite) TestDeleteRecord_ReturnsNotFoundWhenRecordDoesNotExist() {
	// given
	uid := "missing-record"

	dbMock := &dynamodbMock{
		DeleteItemMock: func(
			ctx context.Context,
			params *dynamodb.DeleteItemInput,
			optFns ...func(*dynamodb.Options),
		) (*dynamodb.DeleteItemOutput, error) {
			return nil, &types.ConditionalCheckFailedException{}
		},
	}

	s.repo = persistance.NewDbTable(dbMock, "mock-table")

	// when
	err := s.repo.DeleteRecord(context.Background(), uid)

	// then
	s.Require().ErrorIs(err, persistance.ErrRecordNotFound)
}

func convertToIntFromStr(s string) (int64, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}
