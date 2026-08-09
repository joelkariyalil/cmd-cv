package persistance

import "fmt"

type DbKey struct {
	PK string `dynamodbav:"PK"`
}

type NotFoundError struct {
	DbKey
	TableName string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("could not find dbentry for key: %+v, in table: %s", e.DbKey, e.TableName)
}
