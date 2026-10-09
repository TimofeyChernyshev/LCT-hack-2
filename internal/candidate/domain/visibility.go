package domain

type Visibility struct {
	Contacts   bool
	Links      bool
	FSP        bool
	Experience bool
	Resume     bool
	Salary     bool
	SoftSkills bool
}

func DefaultVisibility() Visibility {
	return Visibility{
		Contacts:   true,
		Links:      true,
		FSP:        true,
		Experience: true,
		Resume:     true,
		Salary:     true,
		SoftSkills: true,
	}
}
