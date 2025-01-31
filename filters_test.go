package pongo2addons

import (
	"errors"
	"testing"
	"time"

	. "github.com/iostrovok/check"

	"github.com/flosch/pongo2/v6"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

// A wrapprt of pongo2.RenderTemplateString
func getResult(s string, ctx pongo2.Context) string {
	result, _ := pongo2.RenderTemplateString(s, ctx)
	return result
}

type TestSuite1 struct{}

var _ = Suite(&TestSuite1{})

func (s *TestSuite1) TestFilters(c *C) {
	// Markdown
	c.Assert(getResult("{{ \"**test**\"|markdown }}", nil), Equals, "<p><strong>test</strong></p>\n")

	// Slugify
	c.Assert(getResult("{{ \"this is ä test!\"|slugify }}", nil), Equals, "this-is-a-test")

	// Filesizeformat
	c.Assert(getResult("{{ 123456789|filesizeformat }}", nil), Equals, "118MiB")

	// Timesince/timeuntil
	baseDate := time.Date(2014, time.February, 1, 8, 30, 00, 00, time.UTC)
	futureDate := baseDate.Add(24*7*4*time.Hour + 2*time.Hour)
	c.Assert(getResult("{{ future_date|timeuntil:base_date }}",
		pongo2.Context{"base_date": baseDate, "future_date": futureDate}), Equals, "4 weeks from now")

	baseDate = time.Date(2014, time.February, 1, 8, 30, 00, 00, time.UTC)
	futureDate = baseDate.Add(2 * time.Hour)
	c.Assert(getResult("{{ future_date|timeuntil:base_date }}",
		pongo2.Context{"base_date": baseDate, "future_date": futureDate}), Equals, "2 hours from now")

	baseDate = time.Date(2014, time.February, 1, 8, 30, 00, 00, time.UTC)
	futureDate = baseDate.Add(2 * time.Hour)
	c.Assert(getResult("{{ base_date|timesince:future_date }}",
		pongo2.Context{"base_date": baseDate, "future_date": futureDate}), Equals, "2 hours ago")

	// Natural time
	baseDate = time.Date(2014, time.February, 1, 8, 30, 00, 00, time.UTC)
	futureDate = baseDate.Add(4 * time.Second)
	c.Assert(getResult("{{ base_date|naturaltime:future_date }}",
		pongo2.Context{"base_date": baseDate, "future_date": futureDate}), Equals, "4 seconds ago")

	// Naturalday
	today := time.Date(2014, time.February, 1, 8, 30, 00, 00, time.UTC)
	yesterday := today.Add(-24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	todayPlus3 := today.Add(3 * 24 * time.Hour)
	c.Assert(getResult("{{ date|naturalday:today }}",
		pongo2.Context{"date": today, "today": today}), Equals, "today")
	c.Assert(getResult("{{ date|naturalday:today }}",
		pongo2.Context{"date": yesterday, "today": today}), Equals, "yesterday")
	c.Assert(getResult("{{ date|naturalday:today }}",
		pongo2.Context{"date": tomorrow, "today": today}), Equals, "tomorrow")
	c.Assert(getResult("{{ date|naturalday:today }}",
		pongo2.Context{"date": todayPlus3, "today": today}), Equals, "3 days from now")

	// Intcomma
	c.Assert(getResult("{{ 123456789|intcomma }}", nil), Equals, "123,456,789")

	// Ordinal
	c.Assert(getResult("{{ 1|ordinal }} {{ 2|ordinal }} {{ 3|ordinal }} {{ 18241|ordinal }}", nil),
		Equals, "1st 2nd 3rd 18241st")

	// Truncatesentences
	c.Assert(getResult("{{ text|truncatesentences:3|safe }}", pongo2.Context{
		"text": `This is a first sentence with a 4.50 number. The second one is even more fun! Isn't it? Last sentence, okay.`}),
		Equals, "This is a first sentence with a 4.50 number. The second one is even more fun! Isn't it?")

	// Truncatesentences_html
	c.Assert(getResult("{{ text|truncatesentences_html:2 }}", pongo2.Context{
		"text": `<div class="test"><ul><li>This is a first sentence with a 4.50 number.</li><li>The second one is even more fun! Isn't it?</li><li>Last sentence, okay.</li></ul></div>`}),
		Equals, `<div class="test"><ul><li>This is a first sentence with a 4.50 number.</li><li>The second one is even more fun!</li></ul></div>`)
	c.Assert(getResult("{{ text|truncatesentences_html:3 }}", pongo2.Context{
		"text": `<div class="test"><ul><li>This is a first sentence with a 4.50 number.</li><li>The second one is even more fun! Isn't it?</li><li>Last sentence, okay.</li></ul></div>`}),
		Equals, `<div class="test"><ul><li>This is a first sentence with a 4.50 number.</li><li>The second one is even more fun! Isn't it?</li></ul></div>`)

	c.Assert(getResult("{{ text|truncatesentences_html:0 }}", pongo2.Context{
		"text": `<div class="test"><ul><li>This is a first sentence with a 4.50 number.</li><li>The second one is even more fun! Isn't it?</li><li>Last sentence, okay.</li></ul></div>`}),
		Equals, ``)

	c.Assert(getResult("{{ text|truncatesentences_html:'-1' }}", pongo2.Context{
		"text": `<div class="test"><ul><li>This is a first sentence with a 4.50 number.</li><li>The second one is even more fun! Isn't it?</li><li>Last sentence, okay.</li></ul></div>`}),
		Equals, ``)

	// Random
	c.Assert(getResult("{{ array|random }}",
		pongo2.Context{"array": []int{42}}),
		Equals, "42")

}

func (s *TestSuite1) TestFiltersNumeric(c *C) {
	c.Assert(getResult("<h1>{{ text|iplus:1 }}</h1>", pongo2.Context{"text": `10`}), Equals, "<h1>11</h1>")
	c.Assert(getResult("<h1>{{ text|iplus:1.7 }}</h1>", pongo2.Context{"text": `10`}), Equals, "<h1>11</h1>")
	c.Assert(getResult("<h1>{{ text|iplus:'' }}</h1>", pongo2.Context{"text": `10`}), Equals, "<h1>10</h1>")
	c.Assert(getResult("<h1>{{ text|iplus:'0' }}</h1>", pongo2.Context{"text": "10"}), Equals, "<h1>10</h1>")
	c.Assert(getResult("<h1>{{ text|iplus:'-20' }}<h1>", pongo2.Context{"text": `10`}), Equals, "<h1>-10<h1>")

	c.Assert(getResult("<h1>{{ text|iminus:1 }}</h1>", pongo2.Context{"text": `10`}), Equals, "<h1>9</h1>")
	c.Assert(getResult("<h1>{{ text|iminus:1.3 }}</h1>", pongo2.Context{"text": `10`}), Equals, "<h1>9</h1>")
	c.Assert(getResult("<h1>{{ text|iminus:'' }}</h1>", pongo2.Context{"text": `10`}), Equals, "<h1>10</h1>")
	c.Assert(getResult("<h1>{{ text|iminus:'0' }}</h1>", pongo2.Context{"text": "10"}), Equals, "<h1>10</h1>")
	c.Assert(getResult("<h1>{{ text|iminus:'-20' }}<h1>", pongo2.Context{"text": `10`}), Equals, "<h1>30<h1>")

	c.Assert(getResult("<h1>{{ text|imultiply:1 }}</h1>", pongo2.Context{"text": `10`}), Equals, "<h1>10</h1>")
	c.Assert(getResult("<h1>{{ text|imultiply:1.5 }}</h1>", pongo2.Context{"text": `10`}), Equals, "<h1>10</h1>")
	c.Assert(getResult("<h1>{{ text|imultiply:11 }}</h1>", pongo2.Context{"text": `10`}), Equals, "<h1>110</h1>")
	c.Assert(getResult("<h1>{{ text|imultiply:'' }}</h1>", pongo2.Context{"text": `10`}), Equals, "<h1>0</h1>")
	c.Assert(getResult("<h1>{{ text|imultiply:'0' }}</h1>", pongo2.Context{"text": "10"}), Equals, "<h1>0</h1>")
	c.Assert(getResult("<h1>{{ text|imultiply:'-20' }}<h1>", pongo2.Context{"text": `10`}), Equals, "<h1>-200<h1>")
}

func (s *TestSuite1) TestFilterPrintError(c *C) {
	err := errors.New("simple error")
	c.Assert(getResult("<h1>{{ err|printerror }}</h1>", pongo2.Context{"err": err}), Equals, "<h1>simple error</h1>")

	err2 := "simple error string"
	c.Assert(getResult("<h1>{{ err|printerror }}</h1>", pongo2.Context{"err": err2}), Equals, "<h1>simple error string</h1>")

	err3 := 10
	c.Assert(getResult("<h1>{{ err|printerror }}</h1>", pongo2.Context{"err": err3}), Equals, "<h1>10</h1>")
}

func (s *TestSuite1) TestFilterSolidLineBreaksBR(c *C) {
	text := "simpleerror"
	c.Assert(getResult("{{ text|solidlinebreaksbr: '6'|safe }}", pongo2.Context{"text": text}), Equals, "simple<br />error")

	c.Assert(getResult("{{ text|solidlinebreaksbr: '6, ' }}", pongo2.Context{"text": text}), Equals, "simple error")

	text = "one 🐘 and three 🐋"
	c.Assert(getResult("{{ text|solidlinebreaksbr: '6'|safe }}", pongo2.Context{"text": text}), Equals, "one 🐘 <br />and th<br />ree 🐋")

	text = "one 🐘 and three 🐋"
	c.Assert(getResult("{{ text|solidlinebreaksbr: '6'|safe }}", pongo2.Context{"text": text}), Equals, "one 🐘 <br />and th<br />ree 🐋")

	c.Assert(getResult("{{ 'one 🐘 and three 🐋'|solidlinebreaksbr: '6,<s>'|safe }}", pongo2.Context{"text": text}), Equals, "one 🐘 <s>and th<s>ree 🐋")
}

func (s *TestSuite1) TestFilterRange0(c *C) {
	text := "-"
	c.Assert(getResult("{% for t in ''|range0: '6' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "0-1-2-3-4-5-")

	c.Assert(getResult("{% for t in ''|range0: '0' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "")

	// error
	c.Assert(getResult("{% for t in ''|range0: '' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "")
	c.Assert(getResult("{% for t in ''|range0: '-1' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "")
}

func (s *TestSuite1) TestFilterRange0ToFrom(c *C) {
	text := "-"
	c.Assert(getResult(`{% for t in ''|range0: '-1,6' %}{{ t }}{{ text }}{% endfor %}`, pongo2.Context{"text": text}), Equals, "-1-0-1-2-3-4-5-")

	c.Assert(getResult("{% for t in ''|range0: '0' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "")
	c.Assert(getResult("{% for t in ''|range0: '10,12' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "10-11-")

	// error
	c.Assert(getResult("{% for t in ''|range0: '-1,-1' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "")
	c.Assert(getResult("{% for t in ''|range0: '10,-1' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "")
}

func (s *TestSuite1) TestFilterRange(c *C) {
	text := "-"
	c.Assert(getResult("{% for t in ''|range: '6' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "1-2-3-4-5-6-")

	c.Assert(getResult("{% for t in ''|range: '1' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "1-")

	// error
	c.Assert(getResult("{% for t in ''|range: '0' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "")
	c.Assert(getResult("{% for t in ''|range: '-1' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "")
}

func (s *TestSuite1) TestFilterRangeToFrom(c *C) {
	text := "-"
	c.Assert(getResult(`{% for t in ''|range: '-1,6' %}{{ t }}{{ text }}{% endfor %}`, pongo2.Context{"text": text}), Equals, "-1-0-1-2-3-4-5-6-")
	c.Assert(getResult(`{% for t in ''|range: '-1,6' %}{{ t }}.{% endfor %}`, pongo2.Context{"text": text}), Equals, "-1.0.1.2.3.4.5.6.")

	c.Assert(getResult("{% for t in ''|range: '0' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "")
	c.Assert(getResult("{% for t in ''|range: '10,12' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "10-11-12-")
	c.Assert(getResult("{% for t in ''|range: '-1,-1' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "-1-")

	// error
	c.Assert(getResult("{% for t in ''|range: '10,-1' %}{{ t }}{{ text }}{% endfor %}", pongo2.Context{"text": text}), Equals, "")
}

func (s *TestSuite1) TestFilterJSON(c *C) {
	T := struct {
		A string `json:"a"`
		B int    `json:"b"`
	}{
		A: "one 🐘 and three 🐋",
		B: 10,
	}

	c.Assert(getResult("{{ tojs|json|safe }}", pongo2.Context{"tojs": T}), Equals, `{"a":"one 🐘 and three 🐋","b":10}`)
	c.Assert(getResult("{{ tojs|json|safe }}", pongo2.Context{"tojs": nil}), Equals, `null`)
	c.Assert(getResult("{{ tojs|json|safe }}", pongo2.Context{"tojs": ""}), Equals, `""`)
	c.Assert(getResult("{{ tojs|json|safe }}", pongo2.Context{"tojs": "A"}), Equals, `"A"`)
}

func (s *TestSuite1) TestFilterJoinBr(c *C) {
	var T []string
	c.Assert(getResult("{{ T|joinBr }}", pongo2.Context{"T": T}), Equals, ``)

	T = make([]string, 0)
	c.Assert(getResult("{{ T|joinBr }}", pongo2.Context{"T": T}), Equals, ``)

	T = []string{"A"}
	c.Assert(getResult("{{ T|joinBr }}", pongo2.Context{"T": T}), Equals, `A`)

	T = []string{"A", "B"}
	c.Assert(getResult("{{ T|joinBr }}", pongo2.Context{"T": T}), Equals, "A\nB")

	T = []string{"A", "B", "C", "D", "E"}
	c.Assert(getResult("{{ T|joinBr }}", pongo2.Context{"T": T}), Equals, "A\nB\nC\nD\nE")

}
