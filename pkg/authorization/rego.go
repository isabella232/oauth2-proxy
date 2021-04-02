package authorization

import (
	"net/http"

	sessionsapi "github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/open-policy-agent/opa/rego"
)

type authInput struct {
	request *http.Request
	session *sessionsapi.SessionState
}

func authorize(req *http.Request, session *sessionsapi.SessionState) (bool, error) {
	r := rego.New(
		rego.Query("auth = data.oauth2proxy.allow"),
		rego.Module("oauth2proxy.rego", `
package oauth2proxy

default allow = false

allow {
  endswith(input.session.email, "@bar.com")
}
    `),
	)

	query, err := r.PrepareForEval(req.Context())
	if err != nil {
		return false, err
	}

	input := rego.EvalInput(authInput{
		request: req,
		session: session,
	})

	result, err := query.Eval(req.Context(), input)
	if err != nil {
		return false, err
	}
	if len(result) == 0 {
		return false, nil
	}

	if auth, ok := result[0].Bindings["auth"].(bool); ok {
		return auth, nil
	}

	return false, nil
}
