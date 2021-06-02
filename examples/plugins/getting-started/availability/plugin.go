package availability

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

const SLIPluginID = "getting_started_availability"

var tpl = template.Must(template.New("").Parse(`
sum(rate(http_request_duration_seconds_count{ {{.filter}}job="{{.job}}",code=~"(5..|429)" }[{{"{{.window}}"}}]))
/
sum(rate(http_request_duration_seconds_count{ {{.filter}}job="{{.job}}" }[{{"{{.window}}"}}]))`))

var filterRegex = regexp.MustCompile(`([^=]+="[^=,"]+",)+`)

func SLIPlugin(meta map[string]string, labels map[string]string, options map[string]string) (string, error) {
	// Get job.
	job, ok := options["job"]
	if !ok {
		return "", fmt.Errorf("job options is required")
	}

	// Validate labels.
	err := validateLabels(labels, "owner", "tier")
	if err != nil {
		return "", fmt.Errorf("invalid labels: %w", err)
	}

	// Sanitize filter.
	filter := options["filter"]
	if filter != "" {
		filter = strings.Trim(filter, "{}")
		filter = strings.Trim(filter, ",")
		filter = filter + ","
		match := filterRegex.MatchString(filter)
		if !match {
			return "", fmt.Errorf("invalid prometheus filter: %s", filter)
		}
	}

	// Create query.
	var b bytes.Buffer
	data := map[string]interface{}{
		"job":    job,
		"filter": filter,
	}
	err = tpl.Execute(&b, data)
	if err != nil {
		return "", fmt.Errorf("could not execute template: %w", err)
	}

	return b.String(), nil
}

func validateLabels(labels map[string]string, requiredKeys ...string) error {
	// Validate that the labels have an owner.
	for _, k := range requiredKeys {
		_, ok := labels[k]
		if !ok {
			return fmt.Errorf("%q label is required", k)
		}
	}

	return nil
}

type sliPlugin = func(meta map[string]string, labels map[string]string, options map[string]string) (string, error)
