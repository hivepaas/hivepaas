package notificationserviceimpl

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

// hostile is what anybody's text may hold: quotes, a backslash, new lines and
// tabs, a control character, and characters JSON or HTML give a meaning to.
const hostile = "Shop \"eu\" \\ <b>&</b>\n\tline two \x1b[31m é 中"

// withHostileText fills every string of a template's data - its own, those of
// the struct it embeds - with hostile text, kept apart by the field's name.
func withHostileText(data any) any {
	v := reflect.New(reflect.TypeOf(data)).Elem()
	v.Set(reflect.ValueOf(data))
	var fill func(v reflect.Value)
	fill = func(v reflect.Value) {
		for i := range v.NumField() {
			f := v.Field(i)
			switch {
			case f.Kind() == reflect.Struct && v.Type().Field(i).Anonymous:
				fill(f)
			case f.Kind() == reflect.String && f.CanSet():
				f.SetString(v.Type().Field(i).Name + ": " + hostile)
			}
		}
	}
	fill(v)
	return v.Interface()
}

// Every template of a JSON body writes a body that is JSON, whatever the text
// it is given; and the text reaches the reader as it was.
func TestJSONTemplatesWriteJSONWhateverTheText(t *testing.T) {
	started := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	kinds := map[string][]any{
		"app_deployment_notification": {
			notificationservice.TemplateDataAppDeployment{Method: "repo", StartedAt: started},
			notificationservice.TemplateDataAppDeployment{Method: "image", Succeeded: true, StartedAt: started},
		},
		"app_clone_notification": {
			notificationservice.TemplateDataAppClone{StartedAt: started},
			notificationservice.TemplateDataAppClone{Succeeded: true, StartedAt: started},
		},
		"sched_task_notification":    {notificationservice.TemplateDataSchedTask{StartedAt: started, Retries: 2}},
		"healthcheck_notification":   {notificationservice.TemplateDataHealthcheck{StartedAt: started}},
		"ssl_expiring_notification":  {notificationservice.TemplateDataSSLExpiring{}},
		"ssl_renewal_notification":   {notificationservice.TemplateDataSSLRenewal{}},
		"system_update_notification": {notificationservice.TemplateDataSystemUpdate{StartedAt: started}},
	}
	for _, channel := range []string{"slack", "discord", "lark"} {
		for kind, variants := range kinds {
			file := channel + "/" + kind + ".tpl"
			tpl, err := parseJSONTemplate(file)
			if err != nil {
				t.Fatalf("parsing %s: %v", file, err)
			}
			for _, data := range variants {
				data = withHostileText(data)
				// The data's methods - StartedAtFormatted - are on the value.
				var buf bytes.Buffer
				if err := tpl.Execute(&buf, data); err != nil {
					t.Fatalf("rendering %s: %v", file, err)
				}
				var body any
				if err := json.Unmarshal(buf.Bytes(), &body); err != nil {
					t.Fatalf("%s is not JSON: %v\n%s", file, err, buf.String())
				}
				// The project's name, when the template shows one, as it was given.
				if project := reflect.ValueOf(data).FieldByName("ProjectName"); project.IsValid() {
					decoded, _ := json.Marshal(body)
					want, _ := json.Marshal(project.String())
					if !strings.Contains(string(decoded), strings.Trim(string(want), `"`)) {
						t.Errorf("%s lost the project's name %q:\n%s", file, project.String(), decoded)
					}
				}
			}
		}
	}
}
