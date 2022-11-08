package reverseproxy

import (
	"net/http"
)

type Transport struct {
	base http.RoundTripper
}

func (t Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	res, err := t.base.RoundTrip(r)

	// Error or (no error but status is a failure)
	if err != nil || (err == nil && t.ResponseCountsAsFailure(res.StatusCode)) {
		u, ok := GetVar(r.Context(), "upstream").(*Upstream)
		if ok {
			u.CountFailure()
		}
	}

	return res, err
}

// ResponseCountsAsFailure decides if an HTTP status code
// counts as a failure towards an upstream becoming ineligible
func (t Transport) ResponseCountsAsFailure(status int) bool {
	switch status {
	case http.StatusBadGateway:
	case http.StatusServiceUnavailable:
	case http.StatusGatewayTimeout:
		return true
	}
	return false
}
