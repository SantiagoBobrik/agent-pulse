package domain

type Provider string

var Providers = struct {
	Claude Provider
	Gemini Provider
}{
	Claude: "claude",
	Gemini: "gemini",
}

func (p Provider) IsValid() bool {
	switch p {
	case Providers.Claude, Providers.Gemini:
		return true
	}
	return false
}
