package fsp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrMemberNotFound = errors.New("fsp member not found in registry")
)

// Registry interface defines methods for accessing the FSP registry
type Registry interface {
	GetMember(ctx context.Context, fspID string) (*Member, error)
	SearchMembers(ctx context.Context, query string, rank SportsRank, region string, limit, offset int) ([]Member, int, error)
	VerifyMember(ctx context.Context, req VerificationRequest) (*VerificationResult, error)
	RegisterMember(ctx context.Context, member Member) error
}

// MockRegistry is an in-memory implementation of the FSP Registry
type MockRegistry struct {
	mu      sync.RWMutex
	members map[string]*Member // keyed by normalized fspID
	aliases map[string]string  // alias -> canonical fspID
}

// NewMockRegistry creates and pre-seeds an in-memory FSP registry
func NewMockRegistry() *MockRegistry {
	r := &MockRegistry{
		members: make(map[string]*Member),
		aliases: make(map[string]string),
	}
	r.seedDefaultMembers()
	return r
}

func (r *MockRegistry) GetMember(_ context.Context, fspID string) (*Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normID := strings.ToUpper(strings.TrimSpace(fspID))
	if canonical, ok := r.aliases[normID]; ok {
		normID = canonical
	}

	m, ok := r.members[normID]
	if !ok {
		return nil, ErrMemberNotFound
	}

	// Return a copy to prevent accidental outside mutation
	copied := *m
	copied.Achievements = make([]Achievement, len(m.Achievements))
	copy(copied.Achievements, m.Achievements)
	return &copied, nil
}

func (r *MockRegistry) SearchMembers(_ context.Context, query string, rank SportsRank, region string, limit, offset int) ([]Member, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(query))
	reg := strings.ToLower(strings.TrimSpace(region))

	var matched []Member
	for _, m := range r.members {
		if q != "" {
			if !strings.Contains(strings.ToLower(m.FullName), q) &&
				!strings.Contains(strings.ToLower(m.FSPID), q) {
				continue
			}
		}
		if rank != "" && m.SportsRank != rank {
			continue
		}
		if reg != "" && !strings.Contains(strings.ToLower(m.Region), reg) {
			continue
		}
		matched = append(matched, *m)
	}

	total := len(matched)
	if offset >= total {
		return []Member{}, total, nil
	}
	end := offset + limit
	if limit <= 0 || end > total {
		end = total
	}

	return matched[offset:end], total, nil
}

func (r *MockRegistry) VerifyMember(ctx context.Context, req VerificationRequest) (*VerificationResult, error) {
	member, err := r.GetMember(ctx, req.FSPID)
	if err != nil {
		return &VerificationResult{
			IsValid:    false,
			Message:    fmt.Sprintf("Участник с ID %s не найден в реестре ФСП", req.FSPID),
			VerifiedAt: time.Now(),
		}, nil
	}

	if req.FullName != "" {
		if !strings.EqualFold(strings.TrimSpace(req.FullName), strings.TrimSpace(member.FullName)) {
			return &VerificationResult{
				IsValid:    false,
				Message:    fmt.Sprintf("ФИО '%s' не совпадает с записью в реестре ФСП ('%s')", req.FullName, member.FullName),
				VerifiedAt: time.Now(),
			}, nil
		}
	}

	return &VerificationResult{
		IsValid:    true,
		Member:     member,
		Message:    "Участник успешно верифицирован в реестре Федерации спортивного программирования",
		VerifiedAt: time.Now(),
	}, nil
}

func (r *MockRegistry) RegisterMember(_ context.Context, member Member) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	normID := strings.ToUpper(strings.TrimSpace(member.FSPID))
	if normID == "" {
		return errors.New("fspId is required")
	}

	r.members[normID] = &member
	return nil
}

func (r *MockRegistry) AddAlias(alias, canonicalID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.aliases[strings.ToUpper(strings.TrimSpace(alias))] = strings.ToUpper(strings.TrimSpace(canonicalID))
}

func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}

func (r *MockRegistry) seedDefaultMembers() {
	defaults := []struct {
		member  Member
		aliases []string
	}{
		{
			member: Member{
				FSPID:      "FSP-RU-77-00101",
				FullName:   "Смирнов Александр Дмитриевич",
				SportsRank: RankMS,
				Rating:     2540,
				Region:     "г. Москва",
				Discipline: DisciplineAlgorithms,
				Status:     "active",
				Verified:   true,
				Achievements: []Achievement{
					{
						ID:          "ach-001",
						ExternalID:  "FSP-ACH-2024-01",
						EventName:   "Чемпионат России по спортивному программированию 2024",
						EventDate:   "2024-05-18",
						Place:       intPtr(1),
						Category:    CategoryChampionship,
						Score:       floatPtr(490.5),
						Weight:      10,
						Badge:       "gold",
						Description: "1 место в дисциплине 'Алгоритмическое программирование'",
					},
					{
						ID:          "ach-002",
						ExternalID:  "FSP-ACH-2024-02",
						EventName:   "Кубок России по спортивному программированию 2024",
						EventDate:   "2024-09-12",
						Place:       intPtr(2),
						Category:    CategoryCup,
						Score:       floatPtr(465.0),
						Weight:      8,
						Badge:       "silver",
						Description: "2 место, серебряная медаль финала Кубка",
					},
					{
						ID:          "ach-003",
						ExternalID:  "FSP-ACH-2024-03",
						EventName:   "Всероссийский хакатон ФСП по спортивному программированию",
						EventDate:   "2024-11-05",
						Place:       intPtr(1),
						Category:    CategoryHackathon,
						Score:       floatPtr(98.5),
						Weight:      7,
						Badge:       "winner",
						Description: "Победитель трека 'Высоконагруженные алгоритмы'",
					},
				},
			},
			aliases: []string{"FSP-10001", "FSP-101"},
		},
		{
			member: Member{
				FSPID:      "FSP-RU-78-00202",
				FullName:   "Иванова Мария Сергеевна",
				SportsRank: RankKMS,
				Rating:     2190,
				Region:     "г. Санкт-Петербург",
				Discipline: DisciplineProducts,
				Status:     "active",
				Verified:   true,
				Achievements: []Achievement{
					{
						ID:          "ach-004",
						ExternalID:  "FSP-ACH-2024-04",
						EventName:   "Чемпионат Санкт-Петербурга по спортивному программированию 2024",
						EventDate:   "2024-04-20",
						Place:       intPtr(1),
						Category:    CategoryRegional,
						Score:       floatPtr(410.0),
						Weight:      7,
						Badge:       "gold",
						Description: "1 место в региональном первенстве СЗФО",
					},
					{
						ID:          "ach-005",
						ExternalID:  "FSP-ACH-2024-05",
						EventName:   "Всероссийский хакатон ФСП 'Код Мира'",
						EventDate:   "2024-08-14",
						Place:       intPtr(1),
						Category:    CategoryHackathon,
						Score:       floatPtr(96.0),
						Weight:      8,
						Badge:       "winner",
						Description: "Победитель трека 'Продуктовая разработка и распределенные системы'",
					},
					{
						ID:          "ach-006",
						ExternalID:  "FSP-ACH-2023-06",
						EventName:   "Кубок ФСП 2023",
						EventDate:   "2023-10-10",
						Place:       intPtr(3),
						Category:    CategoryCup,
						Score:       floatPtr(380.0),
						Weight:      6,
						Badge:       "bronze",
						Description: "3 место в командном зачете",
					},
				},
			},
			aliases: []string{"FSP-10002", "FSP-202"},
		},
		{
			member: Member{
				FSPID:      "FSP-RU-16-00303",
				FullName:   "Галиев Руслан Рамилевич",
				SportsRank: Rank1,
				Rating:     1940,
				Region:     "Республика Татарстан",
				Discipline: DisciplineAI,
				Status:     "active",
				Verified:   true,
				Achievements: []Achievement{
					{
						ID:          "ach-007",
						ExternalID:  "FSP-ACH-2024-07",
						EventName:   "Открытый кубок Казани по ИИ 2024",
						EventDate:   "2024-06-01",
						Place:       intPtr(2),
						Category:    CategoryRegional,
						Score:       floatPtr(375.0),
						Weight:      6,
						Badge:       "silver",
						Description: "2 место в треке машинного обучения",
					},
					{
						ID:          "ach-008",
						ExternalID:  "FSP-ACH-2024-08",
						EventName:   "Хакатон ФСП по алгоритмам искусственного интеллекта",
						EventDate:   "2024-10-22",
						Place:       intPtr(1),
						Category:    CategoryHackathon,
						Score:       floatPtr(94.0),
						Weight:      7,
						Badge:       "winner",
						Description: "1 место, кейс от индустриального партнера",
					},
				},
			},
			aliases: []string{"FSP-10003", "FSP-303"},
		},
		{
			member: Member{
				FSPID:      "FSP-RU-54-00404",
				FullName:   "Ковалев Андрей Павлович",
				SportsRank: Rank2,
				Rating:     1720,
				Region:     "Новосибирская область",
				Discipline: DisciplineAlgorithms,
				Status:     "active",
				Verified:   true,
				Achievements: []Achievement{
					{
						ID:          "ach-009",
						ExternalID:  "FSP-ACH-2024-09",
						EventName:   "Региональный турнир Сибири по спортивному программированию 2024",
						EventDate:   "2024-03-15",
						Place:       intPtr(3),
						Category:    CategoryRegional,
						Score:       floatPtr(310.0),
						Weight:      5,
						Badge:       "bronze",
						Description: "3 место среди юниоров Сибирского федерального округа",
					},
				},
			},
			aliases: []string{"FSP-10004", "FSP-404"},
		},
		{
			member: Member{
				FSPID:      "FSP-RU-66-00505",
				FullName:   "Морозова Дарья Викторовна",
				SportsRank: Rank3,
				Rating:     1530,
				Region:     "Свердловская область",
				Discipline: DisciplineProducts,
				Status:     "active",
				Verified:   true,
				Achievements: []Achievement{
					{
						ID:          "ach-010",
						ExternalID:  "FSP-ACH-2024-10",
						EventName:   "Кубок Екатеринбурга по веб-разработке",
						EventDate:   "2024-02-11",
						Place:       intPtr(5),
						Category:    CategoryRegional,
						Score:       floatPtr(260.0),
						Weight:      4,
						Badge:       "finalist",
						Description: "Финалист турнира",
					},
				},
			},
			aliases: []string{"FSP-10005", "FSP-505"},
		},
		{
			member: Member{
				FSPID:      "FSP-RU-77-00606",
				FullName:   "Соколов Дмитрий Евгеньевич",
				SportsRank: RankMS,
				Rating:     2620,
				Region:     "г. Москва",
				Discipline: DisciplineAlgorithms,
				Status:     "honorary",
				Verified:   true,
				Achievements: []Achievement{
					{
						ID:          "ach-011",
						ExternalID:  "FSP-ACH-2023-11",
						EventName:   "Финал Чемпионата России 2023",
						EventDate:   "2023-05-20",
						Place:       intPtr(1),
						Category:    CategoryChampionship,
						Score:       floatPtr(500.0),
						Weight:      10,
						Badge:       "gold",
						Description: "Абсолютный чемпион России 2023",
					},
					{
						ID:          "ach-012",
						ExternalID:  "FSP-ACH-2024-12",
						EventName:   "Суперкубок ФСП 2024",
						EventDate:   "2024-12-01",
						Place:       intPtr(1),
						Category:    CategoryCup,
						Score:       floatPtr(495.0),
						Weight:      9,
						Badge:       "gold",
						Description: "Победитель Суперкубка",
					},
				},
			},
			aliases: []string{"FSP-10006", "FSP-606"},
		},
	}

	for _, d := range defaults {
		m := d.member
		r.members[strings.ToUpper(m.FSPID)] = &m
		for _, al := range d.aliases {
			r.aliases[strings.ToUpper(al)] = strings.ToUpper(m.FSPID)
		}
	}
}
