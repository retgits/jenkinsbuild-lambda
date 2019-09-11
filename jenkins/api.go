// Package jenkins contains code to connect to Jenkins and trigger builds
package jenkins

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Server contains details to connect to the Jenkins server
type Server struct {
	URL       string
	User      string
	AuthToken string
}

// BuildRequest contains the details to trigger a build
type BuildRequest struct {
	BuildID string
}

// BuildResponse contains the response from Jenkins
type BuildResponse struct {
	HTTPStatusCode    int
	HTTPStatusMessage string
	JenkinsResponse   string
}

// NewServer instantiates a new server struct to connect to Jenkins
func NewServer(URL string, User string, AuthToken string) *Server {
	return &Server{
		URL:       URL,
		User:      User,
		AuthToken: AuthToken,
	}
}

// TriggerBuild sends a request to the Jenkins server to trigger a build and responds accordingly
func (s *Server) TriggerBuild(r *BuildRequest) (*BuildResponse, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/job/%s/build", s.URL, r.BuildID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", s.User, s.AuthToken)))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))

	c := &http.Client{}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n\n", res)
	defer res.Body.Close()

	msg, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n\n", string(msg))

	return &BuildResponse{
		HTTPStatusCode:    res.StatusCode,
		HTTPStatusMessage: res.Status,
		JenkinsResponse:   string(msg),
	}, nil
}

// UnmarshalBuildRequest turns a byte array into a proper BuildRequest struct
func UnmarshalBuildRequest(data []byte) (BuildRequest, error) {
	var r BuildRequest
	err := json.Unmarshal(data, &r)
	return r, err
}
