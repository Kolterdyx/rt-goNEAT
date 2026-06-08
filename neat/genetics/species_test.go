package genetics

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildSpeciesWithOrganisms(id int) (*Species, error) {
	gen := buildTestGenome(1)
	sp := NewSpecies(id)
	for i := 0; i < 3; i++ {
		org, err := NewOrganism(gen, id*(i+1))
		if err != nil {
			return nil, err
		}
		sp.addOrganism(org)
	}
	return sp, nil
}

func TestSpecies_Write(t *testing.T) {
	sp, err := buildSpeciesWithOrganisms(1)
	require.NoError(t, err, "failed to build species")

	outBuf := bytes.NewBufferString("")
	err = sp.Write(outBuf)
	require.NoError(t, err)

	out := outBuf.String()
	// Header line
	assert.Contains(t, out, fmt.Sprintf("Species #%d", sp.Id))
	assert.Contains(t, out, fmt.Sprintf("Size %d", len(sp.Organisms)))
	// Each organism header present
	for _, org := range sp.Organisms {
		assert.Contains(t, out, fmt.Sprintf("Organism #%d", org.Genotype.Id))
	}
}

func TestSpecies_Write_writeError(t *testing.T) {
	sp, err := buildSpeciesWithOrganisms(1)
	require.NoError(t, err, "failed to build species")

	errorWriter := ErrorWriter(1)
	err = sp.Write(&errorWriter)
	assert.EqualError(t, err, alwaysErrorText)
}

func TestSpecies_removeOrganism(t *testing.T) {
	sp, err := buildSpeciesWithOrganisms(1)
	require.NoError(t, err, "failed to build species")

	// test remove
	size := len(sp.Organisms)
	res, err := sp.removeOrganism(sp.Organisms[0])
	require.NoError(t, err, "failed to remove organism")
	require.True(t, res, "organism removal failed")
	require.Len(t, sp.Organisms, size-1, "wrong number of organisms after removal")

	// test fail to remove
	size = len(sp.Organisms)
	gen := buildTestGenome(2)
	org, err := NewOrganism(gen, 1)
	require.NoError(t, err, "failed to create organism")
	res, err = sp.removeOrganism(org)
	assert.False(t, res, "not existing organism can not be removed")
	assert.EqualError(t, err, fmt.Sprintf("attempt to remove nonexistent Organism from Species with #of organisms: %d", size))
	require.Len(t, sp.Organisms, size, "wrong number of organisms in species after unsuccessful removal attempt")
}

func TestSpecies_Size(t *testing.T) {
	sp, err := buildSpeciesWithOrganisms(1)
	require.NoError(t, err)
	assert.Equal(t, 3, sp.Size())
}

func TestSpecies_FindOldest(t *testing.T) {
	sp, err := buildSpeciesWithOrganisms(1)
	require.NoError(t, err)
	oldest := sp.FindOldest()
	require.NotNil(t, oldest)
	// organisms have ticks 1, 2, 3 (id=1, i=0..2 → tick = 1*(i+1))
	assert.Equal(t, 1, oldest.Generation)
}
