package genetics

import (
	"bytes"
	"encoding/gob"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrganisms_Sort verifies that Organisms sorts ascending by birth tick (Generation).
func TestOrganisms_Sort(t *testing.T) {
	gnome := buildTestGenome(1)
	count := 100
	orgs := make(Organisms, count)
	var err error
	for i := 0; i < count; i++ {
		orgs[i], err = NewOrganism(gnome, rand.Intn(1000))
		require.NoError(t, err, "failed to create organism: %d", i)
	}

	// sort ascending
	sort.Sort(orgs)
	tick := -1
	for _, o := range orgs {
		assert.True(t, o.Generation >= tick, "wrong ascending sort order")
		tick = o.Generation
	}

	// sort descending
	for i := 0; i < count; i++ {
		orgs[i], err = NewOrganism(gnome, rand.Intn(1000))
		require.NoError(t, err, "failed to create organism: %d", i)
	}
	sort.Sort(sort.Reverse(orgs))
	tick = 1<<31 - 1
	for _, o := range orgs {
		assert.True(t, o.Generation <= tick, "wrong descending sort order")
		tick = o.Generation
	}
}

func TestOrganism_Phenotype(t *testing.T) {
	gnome := buildTestGenome(1)
	organism, err := NewOrganism(gnome, 1)
	require.NoError(t, err)

	phenotype, err := organism.Phenotype()
	require.NoError(t, err)
	require.NotNil(t, phenotype)

	assert.Equal(t, 4, phenotype.NodeCount(), "wrong nodes count")
	assert.Equal(t, 3, phenotype.LinkCount(), "wrong links count")

	// check that phenotype not created twice
	other, err := organism.Phenotype()
	require.NoError(t, err)
	assert.True(t, phenotype == other, "must be the same pointer")
}

func TestOrganism_MarshalBinary(t *testing.T) {
	gnome := buildTestGenome(1)
	org, err := NewOrganism(gnome, 42)
	require.NoError(t, err, "failed to create organism")

	// Marshal to binary
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(org)
	require.NoError(t, err, "failed to encode")

	// Unmarshal and check if the same
	dec := gob.NewDecoder(&buf)
	decOrg := Organism{}
	err = dec.Decode(&decOrg)
	require.NoError(t, err, "failed to decode")

	assert.Equal(t, org.Generation, decOrg.Generation)

	decGnome := decOrg.Genotype
	assert.Equal(t, gnome.Id, decGnome.Id)

	equals, err := gnome.IsEqual(decGnome)
	require.NoError(t, err, "failed to check equality")
	assert.True(t, equals)
}

func TestOrganism_IsAlive(t *testing.T) {
	gnome := buildTestGenome(1)
	org, err := NewOrganism(gnome, 1)
	require.NoError(t, err)
	assert.True(t, org.IsAlive())

	org.alive = false
	assert.False(t, org.IsAlive())
}

func TestOrganism_UpdatePhenotype(t *testing.T) {
	gnome := buildTestGenome(1)
	org, err := NewOrganism(gnome, 1)
	require.NoError(t, err, "failed to create organism")

	org.orgPhenotype = nil
	assert.Nil(t, org.orgPhenotype, "no phenotype expected")

	err = org.UpdatePhenotype()
	require.NoError(t, err, "failed to recreate phenotype")
	assert.NotNil(t, org.orgPhenotype)
}
