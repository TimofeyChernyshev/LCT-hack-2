package http

import (
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/application"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/infrastructure/http/api"
)

func uuidToAPI(s string) (openapi_types.UUID, error) {
	return uuid.Parse(s)
}

func toCardResponse(v *application.CardView) api.CandidateCard {
	id, _ := uuidToAPI(v.Profile.UserID)
	displayName := v.Profile.DisplayName()

	card := api.CandidateCard{
		UserId:          &id,
		DisplayName:     &displayName,
		Headline:        v.Profile.Headline,
		YearsExperience: v.Profile.YearsExperience,
		Location:        v.Profile.Location,
		SoftSkills:      &v.Profile.SoftSkills,
	}

	if v.Profile.CategoryID != nil {
		u, _ := uuidToAPI(*v.Profile.CategoryID)
		card.CategoryId = &u
	}
	if v.Profile.GradeID != nil {
		u, _ := uuidToAPI(*v.Profile.GradeID)
		card.GradeId = &u
	}
	if v.Profile.SpecializationID != nil {
		u, _ := uuidToAPI(*v.Profile.SpecializationID)
		card.SpecializationId = &u
	}

	card.Fsp = &api.CandidateCardFsp{
		HasFSP:            &v.FSP.HasFSP,
		AchievementsCount: &v.FSP.AchievementsCount,
		BestPlace:         v.FSP.BestPlace,
	}
	card.Salary = &api.CandidateCardSalary{
		Min:      v.Salary.Min,
		Max:      v.Salary.Max,
		Currency: &v.Salary.Currency,
		Masked:   boolPtr(v.Salary.Masked),
	}
	card.Contacts = &api.CandidateContacts{
		Email:        emailPtr(v.Contacts.Email),
		Phone:        v.Contacts.Phone,
		Telegram:     v.Contacts.Telegram,
		Github:       v.Contacts.GitHub,
		Linkedin:     v.Contacts.LinkedIn,
		Website:      v.Contacts.Website,
		Masked:       boolPtr(v.Contacts.Masked),
		MaskedFields: &v.Contacts.MaskedFields,
	}
	return card
}

func toProfileResponse(p *domain.Profile) api.CandidateProfile {
	id, _ := uuidToAPI(p.UserID)
	out := api.CandidateProfile{
		UserId:          id,
		FirstName:       p.FirstName,
		LastName:        p.LastName,
		MiddleName:      p.MiddleName,
		Headline:        p.Headline,
		About:           p.About,
		Location:        p.Location,
		YearsExperience: p.YearsExperience,
		SalaryMin:       p.SalaryMin,
		SalaryMax:       p.SalaryMax,
		SalaryCurrency:  p.SalaryCurrency,
		SoftSkills:      &p.SoftSkills,
		UpdatedAt:       p.UpdatedAt,
	}
	if p.CategoryID != nil {
		u, _ := uuidToAPI(*p.CategoryID)
		out.CategoryId = &u
	}
	if p.GradeID != nil {
		u, _ := uuidToAPI(*p.GradeID)
		out.GradeId = &u
	}
	if p.SpecializationID != nil {
		u, _ := uuidToAPI(*p.SpecializationID)
		out.SpecializationId = &u
	}
	if p.FSPMemberID != nil {
		out.FspMemberId = p.FSPMemberID
	}
	return out
}

func toContactsResponse(c *domain.Contacts) api.CandidateContacts {
	return api.CandidateContacts{
		Email:    emailPtr(c.Email),
		Phone:    c.Phone,
		Telegram: c.Telegram,
		Github:   c.GitHub,
		Linkedin: c.LinkedIn,
		Website:  c.Website,
	}
}

func toVisibilityResponse(v *domain.Visibility) api.Visibility {
	return api.Visibility{
		Contacts:   boolPtr(v.Contacts),
		Links:      boolPtr(v.Links),
		Fsp:        boolPtr(v.FSP),
		Experience: boolPtr(v.Experience),
		Resume:     boolPtr(v.Resume),
		Salary:     boolPtr(v.Salary),
		SoftSkills: boolPtr(v.SoftSkills),
	}
}

func toResumeResponse(r *domain.Resume) api.Resume {
	id, _ := uuidToAPI(r.ID)
	source := api.ResumeSource(r.Source)

	out := api.Resume{
		Id:        &id,
		Title:     &r.Title,
		Source:    &source,
		IsPrimary: boolPtr(r.IsPrimary),
		CreatedAt: &r.CreatedAt,
	}

	if r.Parsed != nil {
		out.Parsed = &r.Parsed
	}
	return out
}

func toResumesResponse(list []domain.Resume) []api.Resume {
	out := make([]api.Resume, 0, len(list))
	for i := range list {
		out = append(out, toResumeResponse(&list[i]))
	}
	return out
}

func toExperienceResponse(e *domain.Experience) api.Experience {
	id, _ := uuidToAPI(e.ID)

	return api.Experience{
		Id:          id,
		Company:     e.Company,
		Position:    e.Position,
		StartedAt:   openapi_types.Date{Time: e.StartedAt},
		EndedAt:     &openapi_types.Date{Time: e.EndedAt},
		Description: e.Description,
	}
}

func toExperiencesResponse(list []domain.Experience) []api.Experience {
	out := make([]api.Experience, 0, len(list))
	for i := range list {
		out = append(out, toExperienceResponse(&list[i]))
	}
	return out
}

func toFSPStateResponse(s *domain.FSPState) api.FSPState {
	out := api.FSPState{
		FspMemberId: s.FSPMemberID,
		LinkedAt:    s.LinkedAt,
	}
	for _, a := range s.Achievements {
		id, _ := uuidToAPI(a.ID)
		*out.Achievements = append(*out.Achievements, api.FSPAchievement{
			Id:         &id,
			ExternalId: a.ExternalID,
			EventName:  &a.EventName,
			EventDate:  &openapi_types.Date{Time: *a.EventDate},
			Place:      a.Place,
			Category:   a.Category,
			Score:      a.Score,
			Weight:     intPtr(a.Weight),
		})
	}
	return out
}

func toCategoryStateResponse(s *domain.CategoryState) api.CategoryState {
	category, _ := uuidToAPI(*s.CategoryID)
	grade, _ := uuidToAPI(*s.GradeID)
	specialization, _ := uuidToAPI(*s.SpecializationID)

	out := api.CategoryState{
		CategoryId:       &category,
		GradeId:          &grade,
		SpecializationId: &specialization,
	}
	for _, h := range s.History {
		cid, _ := uuidToAPI(h.CategoryID)
		gid, _ := uuidToAPI(h.GradeID)
		reason := api.CategoryStateHistoryReason(h.Reason)
		*out.History = append(*out.History, struct {
			CategoryId    *openapi_types.UUID             "json:\"categoryId,omitempty\""
			EffectiveFrom *time.Time                      "json:\"effectiveFrom,omitempty\""
			EffectiveTo   *time.Time                      "json:\"effectiveTo,omitempty\""
			GradeId       *openapi_types.UUID             "json:\"gradeId,omitempty\""
			Reason        *api.CategoryStateHistoryReason "json:\"reason,omitempty\""
		}{
			CategoryId:    &cid,
			GradeId:       &gid,
			Reason:        &reason,
			EffectiveFrom: &h.EffectiveFrom,
			EffectiveTo:   h.EffectiveTo,
		})
	}
	return out
}

func intPtr(i int) *int { return &i }

func emailPtr(s string) *openapi_types.Email {
	if s == "" {
		return nil
	}
	e := openapi_types.Email(s)
	return &e
}

func boolPtr(b bool) *bool { return &b }
