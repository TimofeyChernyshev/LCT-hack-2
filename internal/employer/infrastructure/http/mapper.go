package http

import (
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/application"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/domain"
	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/infrastructure/http/api"
)

func uuidToAPI(s string) (openapi_types.UUID, error) { return uuid.Parse(s) }
func uuidPtr(s *string) *openapi_types.UUID {
	if s == nil {
		return nil
	}
	u, err := uuid.Parse(*s)
	if err != nil {
		return nil
	}
	return &u
}
func uuidSliceToAPI(in []string) []openapi_types.UUID {
	out := make([]openapi_types.UUID, 0, len(in))
	for _, s := range in {
		if u, err := uuid.Parse(s); err == nil {
			out = append(out, u)
		}
	}
	return out
}
func uuidSliceFromAPI(in []openapi_types.UUID) []string {
	out := make([]string, 0, len(in))
	for _, u := range in {
		out = append(out, u.String())
	}
	return out
}

func toCompanyResponse(c *domain.Company) api.Company {
	id, _ := uuidToAPI(c.ID)
	email := openapi_types.Email(*c.ContactEmail)

	return api.Company{
		Id:              id,
		Name:            c.Name,
		Description:     c.Description,
		Industry:        c.Industry,
		Website:         c.Website,
		Size:            c.Size,
		ContactPerson:   c.ContactPerson,
		ContactEmail:    &email,
		ContactPhone:    c.ContactPhone,
		ContactTelegram: c.ContactTelegram,
	}
}

func toNeedResponse(n *domain.Need) api.Need {
	id, _ := uuidToAPI(n.ID)
	stack := uuidSliceToAPI(n.Stack)
	r := api.Need{
		Id:                 id,
		Title:              n.Title,
		Description:        n.Description,
		CategoryId:         uuidPtr(n.CategoryID),
		SpecializationId:   uuidPtr(n.SpecializationID),
		GradeId:            uuidPtr(n.GradeID),
		Stack:              &stack,
		SalaryMin:          n.SalaryMin,
		SalaryMax:          n.SalaryMax,
		Currency:           n.Currency,
		Location:           n.Location,
		MinExperienceYears: n.MinExperienceYears,
		Status:             api.NeedStatus(n.Status),
	}
	if n.WorkFormat != nil {
		wf := api.NeedWorkFormat(*n.WorkFormat)
		r.WorkFormat = &wf
	}
	return r
}

func toNeedsResponse(list []domain.Need) []api.Need {
	out := make([]api.Need, 0, len(list))
	for i := range list {
		out = append(out, toNeedResponse(&list[i]))
	}
	return out
}

func toVacancyResponse(v *domain.Vacancy) api.Vacancy {
	id, _ := uuidToAPI(v.ID)
	cid, _ := uuidToAPI(v.CompanyID)
	stack := uuidSliceToAPI(v.Stack)
	status := api.VacancyStatus(v.Status)
	out := api.Vacancy{
		Id:               &id,
		CompanyId:        &cid,
		Title:            v.Title,
		Description:      v.Description,
		CategoryId:       uuidPtr(v.CategoryID),
		SpecializationId: uuidPtr(v.SpecializationID),
		GradeId:          uuidPtr(v.GradeID),
		Stack:            &stack,
		SalaryMin:        v.SalaryMin,
		SalaryMax:        v.SalaryMax,
		Location:         v.Location,
		Status:           &status,
		PublishedAt:      v.PublishedAt,
		CreatedAt:        &v.CreatedAt,
	}
	if v.WorkFormat != nil {
		wf := api.VacancyWorkFormat(*v.WorkFormat)
		out.WorkFormat = &wf
	}
	return out
}

func toVacanciesResponse(list []domain.Vacancy) []api.Vacancy {
	out := make([]api.Vacancy, 0, len(list))
	for i := range list {
		out = append(out, toVacancyResponse(&list[i]))
	}
	return out
}

func toSearchPageResponse(p *application.SearchPage) api.CandidateSearchPage {
	out := api.CandidateSearchPage{
		Total:  p.Total,
		Limit:  p.Limit,
		Offset: p.Offset,
	}
	for _, h := range p.Items {
		uid, _ := uuidToAPI(h.UserID)
		cid, _ := uuidToAPI(h.CategoryID)
		gid, _ := uuidToAPI(h.GradeID)
		sid, _ := uuidToAPI(h.SpecializationID)
		out.Items = append(out.Items, api.CandidateSearchHit{
			UserId:               uid,
			DisplayName:          h.DisplayName,
			CategoryId:           cid,
			GradeId:              gid,
			SpecializationId:     sid,
			TestScore:            h.TestScore,
			FspAchievementsCount: &h.FSPAchievementsCount,
			FspBestPlace:         h.FSPBestPlace,
			FspWeightSum:         &h.FSPWeightSum,
			YearsExperience:      h.YearsExperience,
			Score:                h.Score,
			Reasons:              h.Reasons,
		})
	}
	return out
}

func toCandidateCardResponse(card *application.CandidateCard) api.CandidateCard {
	uid, _ := uuidToAPI(card.UserID)
	out := api.CandidateCard{
		UserId:           &uid,
		DisplayName:      &card.DisplayName,
		Headline:         card.Headline,
		CategoryId:       uuidPtr(card.CategoryID),
		GradeId:          uuidPtr(card.GradeID),
		SpecializationId: uuidPtr(card.SpecializationID),
		YearsExperience:  card.YearsExperience,
		Location:         card.Location,
		SoftSkills:       &card.SoftSkills,
	}
	if card.Salary != nil {
		out.Salary = &api.CandidateCardSalary{
			Min:      card.Salary.Min,
			Max:      card.Salary.Max,
			Currency: card.Salary.Currency,
			Masked:   boolPtr(card.Salary.Masked),
		}
	}
	if card.FSP != nil {
		out.Fsp = &api.CandidateCardFsp{
			HasFSP:            &card.FSP.HasFSP,
			AchievementsCount: &card.FSP.AchievementsCount,
			BestPlace:         card.FSP.BestPlace,
		}
	}
	maskedFields := []api.CandidateContactsMaskedFields{}
	for _, v := range card.Contacts.MaskedFields {
		maskedFields = append(maskedFields, api.CandidateContactsMaskedFields(v))
	}
	if card.Contacts != nil {
		out.Contacts = &api.CandidateContacts{
			Email:        emailPtr(card.Contacts.Email),
			Phone:        card.Contacts.Phone,
			Telegram:     card.Contacts.Telegram,
			Github:       card.Contacts.GitHub,
			Linkedin:     card.Contacts.LinkedIn,
			Website:      card.Contacts.Website,
			Masked:       boolPtr(card.Contacts.Masked),
			MaskedFields: &maskedFields,
		}
	}
	return out
}

func boolPtr(b bool) *bool { return &b }
func emailPtr(s *string) *openapi_types.Email {
	if s == nil || *s == "" {
		return nil
	}
	e := openapi_types.Email(*s)
	return &e
}
