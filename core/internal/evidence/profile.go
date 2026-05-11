package evidence

type NarrativeProfile struct {
	Name        string            `json:"name"`
	Keywords    []string          `json:"keywords"`
	WeightBonus map[string]float64 `json:"weight_bonus,omitempty"`
}

type ProfileMatch struct {
	Profile    string  `json:"profile"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason,omitempty"`
}

var DefaultProfiles = []NarrativeProfile{
	{
		Name:     "Quarian Migrant Fleet",
		Keywords: []string{"quarian", "migrant fleet", "flotilla", "pilgrimage", "vas", "nar"},
	},
	{
		Name:     "Cerberus Operations",
		Keywords: []string{"cerberus", "illusive man", "lazarus", "cell", "operative"},
	},
	{
		Name:     "Collector Threat",
		Keywords: []string{"collector", "harbinger", "swarm", "abduction", "colony"},
	},
	{
		Name:     "Geth Conflict",
		Keywords: []string{"geth", "heretic", "consensus", "platform", "synthetic"},
	},
	{
		Name:     "Omega Station",
		Keywords: []string{"omega", "aria", "afterlife", "mercenary", "blue suns", "eclipse", "blood pack"},
	},
	{
		Name:     "Citadel Politics",
		Keywords: []string{"citadel", "council", "spectre", "c-sec", "embassy", "ambassador"},
	},
	{
		Name:     "Normandy Crew",
		Keywords: []string{"normandy", "joker", "chakwas", "gardner", "crew"},
	},
}

func MatchProfile(text string) []ProfileMatch {
	var matches []ProfileMatch
	for _, profile := range DefaultProfiles {
		score := 0.0
		hits := 0
		for _, kw := range profile.Keywords {
			if containsCaseFold(text, kw) {
				hits++
			}
		}
		if hits > 0 {
			score = float64(hits) / float64(len(profile.Keywords))
			matches = append(matches, ProfileMatch{
				Profile:    profile.Name,
				Confidence: score,
			})
		}
	}
	return matches
}

func containsCaseFold(s, substr string) bool {
	return len(s) >= len(substr) && caseFoldContains(s, substr)
}

func caseFoldContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			pc := substr[j]
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if pc >= 'A' && pc <= 'Z' {
				pc += 32
			}
			if sc != pc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
