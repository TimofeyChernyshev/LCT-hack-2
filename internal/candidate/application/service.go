package application

type Config struct {
	MaxUploadBytes int64
}

type Service struct {
	profiles     ProfileRepository
	contacts     ContactsRepository
	visibility   VisibilityRepository
	reveals      RevealRepository
	resumes      ResumeRepository
	experiences  ExperienceRepository
	technologies TechnologyRepository
	fsp          FSPRepository
	category     CategoryRepository
	files        FileStorage
	txm          TransactionManager
	cfg          Config
}

func NewService(
	profiles ProfileRepository,
	contacts ContactsRepository,
	visibility VisibilityRepository,
	reveals RevealRepository,
	resumes ResumeRepository,
	experiences ExperienceRepository,
	technologies TechnologyRepository,
	fsp FSPRepository,
	category CategoryRepository,
	files FileStorage,
	txm TransactionManager,
	cfg Config,
) *Service {
	return &Service{
		profiles: profiles, contacts: contacts,
		visibility: visibility, reveals: reveals,
		resumes: resumes, experiences: experiences,
		technologies: technologies,
		fsp:          fsp, category: category,
		files: files, txm: txm, cfg: cfg,
	}
}
