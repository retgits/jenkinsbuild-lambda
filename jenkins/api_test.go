package jenkins

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	URL       = "http://jenkins.dev:8080/jenkins"
	User      = "admin"
	AuthToken = "12345"
	BuildID   = "MyAwesomeBuild"
)

func TestNewServer(t *testing.T) {
	assert := assert.New(t)

	srv := NewServer(URL, User, AuthToken)
	assert.Equal(srv.URL, URL)
	assert.Equal(srv.User, User)
	assert.Equal(srv.AuthToken, AuthToken)
}

func TestTriggerBuildTCPFailure(t *testing.T) {
	assert := assert.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", fmt.Sprintf("%s/job/%s/build", URL, BuildID), httpmock.NewErrorResponder(fmt.Errorf("Post http://jenkins.dev:8080/jenkins/job/MyAwesomeBuild/build: dial tcp: lookup jenkins.dev: no such host")))

	srv := NewServer(URL, User, AuthToken)
	_, err := srv.TriggerBuild(&BuildRequest{BuildID: BuildID})

	assert.Error(err)
}

func TestTriggerBuildSuccess(t *testing.T) {
	assert := assert.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", fmt.Sprintf("%s/job/%s/build", URL, BuildID), httpmock.NewStringResponder(200, `success...`))

	srv := NewServer(URL, User, AuthToken)
	res, err := srv.TriggerBuild(&BuildRequest{BuildID: BuildID})

	assert.NoError(err)
	assert.Equal(res.HTTPStatusCode, 200)
}
