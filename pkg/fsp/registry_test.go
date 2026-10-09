package fsp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockRegistry_GetMember(t *testing.T) {
	reg := NewMockRegistry()
	ctx := context.Background()

	// 1. Direct canonical lookup
	m, err := reg.GetMember(ctx, "FSP-RU-77-00101")
	require.NoError(t, err)
	assert.Equal(t, "Смирнов Александр Дмитриевич", m.FullName)
	assert.Equal(t, RankMS, m.SportsRank)
	assert.NotEmpty(t, m.Achievements)

	// 2. Case insensitive lookup
	m2, err := reg.GetMember(ctx, "fsp-ru-77-00101")
	require.NoError(t, err)
	assert.Equal(t, m.FullName, m2.FullName)

	// 3. Alias lookup
	m3, err := reg.GetMember(ctx, "FSP-10001")
	require.NoError(t, err)
	assert.Equal(t, m.FullName, m3.FullName)

	// 4. Non-existent ID
	_, err = reg.GetMember(ctx, "NON-EXISTENT-ID")
	assert.ErrorIs(t, err, ErrMemberNotFound)
}

func TestMockRegistry_SearchMembers(t *testing.T) {
	reg := NewMockRegistry()
	ctx := context.Background()

	// Search by name
	members, total, err := reg.SearchMembers(ctx, "Иванова", "", "", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.Equal(t, "Иванова Мария Сергеевна", members[0].FullName)

	// Search by rank
	msMembers, totalMS, err := reg.SearchMembers(ctx, "", RankMS, "", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, totalMS, 2)
	for _, m := range msMembers {
		assert.Equal(t, RankMS, m.SportsRank)
	}

	// Search by region
	tatMembers, _, err := reg.SearchMembers(ctx, "", "", "Татарстан", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tatMembers), 1)
}

func TestMockRegistry_VerifyMember(t *testing.T) {
	reg := NewMockRegistry()
	ctx := context.Background()

	// Success verify
	res, err := reg.VerifyMember(ctx, VerificationRequest{
		FSPID:    "FSP-RU-77-00101",
		FullName: "Смирнов Александр Дмитриевич",
	})
	require.NoError(t, err)
	assert.True(t, res.IsValid)
	assert.NotNil(t, res.Member)

	// Fail with mismatched name
	res2, err := reg.VerifyMember(ctx, VerificationRequest{
		FSPID:    "FSP-RU-77-00101",
		FullName: "Петров Иван",
	})
	require.NoError(t, err)
	assert.False(t, res2.IsValid)

	// Fail with non-existent ID
	res3, err := reg.VerifyMember(ctx, VerificationRequest{
		FSPID: "UNKNOWN-999",
	})
	require.NoError(t, err)
	assert.False(t, res3.IsValid)
}
