package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"golang.org/x/text/language"
)

// Wikipedia is a Wiki* instant answer,
// including Wikidata/Wikiquote/Wiktionary
type Wikipedia struct {
	wikipedia.Fetcher
	Answer
}

func (w *Wikipedia) setQuery(r *http.Request, qv string) Answerer {
	w.Answer.setQuery(r, qv)
	return w
}

func (w *Wikipedia) setUserAgent(r *http.Request) Answerer {
	return w
}

func (w *Wikipedia) setLanguage(lang language.Tag) Answerer {
	w.language = lang
	return w
}

func (w *Wikipedia) setType() Answerer {
	w.Type = "wiki"
	return w
}

// trigger words
// age ---> for "how old is x?" we need to change our triggerfuncs to just a regex
const age = "age"
const howOldIs = "how old is"

// birthday
const birthday = "birthday"
const born = "born"

// death
const death = "death"
const died = "died"

// height
const height = "height"
const howTallis = "how tall is"
const howTallwas = "how tall was"

// weight
// will fail on "how much does x weigh?"
const mass = "mass"
const weigh = "weigh"
const weight = "weight"

// quotes
const quote = "quote"
const quotes = "quotes"

// definitions
const define = "define"
const definition = "definition"

func (w *Wikipedia) setRegex() Answerer {
	triggers := []string{
		age, howOldIs,
		birthday, born,
		death, died,
		howTallis, howTallwas, height,
		mass, weigh, weight,
		quote, quotes,
		define, definition,
	}

	t := strings.Join(triggers, "|")
	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))
	w.regex = append(w.regex, regexp.MustCompile(`^(?P<remainder>.*)$`)) // this needs to be last regex here

	return w
}

// Birthday is a person's date of birth
type Birthday struct {
	Birthday wikipedia.DateTime `json:"birthday,omitempty"`
}

// Death is a person's date of death
// TODO: add place of death, cause, etc.
type Death struct {
	Death wikipedia.DateTime `json:"death,omitempty"`
}

// Age is a person's current age (in years) or age when they died
type Age struct {
	*Birthday `json:"birthday,omitempty"`
	*Death    `json:"death,omitempty"`
}

// TODO: Return the Title (and perhaps Image???) as
// confirmation that we fetched the right asset.
func (w *Wikipedia) solve(r *http.Request) Answerer {
	item, err := w.Fetch(w.remainder, w.language)
	if err != nil {
		w.Err = err
		return w
	}

	switch w.triggerWord {
	case age, howOldIs, birthday, born:
		if len(item.Birthday) == 0 {
			return w
		}

		w.Type = "wikidata birthday"
		b := &Birthday{item.Birthday[0]}

		if w.triggerWord == "age" || w.triggerWord == "how old is" {
			w.Type = "wikidata age"

			a := &Age{
				Birthday: b,
			}

			if len(item.Death) > 0 {
				a.Death = &Death{item.Death[0]}
			}

			w.Data.Solution = a

			return w
		}

		w.Data.Solution = b
	case death, died:
		if len(item.Death) > 0 {
			w.Type = "wikidata death"
			w.Data.Solution = &Death{item.Death[0]}
		}
	case howTallis, howTallwas, height:
		if len(item.Height) == 0 {
			return w
		}

		w.Type = "wikidata height"
		w.Data.Solution = item.Height
	case mass, weigh, weight:
		if len(item.Weight) == 0 {
			return w
		}

		w.Type = "wikidata weight"
		w.Data.Solution = item.Weight
	case quote, quotes:
		if len(item.Wikiquote.Quotes) == 0 {
			return w
		}

		w.Type = "wikiquote"
		w.Data.Solution = item.Wikiquote.Quotes
	case define, definition:
		if len(item.Wiktionary.Definitions) == 0 {
			return w
		}

		w.Type = "wiktionary"
		w.Data.Solution = item.Wiktionary
	default: // full Wikipedia box
		w.Type = "wikipedia"
		w.Data.Solution = item
	}

	return w
}

func (w *Wikipedia) setCache() Answerer {
	w.Cache = true
	return w
}

func (w *Wikipedia) tests() []test {

	tests := []test{
		{
			query: "Bob Marley age",
			expected: []Data{
				{
					Type:      "wikidata age",
					Triggered: true,
					Solution: &Age{
						Birthday: &Birthday{
							Birthday: wikipedia.DateTime{
								Value:    "1945-02-06T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
						Death: &Death{
							Death: wikipedia.DateTime{
								Value:    "1981-05-11T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "Jimi hendrix birthday",
			expected: []Data{
				{
					Type:      "wikidata birthday",
					Triggered: true,
					Solution: &Birthday{
						Birthday: wikipedia.DateTime{
							Value:    "1942-11-27T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "death jimi hendrix",
			expected: []Data{
				{
					Type:      "wikidata death",
					Triggered: true,
					Solution: &Death{
						Death: wikipedia.DateTime{
							Value:    "1970-09-18T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "shaquille o'neal height",
			expected: []Data{
				{
					Type:      "wikidata height",
					Triggered: true,
					Solution: []wikipedia.Quantity{
						{
							Amount: "2.16",
							Unit:   wikipedia.Wikidata{ID: "Q11573"},
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "shaquille o'neal weight",
			expected: []Data{
				{
					Type:      "wikidata weight",
					Triggered: true,
					Solution: []wikipedia.Quantity{
						{
							Amount: "147",
							Unit:   wikipedia.Wikidata{ID: "Q11573"},
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "Michael Jordan quotes",
			expected: []Data{
				{
					Type:      "wikiquote",
					Triggered: true,
					Solution: []string{
						"I can accept failure. Everyone fails at something. But I can't accept not trying (no hard work)",
						"ball is life",
					},
					Cache: true,
				},
			},
		},
		{
			query: "define guitar",
			expected: []Data{
				{
					Type:      "wiktionary",
					Triggered: true,
					Solution: wikipedia.Wiktionary{
						Title: "guitar",
						Definitions: []*wikipedia.Definition{
							{Part: "noun", Meaning: "musical instrument"},
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "jimi hendrix",
			expected: []Data{
				{
					Type:      "wikipedia",
					Triggered: true,
					Solution: &wikipedia.Item{
						Wikidata: &wikipedia.Wikidata{
							Claims: &wikipedia.Claims{
								Birthday: []wikipedia.DateTime{
									{
										Value:    "1942-11-27T00:00:00Z",
										Calendar: wikipedia.Wikidata{ID: "Q1985727"},
									},
								},
								Death: []wikipedia.DateTime{
									{
										Value:    "1970-09-18T00:00:00Z",
										Calendar: wikipedia.Wikidata{ID: "Q1985727"},
									},
								},
							},
						},
					},
					Cache: true,
				},
			},
		},
	}

	return tests
}
