package application

type Service struct {
	companies CompanyRepository
	needs     NeedRepository
	vacancies VacancyRepository
	search    SearchClient
	candidate CandidateClient
	txm       TransactionManager
}

func NewService(
	companies CompanyRepository,
	needs NeedRepository,
	vacancies VacancyRepository,
	search SearchClient,
	candidate CandidateClient,
	txm TransactionManager,
) *Service {
	return &Service{
		companies: companies,
		needs:     needs,
		vacancies: vacancies,
		search:    search,
		candidate: candidate,
		txm:       txm,
	}
}
