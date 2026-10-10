package resume

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	emailRegex    = regexp.MustCompile(`(?i)[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	phoneRegex    = regexp.MustCompile(`(?:\+7|8)[\s\-]?(?:\([0-9]{3}\)|[0-9]{3})[\s\-]?[0-9]{3}[\s\-]?[0-9]{2}[\s\-]?[0-9]{2}`)
	telegramRegex = regexp.MustCompile(`(?i)(?:t\.me/|@)([a-zA-Z0-9_]{4,})`)
	githubRegex   = regexp.MustCompile(`(?i)github\.com/([a-zA-Z0-9_-]+)`)
	linkedinRegex = regexp.MustCompile(`(?i)linkedin\.com/in/([a-zA-Z0-9_-]+)`)
	fspIDRegex    = regexp.MustCompile(`(?i)FSP-[A-Z0-9]{2,4}-[0-9]{2,4}-[0-9]+`)

	// Experience years mention regex e.g. "Опыт работы: 5 лет", "Стаж: 3.5 года", "6+ years"
	expYearsRegex = regexp.MustCompile(`(?i)(?:опыт(?:\s+работы)?|стаж|experience)[:\s]+(\d+(?:[.,]\d+)?)\s*(?:лет|года|год|years|yrs)`)

	// Year range regex e.g. "2020 — 2023", "2019 - настоящее время"
	periodRegex = regexp.MustCompile(`(?i)(\d{4})\s*(?:—|-|по)\s*(\d{4}|настоящее\s+время|н\.в\.|present)`)

	knownCities = []string{
		"Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург", "Казань",
		"Нижний Новгород", "Самара", "Томск", "Уфа", "Пермь", "Ростов-на-Дону",
		"Воронеж", "Красноярск", "Краснодар", "Минск", "Алматы",
	}

	skillDictionary = map[string]string{
		"go":             "Go",
		"golang":         "Go",
		"python":         "Python",
		"fastapi":        "FastAPI",
		"django":         "Django",
		"flask":          "Flask",
		"java":           "Java",
		"kotlin":         "Kotlin",
		"swift":          "Swift",
		"typescript":     "TypeScript",
		"ts":             "TypeScript",
		"javascript":     "JavaScript",
		"js":             "JavaScript",
		"react":          "React",
		"react.js":       "React",
		"vue":            "Vue",
		"vue.js":         "Vue",
		"next.js":        "Next.js",
		"nextjs":         "Next.js",
		"c++":            "C++",
		"c#":             "C#",
		".net":           ".NET",
		"rust":           "Rust",
		"postgresql":     "PostgreSQL",
		"postgres":       "PostgreSQL",
		"mysql":          "MySQL",
		"redis":          "Redis",
		"mongodb":        "MongoDB",
		"clickhouse":     "ClickHouse",
		"elasticsearch":  "Elasticsearch",
		"docker":         "Docker",
		"kubernetes":     "Kubernetes",
		"k8s":            "Kubernetes",
		"kafka":          "Kafka",
		"rabbitmq":       "RabbitMQ",
		"nginx":          "Nginx",
		"linux":          "Linux",
		"git":            "Git",
		"ci/cd":          "CI/CD",
		"grpc":           "gRPC",
		"graphql":        "GraphQL",
		"rest api":       "REST API",
		"rest":           "REST API",
		"pytorch":        "PyTorch",
		"tensorflow":     "TensorFlow",
		"pandas":         "Pandas",
		"numpy":          "NumPy",
		"spring boot":    "Spring Boot",
		"spring":         "Spring",
		"microservices":  "Микросервисы",
		"микросервисы":   "Микросервисы",
		"sql":            "SQL",
	}

	fspSportsRanks = []string{
		"Заслуженный мастер спорта", "ЗМС",
		"Мастер спорта международного класса", "МСМК",
		"Мастер спорта", "МС",
		"Кандидат в мастера спорта", "КМС",
		"1-й спортивный разряд", "1 спортивный разряд", "1 разряд",
		"2-й спортивный разряд", "2 спортивный разряд", "2 разряд",
		"3-й спортивный разряд", "3 спортивный разряд", "3 разряд",
	}

	universities = []string{
		"МГУ", "МГТУ им. Баумана", "МГТУ", "МФТИ", "НИУ ВШЭ", "ВШЭ", "СПбГУ",
		"ИТМО", "УрФУ", "НГУ", "МИФИ", "МАИ", "СПбПУ", "КФУ", "ТГУ", "ТПУ",
	}
)

// ExtractAndParse takes binary PDF data or plain text bytes, extracts text, and parses structured candidate resume
func ExtractAndParse(data []byte) (ParsedResume, error) {
	text, err := ExtractTextFromPDF(data)
	if err != nil {
		return ParsedResume{}, err
	}
	return ParseResume(text), nil
}

// ParseResume extracts structured data from resume text
func ParseResume(rawText string) ParsedResume {
	res := ParsedResume{
		RawTextPreview: truncateString(rawText, 500),
	}

	lines := splitLines(rawText)

	// 1. Extract Contacts
	res.Contacts = extractContacts(rawText)

	// 2. Extract Full Name
	res.FullName, res.FirstName, res.LastName, res.MiddleName = extractName(lines)

	// 3. Extract Headline & Specialization
	res.Headline, res.Specialization, res.SuggestedGrade = extractHeadlineAndGrade(lines, rawText)

	// 4. Extract Location
	res.Location = extractLocation(rawText)

	// 5. Extract Skills
	res.Skills = extractSkills(rawText)

	// 6. Extract Experience
	res.TotalYearsExperience, res.Experiences = extractExperience(lines, rawText)

	// 7. Extract Education
	res.Education = extractEducation(lines, rawText)

	// 8. Extract FSP Mentions
	res.FSPMention = extractFSPMention(rawText)

	return res
}

func extractContacts(text string) ExtractedContacts {
	var c ExtractedContacts

	if email := emailRegex.FindString(text); email != "" {
		c.Email = email
	}

	if phone := phoneRegex.FindString(text); phone != "" {
		c.Phone = strings.TrimSpace(phone)
	}

	// Remove emails before matching Telegram to prevent matching @domain.com as telegram handle
	cleanTextForTG := emailRegex.ReplaceAllString(text, " ")
	tgRegexWithBoundary := regexp.MustCompile(`(?i)(?:t\.me/|(?:^|[\s,;:(\[])@)([a-zA-Z0-9_]{4,})`)
	if tgMatches := tgRegexWithBoundary.FindStringSubmatch(cleanTextForTG); len(tgMatches) > 1 {
		c.Telegram = "@" + tgMatches[1]
	}

	if ghMatches := githubRegex.FindStringSubmatch(text); len(ghMatches) > 1 {
		c.GitHub = "https://github.com/" + ghMatches[1]
	}

	if inMatches := linkedinRegex.FindStringSubmatch(text); len(inMatches) > 1 {
		c.LinkedIn = "https://linkedin.com/in/" + inMatches[1]
	}

	return c
}

func isPatronymic(w string) bool {
	lower := strings.ToLower(w)
	return strings.HasSuffix(lower, "вич") ||
		strings.HasSuffix(lower, "вна") ||
		strings.HasSuffix(lower, "ич") ||
		strings.HasSuffix(lower, "ична") ||
		strings.HasSuffix(lower, "ишна") ||
		strings.HasSuffix(lower, "vich") ||
		strings.HasSuffix(lower, "vna")
}

func isSurname(w string) bool {
	lower := strings.ToLower(w)
	return strings.HasSuffix(lower, "ов") || strings.HasSuffix(lower, "ова") ||
		strings.HasSuffix(lower, "ев") || strings.HasSuffix(lower, "ева") ||
		strings.HasSuffix(lower, "ёв") || strings.HasSuffix(lower, "ёва") ||
		strings.HasSuffix(lower, "ин") || strings.HasSuffix(lower, "ина") ||
		strings.HasSuffix(lower, "ский") || strings.HasSuffix(lower, "ская") ||
		strings.HasSuffix(lower, "цкий") || strings.HasSuffix(lower, "цкая") ||
		strings.HasSuffix(lower, "ых") || strings.HasSuffix(lower, "их") ||
		strings.HasSuffix(lower, "ov") || strings.HasSuffix(lower, "ova") ||
		strings.HasSuffix(lower, "ev") || strings.HasSuffix(lower, "eva") ||
		strings.HasSuffix(lower, "in") || strings.HasSuffix(lower, "ina") ||
		strings.HasSuffix(lower, "skiy") || strings.HasSuffix(lower, "skaya")
}

func extractName(lines []string) (fullName, first, last, middle string) {
	// Look at first 8 lines for name pattern
	namePattern := regexp.MustCompile(`^[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+(?:\s+[А-ЯЁ][а-яё]+)?$`)
	latinPattern := regexp.MustCompile(`^[A-Z][a-z]+\s+[A-Z][a-z]+(?:\s+[A-Z][a-z]+)?$`)

	for i := 0; i < len(lines) && i < 8; i++ {
		line := strings.TrimSpace(lines[i])
		if namePattern.MatchString(line) || latinPattern.MatchString(line) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				fullName = line
				if len(parts) == 2 {
					if isSurname(parts[0]) && !isSurname(parts[1]) {
						last = parts[0]
						first = parts[1]
					} else {
						first = parts[0]
						last = parts[1]
					}
				} else if len(parts) >= 3 {
					// Check patronymic position
					if isPatronymic(parts[1]) {
						// e.g. "Александр Дмитриевич Смирнов" (First Middle Last)
						first = parts[0]
						middle = parts[1]
						last = parts[2]
					} else if isPatronymic(parts[2]) {
						// e.g. "Смирнов Александр Дмитриевич" (Last First Middle)
						last = parts[0]
						first = parts[1]
						middle = parts[2]
					} else if isSurname(parts[0]) {
						last = parts[0]
						first = parts[1]
						middle = parts[2]
					} else {
						first = parts[0]
						middle = parts[1]
						last = parts[2]
					}
				}
				return
			}
		}
	}

	// Fallback to first non-empty line
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		parts := strings.Fields(firstLine)
		if len(parts) >= 2 && len(firstLine) < 50 {
			fullName = firstLine
			first = parts[0]
			last = parts[1]
		}
	}

	return
}

func extractHeadlineAndGrade(lines []string, text string) (headline, spec, grade string) {
	lowerText := strings.ToLower(text)

	// Determine grade
	switch {
	case strings.Contains(lowerText, "lead") || strings.Contains(lowerText, "руководитель") || strings.Contains(lowerText, "тимлид"):
		grade = "Lead"
	case strings.Contains(lowerText, "senior") || strings.Contains(lowerText, "сеньор") || strings.Contains(lowerText, "ведущий"):
		grade = "Senior"
	case strings.Contains(lowerText, "middle") || strings.Contains(lowerText, "мидл"):
		grade = "Middle"
	case strings.Contains(lowerText, "junior") || strings.Contains(lowerText, "джуниор") || strings.Contains(lowerText, "стажер"):
		grade = "Junior"
	default:
		grade = "Middle"
	}

	// Determine specialization
	switch {
	case strings.Contains(lowerText, "backend") || strings.Contains(lowerText, "бэкенд") || strings.Contains(lowerText, "golang") || strings.Contains(lowerText, "python разработчик"):
		spec = "Backend"
	case strings.Contains(lowerText, "frontend") || strings.Contains(lowerText, "фронтенд") || strings.Contains(lowerText, "react"):
		spec = "Frontend"
	case strings.Contains(lowerText, "mobile") || strings.Contains(lowerText, "ios") || strings.Contains(lowerText, "android"):
		spec = "Mobile"
	case strings.Contains(lowerText, "devops") || strings.Contains(lowerText, "sre"):
		spec = "DevOps"
	case strings.Contains(lowerText, "data") || strings.Contains(lowerText, "machine learning") || strings.Contains(lowerText, "ml"):
		spec = "Data / AI"
	default:
		spec = "Software Engineer"
	}

	// Scan lines 1..8 for explicit position headline
	for i := 0; i < len(lines) && i < 8; i++ {
		line := strings.TrimSpace(lines[i])
		low := strings.ToLower(line)
		if strings.Contains(low, "developer") || strings.Contains(low, "разработчик") ||
			strings.Contains(low, "engineer") || strings.Contains(low, "инженер") ||
			strings.Contains(low, "backend") || strings.Contains(low, "frontend") {
			headline = line
			return
		}
	}

	headline = grade + " " + spec + " Developer"
	return
}

func extractLocation(text string) string {
	for _, city := range knownCities {
		if strings.Contains(text, city) {
			return city
		}
	}
	return "г. Москва"
}

func extractSkills(text string) []string {
	lower := strings.ToLower(text)
	foundMap := make(map[string]bool)
	var skills []string

	// Word boundary matching for technology aliases
	for alias, canonical := range skillDictionary {
		if foundMap[canonical] {
			continue
		}

		pattern := `(?i)(?:^|[^a-zA-Z0-9#+])` + regexp.QuoteMeta(alias) + `(?:[^a-zA-Z0-9#+]|$)`
		if matched, _ := regexp.MatchString(pattern, lower); matched {
			foundMap[canonical] = true
			skills = append(skills, canonical)
		}
	}

	return skills
}

func extractExperience(lines []string, text string) (float64, []ExperienceItem) {
	var totalYears float64
	var experiences []ExperienceItem

	// 1. Direct mention e.g. "Опыт работы: 4 года"
	if matches := expYearsRegex.FindStringSubmatch(text); len(matches) > 1 {
		valStr := strings.ReplaceAll(matches[1], ",", ".")
		if val, err := strconv.ParseFloat(valStr, 64); err == nil {
			totalYears = val
		}
	}

	// 2. Scan timeline periods
	currentYear := time.Now().Year()
	periodMatches := periodRegex.FindAllStringSubmatch(text, -1)
	var maxSpan float64

	for _, m := range periodMatches {
		if len(m) > 2 {
			startYear, err1 := strconv.Atoi(m[1])
			endYear := currentYear
			if m[2] != "настоящее время" && m[2] != "н.в." && m[2] != "present" {
				if ey, err2 := strconv.Atoi(m[2]); err2 == nil {
					endYear = ey
				}
			}
			if err1 == nil && endYear >= startYear {
				diff := float64(endYear - startYear)
				if diff > maxSpan && diff <= 40 {
					maxSpan = diff
				}
			}
		}
	}

	if totalYears == 0 && maxSpan > 0 {
		totalYears = maxSpan
	} else if totalYears == 0 {
		totalYears = 3.0 // Reasonable median fallback
	}

	// 3. Scan experience items
	for i, line := range lines {
		if periodRegex.MatchString(line) {
			periodStr := strings.TrimSpace(line)
			company := "Компания"
			position := "Разработчик"

			// Check previous or next lines for company/position
			if i > 0 && len(lines[i-1]) < 60 {
				company = strings.TrimSpace(lines[i-1])
			}
			if i+1 < len(lines) && len(lines[i+1]) < 60 {
				position = strings.TrimSpace(lines[i+1])
			}

			experiences = append(experiences, ExperienceItem{
				Company:  company,
				Position: position,
				Period:   periodStr,
			})
		}
	}

	return totalYears, experiences
}

func extractEducation(lines []string, text string) []EducationItem {
	var edItems []EducationItem

	for _, u := range universities {
		if strings.Contains(text, u) {
			edItems = append(edItems, EducationItem{
				Institution: u,
				Degree:      "Высшее техническое",
			})
		}
	}

	return edItems
}

func extractFSPMention(text string) *FSPMention {
	hasMention := false
	var fspID string
	var rank string
	var achievements []string

	if id := fspIDRegex.FindString(text); id != "" {
		hasMention = true
		fspID = strings.ToUpper(id)
	}

	lower := strings.ToLower(text)
	if strings.Contains(lower, "фсп") || strings.Contains(lower, "спортивное программирование") || strings.Contains(lower, "icpc") {
		hasMention = true
	}

	for _, r := range fspSportsRanks {
		if strings.Contains(text, r) {
			hasMention = true
			rank = r
			break
		}
	}

	if strings.Contains(lower, "чемпионат") {
		achievements = append(achievements, "Участник чемпионата по программированию")
	}
	if strings.Contains(lower, "хакатон") {
		achievements = append(achievements, "Призёр хакатона")
	}

	if !hasMention {
		return nil
	}

	return &FSPMention{
		HasFSP:       true,
		FSPID:        fspID,
		SportsRank:   rank,
		Achievements: achievements,
	}
}

func splitLines(text string) []string {
	var lines []string
	raw := strings.Split(text, "\n")
	for _, l := range raw {
		trimmed := strings.TrimSpace(l)
		if len(trimmed) > 0 {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
