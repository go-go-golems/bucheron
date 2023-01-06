package pkg

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUUIDToHorseStaple(t *testing.T) {
	for i := 0; i < 10000; i++ {
		id := uuid.New()
		horseStaple := UUIDToHorseStaple(id)
		id2, err := HorseStapleToUUID(horseStaple)
		require.NoError(t, err)
		assert.Equal(t, id, id2)
	}
}
