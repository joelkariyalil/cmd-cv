package generate

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

const base = 10

func GenerateCode(size int) (string, error) {
	if size <= 0 {
		return "", errors.New("code size must be greater than 0")
	}

	maxVal := new(big.Int).Exp(
		big.NewInt(base),
		big.NewInt(int64(size)),
		nil,
	)

	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", size, n), nil
}
