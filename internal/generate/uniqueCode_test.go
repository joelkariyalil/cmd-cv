package generate_test

import (
	"testing"

	"github.com/joelkariyalil/cmd-cv/internal/generate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCode_IsUnique(t *testing.T) {
	code1, err := generate.GenerateCode(5)
	require.NoError(t, err)

	code2, err := generate.GenerateCode(5)
	require.NoError(t, err)

	assert.NotEqual(t, code1, code2)
}

func TestGenerateCode_ReturnsErrOnInvalidInput(t *testing.T) {
	_, err := generate.GenerateCode(0)
	assert.Error(t, err)
}
