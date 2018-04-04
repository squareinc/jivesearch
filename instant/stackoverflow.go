package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jivesearch/jivesearch/instant/stackoverflow"
	"golang.org/x/text/language"
)

// StackOverflow is an instant answer
// Alternative (but out-of-date): http://archive.org/download/stackexchange/
type StackOverflow struct {
	stackoverflow.Fetcher
	Answer
}

// StackOverflowAnswer is a question and answer
type StackOverflowAnswer struct {
	Question string
	Link     string
	Answer   SOAnswer
}

// SOAnswer is the answer portion of a SO question
type SOAnswer struct {
	User string
	Text string
}

func (s *StackOverflow) setQuery(r *http.Request, qv string) answerer {
	s.Answer.setQuery(r, qv)
	return s
}

func (s *StackOverflow) setUserAgent(r *http.Request) answerer {
	return s
}

func (s *StackOverflow) setLanguage(lang language.Tag) answerer {
	s.language = lang
	return s
}

func (s *StackOverflow) setType() answerer {
	s.Type = "stackoverflow"
	return s
}

func (s *StackOverflow) setRegex() answerer {
	// https://stackoverflow.com/tags?page=1&tab=popular
	// Please convert to the trigger to the official tag in "tagger" func.
	// e.g. golang => go
	triggers := []string{
		"ajax", "android", "angular", "angularjs", "apache", "asp.net",
		"bash",
		"c", `c\+\+`, "c#", "css", "css3", "csv",
		"database", "django",
		"eclipse", "elasticsearch", "excel",
		"git", "golang", "go",
		"html", "html5",
		"ios", "iphone",
		"java", "javascript", "jquery", "json",
		"linux",
		"macos", "mac os", "matlab", "mongodb", "mysql",
		".net", "node.js",
		"objective-c", "oracle",
		"php", "perl", "postgresql", "python",
		"r", "reactjs", "redis", "regex", "regexp", "ruby-on-rails", "ruby",
		"scala", "selenium", "spring", "sql", "sqlite", "swift",
		"vba", "vue.js", "vue.js",
		"windows", "wordpress",
		"xml",
	}

	t := strings.Join(triggers, "|")
	s.regex = append(s.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	s.regex = append(s.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return s
}

// tagger converts a trigger word to a Stack Overflow tag
// golang => go
func tagger(txt string) string {
	var tag string

	switch txt {
	case "golang":
		tag = "go"
	case "mac os":
		tag = "macos"
	case "regexp":
		tag = "regex"
	default:
		tag = txt
	}

	return tag
}

func (s *StackOverflow) solve(r *http.Request) answerer {
	a := StackOverflowAnswer{}

	resp, err := s.Fetch(s.remainder, []string{tagger(s.triggerWord)})
	if err != nil {
		s.Err = err
		return s
	}

	// Find the answer with the most votes
	// Is there a way to return just this answer in the API?
	var score int
	for _, item := range resp.Items {
		a.Question = item.Title
		a.Link = item.Link

		for _, answer := range item.Answers {
			if answer.Score >= score {
				score = answer.Score
				a.Answer = SOAnswer{
					User: answer.Owner.DisplayName,
					Text: answer.Body,
				}
			}
		}
	}

	s.Data.Solution = a

	return s
}

func (s *StackOverflow) setCache() answerer {
	s.Cache = true
	return s
}

func (s *StackOverflow) tests() []test {
	typ := "stackoverflow"

	tests := []test{
		{
			query: "php loop",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: StackOverflowAnswer{
						Question: "How does PHP &#39;foreach&#39; actually work?",
						Link:     "https://stackoverflow.com/questions/10057671/how-does-php-foreach-actually-work",
						Answer: SOAnswer{
							User: "NikiC",
							Text: "an answer",
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "loop c++",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: StackOverflowAnswer{
						Question: "Some made-up question",
						Link:     "https://stackoverflow.com/questions/90210/c++-loop",
						Answer: SOAnswer{
							User: "JamesT",
							Text: "a very good answer",
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "golang loop",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: StackOverflowAnswer{
						Question: "Some made-up question",
						Link:     "https://stackoverflow.com/questions/90210/go-loop",
						Answer: SOAnswer{
							User: "Danny Zuko",
							Text: "a superbly good answer",
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "mac os loop",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: StackOverflowAnswer{
						Question: "Some made-up question",
						Link:     "https://stackoverflow.com/questions/90210/macos-loop",
						Answer: SOAnswer{
							User: "Danny Zuko",
							Text: "a superbly good answer",
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "regexp loop",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: StackOverflowAnswer{
						Question: "Some made-up question",
						Link:     "https://stackoverflow.com/questions/90210/regex-loop",
						Answer: SOAnswer{
							User: "Danny Zuko",
							Text: "a superbly good answer",
						},
					},
					Cache: true,
				},
			},
		},
	}

	return tests
}
