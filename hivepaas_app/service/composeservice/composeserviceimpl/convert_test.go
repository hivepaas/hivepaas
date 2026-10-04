package composeserviceimpl

import (
	"context"
	"os"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	envPath   = "projects/blog/envs/prod"
	dbPath    = envPath + "/apps/db"
	wpPath    = envPath + "/apps/wordpress"
	projectVl = "projects/blog/volumes/default"
)

func convertReq(compose string) *composeservice.ConvertReq {
	return &composeservice.ConvertReq{
		Compose: compose, ProjectKey: "blog", ProjectName: "Blog", EnvKey: "prod", EnvName: "production",
		NetworkName: "blog_prod_net", RootDomain: "example.com",
	}
}

func fixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return string(data)
}

func convert(t *testing.T, req *composeservice.ConvertReq) *composeservice.ConvertResp {
	t.Helper()
	resp, err := New().Convert(context.Background(), req)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return resp
}

func appOf(t *testing.T, resp *composeservice.ConvertResp, key string) *specmodel.AppDoc {
	t.Helper()
	if !assert.NotNil(t, resp.Bundle) {
		t.FailNow()
	}
	app := resp.Bundle.Envs["blog"]["prod"].Apps[key]
	if !assert.NotNil(t, app, "no app %s", key) {
		t.FailNow()
	}
	return app
}

func envVar(app *specmodel.AppDoc, key string) map[string]any {
	envVars, _ := app.Settings["envVars"].(map[string]any)
	data, _ := envVars["data"].([]any)
	for _, entry := range data {
		if fields, _ := entry.(map[string]any); fields["k"] == key {
			return fields
		}
	}
	return nil
}

func codes(issues []specmodel.Issue) []string {
	var out []string
	for _, issue := range issues {
		out = append(out, issue.Code)
	}
	return out
}

func wordpressReq(t *testing.T) *composeservice.ConvertReq {
	t.Helper()
	req := convertReq(fixture(t, "wordpress.yaml"))
	req.DotEnv = "DB_PASSWORD=s3cret\nDB_ROOT_PASSWORD=r00t\n"
	return req
}

// A WordPress and its database: an app each, their passwords env secrets the
// variables refer to, each volume a directory of the project's, the database's
// port kept in and the site's on a domain.
func TestConvertAWordPressAndItsDatabase(t *testing.T) {
	resp := convert(t, wordpressReq(t))
	assert.Equal(t, "blog", resp.FileName)

	db := appOf(t, resp, "db")
	assert.Equal(t, "${secrets.DB_PASSWORD}", envVar(db, "MYSQL_PASSWORD")["v"])
	assert.Equal(t, "${secrets.DB_ROOT_PASSWORD}", envVar(db, "MYSQL_ROOT_PASSWORD")["v"])
	assert.Equal(t, "wordpress", envVar(db, "MYSQL_USER")["v"])
	secrets, _ := resp.Bundle.Envs["blog"]["prod"].Settings["secrets"].(map[string]any)
	if assert.Contains(t, secrets, "DB_PASSWORD") {
		assert.Equal(t, "s3cret", secrets["DB_PASSWORD"].(map[string]any)["value"])
	}

	mounts := db.Deployment.Storage.Mounts
	assert.Equal(t, specmodel.Mount{Type: mount.TypeVolume, Source: projectVl,
		VolumeOptions: &specmodel.VolumeOptions{Subpath: "db_data"}}, mounts["/var/lib/mysql"])
	assert.Nil(t, db.Deployment.Networks.EndpointSpec, "a database's port is not published")
	assert.Equal(t, docker.HealthcheckModeCmd, db.Deployment.Container.Healthcheck.Mode)
	assert.Equal(t, `mysqladmin ping -h localhost '--password=x y'`, db.Deployment.Container.Healthcheck.Command)
	assert.Equal(t, "any", string(db.Deployment.Container.RestartPolicy.Condition))
	assert.Equal(t, "db", db.Deployment.Container.Hostname)
	assert.Equal(t, []string{"db"}, db.Deployment.Networks.Attachments[0].Aliases)

	wp := appOf(t, resp, "wordpress")
	assert.Equal(t, "db:3306", envVar(wp, "WORDPRESS_DB_HOST")["v"])
	assert.Equal(t, true, envVar(wp, "CRON_LINE")["literal"], "written $${ in compose: no reference of HivePaaS's")
	routing, _ := wp.Settings["routing"].(map[string]any)
	assert.Equal(t, 80, routing["port"])
	domains, _ := routing["domains"].([]any)
	if assert.Len(t, domains, 1) {
		assert.Equal(t, "wordpress-blog.example.com", domains[0].(map[string]any)["domain"])
	}
	assert.Contains(t, codes(resp.Issues[wpPath]), composeservice.CodeStartOrder)
	assert.Contains(t, codes(resp.Issues[envPath]), composeservice.CodeSecretVariable)
}

// Until a required variable has a value nothing is read: the review asks for
// it.
func TestConvertWaitsForARequiredVariable(t *testing.T) {
	req := wordpressReq(t)
	req.DotEnv = "DB_PASSWORD=s3cret\n"
	resp := convert(t, req)

	assert.Nil(t, resp.Bundle)
	for _, v := range resp.Variables {
		if v.Name == "DB_ROOT_PASSWORD" {
			assert.True(t, v.Required)
			assert.False(t, v.Given)
			assert.True(t, v.Secret, "by its name")
		}
	}

	value := "typed"
	req.Variables = map[string]*composeservice.VariableReq{"DB_ROOT_PASSWORD": {Value: &value}}
	assert.NotNil(t, convert(t, req).Bundle, "given in the review")
}

// A variable the review says is not secret is written as its value.
func TestConvertWritesAVariableTheReviewSaysIsNotSecret(t *testing.T) {
	req := wordpressReq(t)
	no := false
	req.Variables = map[string]*composeservice.VariableReq{"DB_PASSWORD": {Secret: &no}}
	db := appOf(t, convert(t, req), "db")
	assert.Equal(t, "s3cret", envVar(db, "MYSQL_PASSWORD")["v"])
}

// The review's choices: a port on the nodes rather than a domain.
func TestConvertTakesThePortTheReviewChose(t *testing.T) {
	req := wordpressReq(t)
	req.Services = map[string]*composeservice.ServiceReq{"wordpress": {Ports: []*composeservice.PortReq{
		{Published: 8080, Target: 80, Protocol: "tcp", As: composeservice.PortAsNode},
	}}}
	wp := appOf(t, convert(t, req), "wordpress")
	assert.Nil(t, wp.Settings["routing"])
	if assert.NotNil(t, wp.Deployment.Networks.EndpointSpec) {
		assert.Equal(t, uint32(8080), wp.Deployment.Networks.EndpointSpec.Ports[0].Published)
	}
}

// Nothing on the server is read: not a file outside the compose file's
// directory, nor one beside it that the request did not give.
func TestConvertReadsNothingOfTheServers(t *testing.T) {
	cases := map[string]string{
		"extends from an absolute path": "services:\n  a:\n    extends: {file: /etc/hostname, service: x}\n",
		"extends leaving the directory": "services:\n  a:\n    extends: {file: ../../x.yaml, service: x}\n",
		"include of an absolute path":   "include: [/etc/compose.yaml]\nservices:\n  a: {image: x}\n",
	}
	for name, compose := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := New().Convert(context.Background(), convertReq(compose))
			assert.Error(t, err)
		})
	}

	_, err := New().Convert(context.Background(), &composeservice.ConvertReq{
		Compose: "services:\n  a: {image: x}\n", Files: map[string][]byte{"../x": nil},
	})
	assert.ErrorIs(t, err, hperrors.ErrComposeFilePath)

	t.Setenv("HIVEPAAS_TEST_SECRET", "from the server")
	resp := convert(t, convertReq("services:\n  a:\n    image: x\n    environment: {V: \"${HIVEPAAS_TEST_SECRET}\"}\n"))
	assert.Equal(t, "", envVar(appOf(t, resp, "a"), "V")["v"], "the process's environment is not the file's")

	resp = convert(t, convertReq("services:\n  a:\n    image: x\n    env_file: [/etc/hostname]\n"))
	assert.Contains(t, codes(resp.Issues[envPath+"/apps/a"]), composeservice.CodeFileMissing)
	assert.Nil(t, appOf(t, resp, "a").Settings["envVars"])
}
