package schedjobexecserviceimpl

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

func envMap(env []string) map[string]string {
	m := map[string]string{}
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	return m
}

func TestSequenceEnvTellsAStepHowTheOnesBeforeItWent(t *testing.T) {
	start := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	exitCode := 0
	step := &schedjobexecservice.SequenceStep{
		Step: 3, Steps: 4, OutputFile: "/tmp/hivepaas-output-t1-3",
		Earlier: []*entity.SchedJobSeqStepResult{
			{Job: entity.ObjectID{ID: "job-a"}, Name: "migrate", Status: base.SchedJobSeqStepDone, ExitCode: &exitCode,
				StartedAt: start, EndedAt: start.Add(2 * time.Second), Outputs: map[string]string{"version": "42"}},
			{Job: entity.ObjectID{ID: "job-b"}, Status: base.SchedJobSeqStepFailed, Error: "boom",
				Outputs: map[string]string{"count": "7", "file_name": "a.sql"}},
		},
	}

	env := envMap(sequenceEnv(step))

	assert.Equal(t, "3", env["HIVEPAAS_SEQ_STEP"])
	assert.Equal(t, "4", env["HIVEPAAS_SEQ_STEPS"])
	assert.Equal(t, "failed", env["HIVEPAAS_SEQ_PREV_STATUS"])
	assert.Equal(t, "/tmp/hivepaas-output-t1-3", env["HIVEPAAS_OUTPUT"])
	assert.Equal(t, "7", env["HIVEPAAS_SEQ_OUTPUT_COUNT"], "the previous step's outputs, one variable each")
	assert.Equal(t, "a.sql", env["HIVEPAAS_SEQ_OUTPUT_FILE_NAME"])
	assert.NotContains(t, env, "HIVEPAAS_SEQ_OUTPUT_VERSION", "only the previous step's")
	assert.JSONEq(t, `[{"job":"job-a","name":"migrate","status":"done","exitCode":0,"durationMs":2000},
		{"job":"job-b","status":"failed","error":"boom"}]`, env["HIVEPAAS_SEQ_RESULTS"])
	assert.JSONEq(t, `[{"job":"job-a","name":"migrate","outputs":{"version":"42"}},
		{"job":"job-b","outputs":{"count":"7","file_name":"a.sql"}}]`, env["HIVEPAAS_SEQ_OUTPUTS"])
}

func TestSequenceEnvOfTheFirstStep(t *testing.T) {
	env := envMap(sequenceEnv(&schedjobexecservice.SequenceStep{Step: 1, Steps: 2, OutputFile: "/tmp/x"}))

	assert.Equal(t, "", env["HIVEPAAS_SEQ_PREV_STATUS"])
	assert.Equal(t, "[]", env["HIVEPAAS_SEQ_RESULTS"])
	assert.Equal(t, "[]", env["HIVEPAAS_SEQ_OUTPUTS"])
	assert.Nil(t, sequenceEnv(nil), "outside a sequence")
}

func TestParseOutputs(t *testing.T) {
	outputs, warnings := parseOutputs([]byte("VERSION=42\nfile_name=a=b.sql\r\n\n# a comment\n" +
		"bad key=1\n1st=x\nnoequals\nEMPTY=\n"))

	assert.Equal(t, map[string]string{"VERSION": "42", "file_name": "a=b.sql", "EMPTY": ""}, outputs)
	assert.Len(t, warnings, 3, "the bad key, the key starting with a digit, the line with no =")
}

func TestParseOutputsKeepsTheFirst64KB(t *testing.T) {
	line := "K=" + strings.Repeat("v", 1000) + "\n"
	data := strings.Repeat(line, 70) // about 70 KB

	outputs, warnings := parseOutputs([]byte(data))

	assert.Len(t, outputs, 1, "one key, written again and again")
	if assert.NotEmpty(t, warnings) {
		assert.Contains(t, warnings[len(warnings)-1], "64 KB")
	}
}

func TestOutputFilePath(t *testing.T) {
	assert.Equal(t, "/tmp/hivepaas-output-task1-2", schedjobexecservice.OutputFilePath("task1", 2))
}
