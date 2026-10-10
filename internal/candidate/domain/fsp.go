package domain

import "time"

// FSPAchievement - одно достижение участника ФСП.
// Все поля, кроме id/eventName/weight, nullable - данных из реестра ФСП
// может быть мало или не быть вовсе.
type FSPAchievement struct {
	ID         string
	UserID     string
	ExternalID *string
	EventName  string
	EventDate  *time.Time
	Place      *int
	Category   *string
	Score      *float32
	Weight     int
	Source     string
}

// FSPState - агрегат по пользователю: привязка к ФСП ID + список достижений.
// Используется в /me/fsp и при ранжировании.
type FSPState struct {
	FSPMemberID  *string
	LinkedAt     *time.Time
	Achievements []FSPAchievement
}

// HasFSP - есть ли у кандидата история ФСП (для публичной карточки).
func (s FSPState) HasFSP() bool {
	return s.FSPMemberID != nil || len(s.Achievements) > 0
}

// BestPlace - лучшее (наименьшее) место среди достижений.
// nil, если ни у одного достижения нет места.
func (s FSPState) BestPlace() *int {
	var best *int
	for _, a := range s.Achievements {
		if a.Place == nil {
			continue
		}
		if best == nil || *a.Place < *best {
			p := *a.Place
			best = &p
		}
	}
	return best
}

// TotalWeight - сумма весов - базовый сигнал для ранжирования.
func (s FSPState) TotalWeight() int {
	var sum int
	for _, a := range s.Achievements {
		sum += a.Weight
	}
	return sum
}
