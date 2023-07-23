package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"sigs.k8s.io/yaml"
)

const ImdsIPv4 = "169.254.169.254"
const ImdsIPv6 = "[fd00:ec2::254]"

// A small helper class designed to make calls to the IMDS endpoint with
// a v2 token.
type ImdsSession struct {
	token string
	Url   string
}

// Creates a new ImdsSession object with a valid token
func NewImdsSession(ttl int) (*ImdsSession, error) {
	c := make(chan *ImdsSession)

	// There is not really an easy way to tell in advance if the IPv6
	// metadata service is enabled (even if the host has IPv6, the IPv6
	// IMDS endpoint has to be explictly enabled - thanks AWS =_=) so we
	// just have to connect to both and use whichever doesn't fail.
	for _, ip := range []string{ImdsIPv4, ImdsIPv6} {
		go func(ip string) {
			s := ImdsSession{
				Url: fmt.Sprintf("http://%s/latest/", ip),
			}

			if err := s.RefreshToken(ttl); err != nil {
				warn(err.Error())
			} else {
				c <- &s
			}
		}(ip)
	}

	s := <-c
	info(fmt.Sprintf("IMDS Session created, using %s", s.Url))

	return s, nil
}

// Grabs a new token from the IMDSv2 endpoint with the given TTL
func (s *ImdsSession) RefreshToken(ttl int) error {
	req, err := http.NewRequest(http.MethodPut, s.Url+"api/token", nil)
	if err != nil {
		return fmt.Errorf("Could not create request: %s", err)
	}

	// We want the IMDSv2 token to have a very low TTL as we are only
	// going to use it during this bootstrap process.
	req.Header.Add("X-aws-ec2-metadata-token-ttl-seconds", strconv.Itoa(ttl))

	tokenResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Could not complete request: %s", err)
	}

	rawToken, err := io.ReadAll(tokenResponse.Body)
	if err != nil {
		return fmt.Errorf("Could not read token response: %s", err)
	}

	s.token = string(rawToken)

	return nil
}

// Loads arbitrary metadata from the IMDS endpoint, and returns it as
// a byte array
func (s *ImdsSession) GetMetadata(data string) ([]byte, error) {
	req, err := http.NewRequest("GET", s.Url+data, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not create request: %s", err)
	}

	req.Header.Add("X-aws-ec2-metadata-token", s.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Could not complete request: %s", err)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not read response: %s", err)
	}

	return raw, nil
}

// Loads arbitrary function from the IMDS endpoint, and returns it as a
// string
func (s *ImdsSession) GetString(data string) (string, error) {
	raw, err := s.GetMetadata(data)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// Loads the user data for the instance, assumes that it is YAML and
// unmarshals it as a MetadataInformation object
func (s *ImdsSession) GetUserData() (*MetadataInformation, error) {
	raw, err := s.GetMetadata("user-data")
	if err != nil {
		return nil, err
	}

	data := MetadataInformation{
		Node: Node{
			MaxPods: Limits{
				Set:    true,
				Offset: 3,
			},
		},
	}
	yaml.Unmarshal(raw, &data)

	return &data, nil
}
