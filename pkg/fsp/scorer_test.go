package fsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateScore_NoFSP(t *testing.T) {
	// Case 1: nil member
	res1 := CalculateScore(nil)
	assert.Equal(t, 0.0, res1.Score)
	assert.Equal(t, 0, res1.WeightSum)
	assert.Equal(t, 0, res1.AchievementsCount)
	assert.Nil(t, res1.BestPlace)
	assert.Equal(t, NoFSPExplanation, res1.Explanation)

	// Case 2: empty FSP ID
	res2 := CalculateScore(&Member{FSPID: ""})
	assert.Equal(t, 0.0, res2.Score)
	assert.Equal(t, NoFSPExplanation, res2.Explanation)
}

func TestCalculateScore_MasterOfSportsPodium(t *testing.T) {
	place1 := 1
	place2 := 2
	member := &Member{
		FSPID:      "FSP-RU-77-00101",
		FullName:   "Смирнов Александр Дмитриевич",
		SportsRank: RankMS,
		Rating:     2540,
		Region:     "г. Москва",
		Discipline: DisciplineAlgorithms,
		Achievements: []Achievement{
			{
				EventName: "Чемпионат России по спортивному программированию 2024",
				Place:     &place1,
				Category:  CategoryChampionship,
				Weight:    10,
			},
			{
				EventName: "Кубок России 2024",
				Place:     &place2,
				Category:  CategoryCup,
				Weight:    8,
			},
		},
	}

	res := CalculateScore(member)
	require.NotNil(t, res.BestPlace)
	assert.Equal(t, 1, *res.BestPlace)
	assert.Equal(t, 18, res.WeightSum)
	assert.Equal(t, 2, res.AchievementsCount)
	assert.GreaterOrEqual(t, res.Score, 80.0)
	assert.Contains(t, res.Explanation, "Мастер спорта")
	assert.Contains(t, res.Explanation, "2540")
	assert.Contains(t, res.Explanation, "Чемпионат России")
}

func TestCalculateScore_UnrankedBeginner(t *testing.T) {
	member := &Member{
		FSPID:      "FSP-TEST-01",
		FullName:   "Новичок Иван",
		SportsRank: RankUnranked,
		Rating:     1250,
		Region:     "г. Воронеж",
	}

	res := CalculateScore(member)
	assert.Nil(t, res.BestPlace)
	assert.Equal(t, 0, res.AchievementsCount)
	assert.Equal(t, 0, res.WeightSum)
	assert.Greater(t, res.Score, 0.0)
	assert.Less(t, res.Score, 20.0)
	assert.Contains(t, res.Explanation, "Без разряда")
}
