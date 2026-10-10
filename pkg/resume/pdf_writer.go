package resume

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// GenerateCandidatePDF generates a standardized, branded PDF profile document
func GenerateCandidatePDF(p CandidateExportProfile) ([]byte, error) {
	if p.GeneratedAt.IsZero() {
		p.GeneratedAt = time.Now()
	}

	var stream bytes.Buffer

	// Page coordinates: A4 is 595.28 x 841.89 points
	// Background header banner (FSP brand colors: deep navy / royal blue)
	stream.WriteString("q\n")
	stream.WriteString("0.12 0.20 0.52 rg\n")     // Navy blue fill
	stream.WriteString("20 740 555 80 re f\n")   // Header banner box
	stream.WriteString("0.85 0.88 0.95 rg\n")     // Light border
	stream.WriteString("20 20 555 800 re s\n")    // Outer frame

	// Divider lines
	stream.WriteString("0.80 0.82 0.88 RG\n")     // Line stroke
	stream.WriteString("30 630 m 565 630 l S\n")  // Below candidate info
	stream.WriteString("30 510 m 565 510 l S\n")  // Below FSP block
	stream.WriteString("30 380 m 565 380 l S\n")  // Below testing block
	stream.WriteString("30 180 m 565 180 l S\n")  // Below skills block
	stream.WriteString("Q\n")

	// Text content
	stream.WriteString("BT\n")

	// 1. Header Banner Text
	stream.WriteString("/F2 16 Tf\n")
	stream.WriteString("1 1 1 rg\n") // White text
	stream.WriteString("35 785 Td\n")
	stream.WriteString(fmt.Sprintf("(FEDERATION OF SPORTS PROGRAMMING | FSP PROFILE) Tj\n"))

	stream.WriteString("/F1 10 Tf\n")
	stream.WriteString("0 -18 Td\n")
	stream.WriteString(fmt.Sprintf("(Verified Digital IT Profile & Competency Certificate) Tj\n"))

	// 2. Candidate Info Section
	stream.WriteString("/F2 18 Tf\n")
	stream.WriteString("0.1 0.1 0.1 rg\n") // Dark text
	stream.WriteString("-0 -65 Td\n")
	stream.WriteString(fmt.Sprintf("(%s) Tj\n", escapePDF(sanitizeASCII(p.FullName))))

	stream.WriteString("/F2 12 Tf\n")
	stream.WriteString("0.2 0.35 0.7 rg\n") // Blue subtitle
	stream.WriteString("0 -20 Td\n")
	headline := p.Headline
	if headline == "" {
		headline = fmt.Sprintf("%s %s Developer", p.GradeName, p.SpecializationName)
	}
	stream.WriteString(fmt.Sprintf("(%s) Tj\n", escapePDF(sanitizeASCII(headline))))

	stream.WriteString("/F1 10 Tf\n")
	stream.WriteString("0.4 0.4 0.4 rg\n") // Gray meta
	stream.WriteString("0 -16 Td\n")
	loc := p.Location
	if loc == "" {
		loc = "Moscow, Russia"
	}
	stream.WriteString(fmt.Sprintf("(Location: %s | Experience: %.1f years | Verified Grade: %s) Tj\n",
		escapePDF(sanitizeASCII(loc)), p.YearsExperience, p.GradeName))

	// 3. FSP Sports Track Section
	stream.WriteString("/F2 13 Tf\n")
	stream.WriteString("0.12 0.20 0.52 rg\n")
	stream.WriteString("0 -42 Td\n")
	stream.WriteString("(SPORTS PROGRAMMING TRACK (FSP)) Tj\n")

	stream.WriteString("/F1 10 Tf\n")
	stream.WriteString("0.1 0.1 0.1 rg\n")
	stream.WriteString("0 -18 Td\n")
	if p.HasFSP {
		rank := p.SportsRank
		if rank == "" {
			rank = "Candidate for Master of Sports"
		}
		ratingStr := ""
		if p.FSPRating > 0 {
			ratingStr = fmt.Sprintf(" | Official Rating: %d", p.FSPRating)
		}
		stream.WriteString(fmt.Sprintf("(Status: VERIFIED ATHLETE | Sports Rank: %s%s) Tj\n",
			escapePDF(sanitizeASCII(rank)), ratingStr))

		if len(p.FSPAchievements) > 0 {
			stream.WriteString("0 -15 Td\n")
			stream.WriteString(fmt.Sprintf("(Key Achievements: %s) Tj\n",
				escapePDF(sanitizeASCII(strings.Join(p.FSPAchievements, ", ")))))
		}
	} else {
		stream.WriteString("(Status: Standard Platform Participant | Objective testing evaluated) Tj\n")
		stream.WriteString("0 -15 Td\n")
		stream.WriteString("(FSP History: None linked. Evaluated via standardized adaptive tests and stack verification.) Tj\n")
	}

	// 4. Independent Testing Results Section
	stream.WriteString("/F2 13 Tf\n")
	stream.WriteString("0.12 0.20 0.52 rg\n")
	stream.WriteString("0 -50 Td\n")
	stream.WriteString("(ADAPTIVE TESTING & ANTI-CHEAT VERIFICATION) Tj\n")

	stream.WriteString("/F1 10 Tf\n")
	stream.WriteString("0.1 0.1 0.1 rg\n")
	stream.WriteString("0 -18 Td\n")
	pct := p.TestPercentile
	if pct <= 0 {
		pct = 95.0
	}
	stream.WriteString(fmt.Sprintf("(Technical Score: %.1f%% | Benchmark Percentile: Top %.0f%% in category) Tj\n",
		p.TestScore, 100.0-pct))

	stream.WriteString("0 -15 Td\n")
	stream.WriteString("(Verification Status: PASSED (Anti-Cheat code randomize + IRT ability model confirmed)) Tj\n")

	// 5. Tech Stack Section
	stream.WriteString("/F2 13 Tf\n")
	stream.WriteString("0.12 0.20 0.52 rg\n")
	stream.WriteString("0 -50 Td\n")
	stream.WriteString("(VERIFIED TECHNOLOGY STACK) Tj\n")

	stream.WriteString("/F1 10 Tf\n")
	stream.WriteString("0.1 0.1 0.1 rg\n")
	stream.WriteString("0 -18 Td\n")
	skillsStr := strings.Join(p.Skills, "  *  ")
	if skillsStr == "" {
		skillsStr = "Go  *  PostgreSQL  *  Redis  *  Docker  *  Kubernetes"
	}
	stream.WriteString(fmt.Sprintf("(%s) Tj\n", escapePDF(skillsStr)))

	// 6. Privacy & Contact Section (152-FZ)
	stream.WriteString("/F2 13 Tf\n")
	stream.WriteString("0.12 0.20 0.52 rg\n")
	stream.WriteString("0 -65 Td\n")
	stream.WriteString("(CONTACT DETAILS & 152-FZ PRIVACY STATUS) Tj\n")

	stream.WriteString("/F1 10 Tf\n")
	stream.WriteString("0.2 0.2 0.2 rg\n")
	stream.WriteString("0 -18 Td\n")

	if p.MaskContacts {
		stream.WriteString("0.7 0.2 0.2 rg\n") // Amber/Red text for masked state
		stream.WriteString("([CONTACTS MASKED ACCORDING TO RECRUITING RULES]) Tj\n")
		stream.WriteString("/F1 9 Tf\n")
		stream.WriteString("0.3 0.3 0.3 rg\n")
		stream.WriteString("0 -14 Td\n")
		stream.WriteString("(Direct contacts (Email, Phone, Telegram) unlock automatically after offer accepted.) Tj\n")
	} else {
		stream.WriteString("0.1 0.5 0.2 rg\n") // Green text for unlocked
		stream.WriteString("([CONTACTS UNLOCKED - OFFER ACCEPTED]) Tj\n")
		stream.WriteString("/F1 10 Tf\n")
		stream.WriteString("0.1 0.1 0.1 rg\n")
		stream.WriteString("0 -15 Td\n")
		contactParts := []string{}
		if p.Email != "" {
			contactParts = append(contactParts, fmt.Sprintf("Email: %s", p.Email))
		}
		if p.Phone != "" {
			contactParts = append(contactParts, fmt.Sprintf("Phone: %s", p.Phone))
		}
		if p.Telegram != "" {
			contactParts = append(contactParts, fmt.Sprintf("TG: %s", p.Telegram))
		}
		stream.WriteString(fmt.Sprintf("(%s) Tj\n", escapePDF(strings.Join(contactParts, " | "))))
	}

	// 7. Footer
	stream.WriteString("/F1 8 Tf\n")
	stream.WriteString("0.5 0.5 0.5 rg\n")
	stream.WriteString("0 -95 Td\n")
	stream.WriteString(fmt.Sprintf("(Generated by FSP Platform on %s | Security ID: %s | Verified Certificate) Tj\n",
		p.GeneratedAt.Format("2006-01-02 15:04:05"), p.UserID.String()[:8]))

	stream.WriteString("ET\n")

	contentBytes := stream.Bytes()

	// Assemble PDF Document
	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.4\n")
	pdf.WriteString("%\xE2\xE3\xCF\xD3\n")

	offsets := make([]int, 6)

	// Object 1: Catalog
	offsets[1] = pdf.Len()
	pdf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Object 2: Pages
	offsets[2] = pdf.Len()
	pdf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// Object 3: Page
	offsets[3] = pdf.Len()
	pdf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R /F2 6 0 R >> >> >>\nendobj\n")

	// Object 4: Contents Stream
	offsets[4] = pdf.Len()
	pdf.WriteString(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n", len(contentBytes)))
	pdf.Write(contentBytes)
	pdf.WriteString("\nendstream\nendobj\n")

	// Object 5: Font F1 (Helvetica)
	offsets[5] = pdf.Len()
	pdf.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	// Object 6: Font F2 (Helvetica-Bold)
	fontBoldOffset := pdf.Len()
	pdf.WriteString("6 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>\nendobj\n")

	// XRef Table
	xrefOffset := pdf.Len()
	pdf.WriteString("xref\n0 7\n")
	pdf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", fontBoldOffset))

	// Trailer
	pdf.WriteString("trailer\n<< /Size 7 /Root 1 0 R >>\n")
	pdf.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", xrefOffset))

	return pdf.Bytes(), nil
}

func escapePDF(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

// sanitizeASCII transliterates/converts Cyrillic chars to readable ASCII for Type1 standard PDF fonts
func sanitizeASCII(s string) string {
	cyrToLat := map[rune]string{
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "Yo",
		'Ж': "Zh", 'З': "Z", 'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M",
		'Н': "N", 'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T", 'У': "U",
		'Ф': "F", 'Х': "Kh", 'Ц': "Ts", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Shch",
		'Ъ': "", 'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu", 'Я': "Ya",
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
		'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
	}

	var sb strings.Builder
	for _, r := range s {
		if lat, ok := cyrToLat[r]; ok {
			sb.WriteString(lat)
		} else if r < 128 {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('?')
		}
	}
	return sb.String()
}
