package instant

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"

	"github.com/jivesearch/jivesearch/instant/contributors"
)

// Coin is an instant answer
type Coin struct {
	Answer
}

func (c *Coin) setQuery(r *http.Request, qv string) answerer {
	c.Answer.setQuery(r, qv)
	return c
}

func (c *Coin) setUserAgent(r *http.Request) answerer {
	return c
}

func (c *Coin) setType() answerer {
	c.Type = "coin toss"
	return c
}

func (c *Coin) setContributors() answerer {
	c.Contributors = contributors.Load(
		[]string{
			"brentadamson",
		},
	)
	return c
}

func (c *Coin) setRegex() answerer {
	triggers := []string{
		"flip a coin", "heads or tails", "coin toss",
	}

	t := strings.Join(triggers, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, t)))

	return c
}

func (c *Coin) setSolution() answerer {
	choices := []string{"Heads", "Tails"}

	c.Text = choices[rand.Intn(2)]

	return c
}

func (c *Coin) setCache() answerer {
	c.Cache = false
	return c
}

func (c *Coin) tests() []test {
	contrib := contributors.Load([]string{"brentadamson"})

	tests := []test{}

	for _, q := range []string{"flip a coin", "heads or tails", "Coin Toss"} {
		tst := test{
			query: q,
			expected: []Solution{
				{
					Type:         "coin toss",
					Triggered:    true,
					Contributors: contrib,
					Text:         "Heads",
					Cache:        false,
				},
				{
					Type:         "coin toss",
					Triggered:    true,
					Contributors: contrib,
					Text:         "Tails",
					Cache:        false,
				},
			},
		}

		tests = append(tests, tst)
	}

	return tests
}
