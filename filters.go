package pongo2addons

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/extemporalgenome/slug"
	"github.com/flosch/go-humanize"
	"github.com/flosch/pongo2/v6"
	"github.com/russross/blackfriday/v2"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	// Regulars
	pongo2.RegisterFilter("slugify", filterSlugify)
	pongo2.RegisterFilter("filesizeformat", filterFilesizeformat)
	pongo2.RegisterFilter("truncatesentences", filterTruncatesentences)
	pongo2.RegisterFilter("truncatesentences_html", filterTruncatesentencesHTML)
	pongo2.RegisterFilter("random", filterRandom)

	// Markup
	pongo2.RegisterFilter("markdown", filterMarkdown)

	// Humanize
	pongo2.RegisterFilter("timeuntil", filterTimeuntilTimesince)
	pongo2.RegisterFilter("timesince", filterTimeuntilTimesince)
	pongo2.RegisterFilter("naturaltime", filterTimeuntilTimesince)
	pongo2.RegisterFilter("naturalday", filterNaturalday)
	pongo2.RegisterFilter("intcomma", filterIntcomma)
	pongo2.RegisterFilter("ordinal", filterOrdinal)

	// Numeric, Plus and minus signs
	pongo2.RegisterFilter("iplus", filterIPlus)
	pongo2.RegisterFilter("iminus", filterIMinus)
	pongo2.RegisterFilter("imultiply", filterIMultiply)

	// Halpers
	// prints error as error.Error()
	pongo2.RegisterFilter("printerror", filterPrintError)

	// break line each N symbols
	pongo2.RegisterFilter("solidlinebreaksbr", filterSolidLineBreaksBR)

	// range integers for 0 to N-1
	pongo2.RegisterFilter("range0", filterRange0)
	// range integers for 1 to N
	pongo2.RegisterFilter("range", filterRange)

	// range integers for 1 to N
	pongo2.RegisterFilter("json", filterJSON)

	// range integers for 1 to N
	pongo2.RegisterFilter("joinBr", filterJoinBr)
}

func filterMarkdown(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsSafeValue(string(blackfriday.Run([]byte(in.String())))), nil
}

func filterSlugify(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(slug.Slug(in.String())), nil
}

func filterFilesizeformat(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(humanize.IBytes(uint64(in.Integer()))), nil
}

var filterTruncatesentencesRe = regexp.MustCompile(`(?U:.*[\w]{3,}.*([\d][\.!?][\D]|[\D][\.!?][\s]|[\n$]))`)

func filterTruncatesentences(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	count := param.Integer()
	if count <= 0 {
		return pongo2.AsValue(""), nil
	}
	sentencens := filterTruncatesentencesRe.FindAllString(strings.TrimSpace(in.String()), -1)
	return pongo2.AsValue(strings.TrimSpace(strings.Join(sentencens[:min(count, len(sentencens))], ""))), nil
}

// Taken from pongo2/filters_builtin.go
func filterTruncateHTMLHelper(value string, newOutput *bytes.Buffer, cond func() bool, fn func(c rune, s int, idx int) int, finalize func()) {
	vLen := len(value)
	tagStack := make([]string, 0)
	idx := 0

	for idx < vLen && !cond() {
		c, s := utf8.DecodeRuneInString(value[idx:])
		if c == utf8.RuneError {
			idx += s
			continue
		}

		if c == '<' {
			newOutput.WriteRune(c)
			idx += s // consume "<"

			if idx+1 < vLen {
				if value[idx] == '/' {
					// Close tag

					newOutput.WriteString("/")

					tag := ""
					idx++ // consume "/"

					for idx < vLen {
						c2, size2 := utf8.DecodeRuneInString(value[idx:])
						if c2 == utf8.RuneError {
							idx += size2
							continue
						}

						// End of tag found
						if c2 == '>' {
							idx++ // consume ">"
							break
						}
						tag += string(c2)
						idx += size2
					}

					if len(tagStack) > 0 {
						// Ideally, the close tag is TOP of tag stack
						// In malformed HTML, it must not be, so iterate through the stack and remove the tag
						for i := len(tagStack) - 1; i >= 0; i-- {
							if tagStack[i] == tag {
								// Found the tag
								tagStack[i] = tagStack[len(tagStack)-1]
								tagStack = tagStack[:len(tagStack)-1]
								break
							}
						}
					}

					newOutput.WriteString(tag)
					newOutput.WriteString(">")
				} else {
					// Open tag

					tag := ""

					params := false
					for idx < vLen {
						c2, size2 := utf8.DecodeRuneInString(value[idx:])
						if c2 == utf8.RuneError {
							idx += size2
							continue
						}

						newOutput.WriteRune(c2)

						// End of tag found
						if c2 == '>' {
							idx++ // consume ">"
							break
						}

						if !params {
							if c2 == ' ' {
								params = true
							} else {
								tag += string(c2)
							}
						}

						idx += size2
					}

					// Add tag to stack
					tagStack = append(tagStack, tag)
				}
			}
		} else {
			idx = fn(c, s, idx)
		}
	}

	finalize()

	for i := len(tagStack) - 1; i >= 0; i-- {
		tag := tagStack[i]
		// Close everything from the regular tag stack
		newOutput.WriteString(fmt.Sprintf("</%s>", tag))
	}
}

func filterTruncatesentencesHTML(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	count := param.Integer()
	if count <= 0 {
		return pongo2.AsValue(""), nil
	}

	value := in.String()
	newLen := max(param.Integer(), 0)

	newOutput := bytes.NewBuffer(nil)

	sentencefilter := 0

	filterTruncateHTMLHelper(value, newOutput, func() bool {
		return sentencefilter >= newLen
	}, func(_ rune, _ int, idx int) int {
		// Get next word
		wordFound := false

		for idx < len(value) {
			c2, size2 := utf8.DecodeRuneInString(value[idx:])
			if c2 == utf8.RuneError {
				idx += size2
				continue
			}

			if c2 == '<' {
				// HTML tag start, don't consume it
				return idx
			}

			newOutput.WriteRune(c2)
			idx += size2

			if (c2 == '.' && !(idx+1 < len(value) && value[idx+1] >= '0' && value[idx+1] <= '9')) ||
				c2 == '!' || c2 == '?' || c2 == '\n' {
				// Sentence ends here, stop capturing it now
				break
			} else {
				wordFound = true
			}
		}

		if wordFound {
			sentencefilter++
		}

		return idx
	}, func() {})

	return pongo2.AsSafeValue(newOutput.String()), nil
}

func filterRandom(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.CanSlice() {
		return nil, &pongo2.Error{
			Sender:    "filter:random",
			OrigError: errors.New("input is not sliceable"),
		}
	}

	if in.Len() <= 0 {
		return nil, &pongo2.Error{
			Sender:    "filter:random",
			OrigError: errors.New("input slice is empty"),
		}
	}

	return in.Index(rand.Intn(in.Len())), nil
}

func filterTimeuntilTimesince(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	basetime, isTime := in.Interface().(time.Time)
	if !isTime {
		return nil, &pongo2.Error{
			Sender:    "filter:timeuntil/timesince",
			OrigError: errors.New("time-value is not a time.Time-instance"),
		}
	}
	var paramtime time.Time
	if !param.IsNil() {
		paramtime, isTime = param.Interface().(time.Time)
		if !isTime {
			return nil, &pongo2.Error{
				Sender:    "filter:timeuntil/timesince",
				OrigError: errors.New("time-parameter is not a time.Time-instance"),
			}
		}
	} else {
		paramtime = time.Now()
	}

	return pongo2.AsValue(humanize.TimeDuration(basetime.Sub(paramtime))), nil
}

func filterIntcomma(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(humanize.Comma(int64(in.Integer()))), nil
}

func filterOrdinal(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(humanize.Ordinal(in.Integer())), nil
}

func filterNaturalday(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	basetime, isTime := in.Interface().(time.Time)
	if !isTime {
		return nil, &pongo2.Error{
			Sender:    "filter:naturalday",
			OrigError: errors.New("naturalday-value is not a time.Time-instance"),
		}
	}

	var referenceTime time.Time
	if !param.IsNil() {
		referenceTime, isTime = param.Interface().(time.Time)
		if !isTime {
			return nil, &pongo2.Error{
				Sender:    "filter:naturalday",
				OrigError: errors.New("naturalday-parameter is not a time.Time-instance"),
			}
		}
	} else {
		referenceTime = time.Now()
	}

	d := referenceTime.Sub(basetime) / time.Hour

	switch {
	case d >= 0 && d < 24:
		// Today
		return pongo2.AsValue("today"), nil
	case d >= 24:
		return pongo2.AsValue("yesterday"), nil
	case d < 0 && d >= -24:
		return pongo2.AsValue("tomorrow"), nil
	}

	// Default behaviour
	return pongo2.ApplyFilter("naturaltime", in, param)
}

func filterIPlus(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(in.Integer() + param.Integer()), nil
}

func filterIMinus(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(in.Integer() - param.Integer()), nil
}

func filterIMultiply(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(in.Integer() * param.Integer()), nil
}

func filterPrintError(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	i := in.Interface()
	switch i.(type) {
	case error:
		return pongo2.AsValue(i.(error).Error()), nil
	}

	return pongo2.AsValue(i), nil
}

func filterJSON(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	js, err := json.Marshal(in.Interface())
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:json",
			OrigError: err,
		}
	}

	return pongo2.AsValue(string(js)), nil
}

func filterSolidLineBreaksBR(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	eachBr := pongoParam(param, 0).Integer()
	line := in.String()

	if len(line) == 0 || eachBr < 2 {
		return in, nil
	}

	breaker := pongoParam(param, 1).String()
	if breaker == "" {
		breaker = "<br />"
	}

	var b bytes.Buffer
	data := []rune(line)
	for i := 0; i < len(data); i++ {
		b.WriteString(string(data[i]))
		if (i+1)%eachBr == 0 && i != len(data)-1 {
			b.WriteString(breaker)
		}
	}

	return pongo2.AsValue(b.String()), nil
}

func filterJoinBr(in *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.CanSlice() {
		return in, nil
	}

	var b bytes.Buffer
	l := in.Len() - 1
	for i := 0; i <= l; i++ {
		b.WriteString(in.Index(i).String())
		if i < l {
			b.WriteString("\n")
		}
	}

	return pongo2.AsValue(b.String()), nil
}

func filterRange(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	from := pongoParam(param, 0).Integer()

	if pongoParam(param, 1).String() == "" {
		// one argument call
		if from < 1 {
			return nil, &pongo2.Error{
				Sender:    "filter:filterRange",
				OrigError: errors.New("range-value is less than 1"),
			}
		}

		out := make([]int, from, from)
		for i := 1; i <= from; i++ {
			out[i-1] = i
		}

		return pongo2.AsValue(out), nil
	}

	to := pongoParam(param, 1).Integer()
	if to < from {
		return nil, &pongo2.Error{
			Sender:    "filter:range0",
			OrigError: errors.New("range0 second parameter is less than first"),
		}
	}

	out := make([]int, to-from+1, to-from+1)
	for i := from; i <= to; i++ {
		out[i-from] = i
	}
	return pongo2.AsValue(out), nil
}

func filterRange0(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	from := pongoParam(param, 0).Integer()

	if pongoParam(param, 1).String() == "" {
		// one argument call
		if from < 0 {
			return nil, &pongo2.Error{
				Sender:    "filter:range0",
				OrigError: errors.New("range0 param is less than 0"),
			}
		}

		out := make([]int, from, from)
		for i := 0; i < from; i++ {
			out[i] = i
		}

		return pongo2.AsValue(out), nil
	}

	to := pongoParam(param, 1).Integer()
	if to < from {
		return nil, &pongo2.Error{
			Sender:    "filter:range0",
			OrigError: errors.New("range0 second parameter is less than first"),
		}
	}

	out := make([]int, to-from, to-from)
	for i := from; i < to; i++ {
		out[i-from] = i
	}

	return pongo2.AsValue(out), nil
}

func pongoParam(param *pongo2.Value, point int) *pongo2.Value {
	if param.Len() > point && point > -1 {
		endings := strings.Split(param.String(), ",")
		for i := range endings {
			if i == point {
				return pongo2.AsValue(endings[i])
			}
		}
	}

	return pongo2.AsValue("")
}
