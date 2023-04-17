
package main

import (
  "io"
  "fmt"
  "strconv"
  "net/http"

  "sigs.k8s.io/yaml"
)

const ImdsUrl = "http://169.254.169.254/latest/"

// A small helper class designed to make calls to the IMDS endpoint with
// a v2 token.
type ImdsSession struct {
  token string
}

// Creates a new ImdsSession object with a valid token
func NewImdsSession(ttl int) (*ImdsSession, error) {
  s := ImdsSession{}

  if err := s.RefreshToken(ttl); err != nil {
    return nil, fmt.Errorf("Could not get IMDS Token: %s", err)
  }

  return &s, nil
}

// Grabs a new token from the IMDSv2 endpoint with the given TTL
func (s *ImdsSession) RefreshToken(ttl int) error {
  req, err := http.NewRequest(http.MethodPut, ImdsUrl + "api/token", nil)
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
  req, err := http.NewRequest("GET", ImdsUrl + data, nil)
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

  data := MetadataInformation{}
  yaml.Unmarshal(raw, &data)

  return &data, nil
}
