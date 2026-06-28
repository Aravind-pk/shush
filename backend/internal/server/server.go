// Package server contains the ConnectRPC service implementation for ShushService.

package server

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	shushv1 "github.com/Aravind-pk/shush/backend/gen/shush/v1"
	"github.com/Aravind-pk/shush/backend/gen/shush/v1/shushv1connect"
)

// Server is the concrete ConnectRPC service implementation.
//
// Embedding UnimplementedShushServiceHandler gives us default "Unimplemented"
// stubs for every RPC we haven't written yet, which keeps the file compilable
// as the proto evolves. We only override the methods we actually support.
type Server struct {
	shushv1connect.UnimplementedShushServiceHandler

	// secrets is a temporary hardcoded store keyed by "<project>/<env>".
	// Replaced with the PostgreSQL store + crypto layer in a later slice.
	secrets map[string]map[string]string
}

// New returns a Server seeded with a couple of demo secrets so the CLI
// has something to fetch.
func New() *Server {
	return &Server{
		secrets: map[string]map[string]string{
			"demo/dev": {
				"DATABASE_URL": "postgres://dev:devpass@localhost:5432/demo",
				"API_KEY":      "dev-sk-1234567890",
				"FEATURE_X":    "true",
			},
			"demo/prod": {
				"DATABASE_URL": "postgres://prod:prodpass@db.internal:5432/demo",
				"API_KEY":      "prod-sk-deadbeef",
				"FEATURE_X":    "false",
			},
		},
	}
}

// ListSecrets returns every secret for a given project + environment.
//
// In the real implementation this will:
//   - check the caller's auth context for project access
//   - fetch encrypted rows from PostgreSQL
//   - decrypt each value via envelope encryption
//   - record an audit log entry
//
// For now: a flat lookup against the hardcoded map.
// Connect wraps the request and response messages: the proto message lives on
// req.Msg, and we hand back a *connect.Response built around the proto reply.
func (s *Server) ListSecrets(ctx context.Context, req *connect.Request[shushv1.ListSecretsRequest]) (*connect.Response[shushv1.ListSecretsResponse], error) {
	key := fmt.Sprintf("%s/%s", req.Msg.GetProjectId(), req.Msg.GetEnvironment())
	bucket, ok := s.secrets[key]
	if !ok {
		// Empty list, not an error — "no secrets here" is a valid state.
		return connect.NewResponse(&shushv1.ListSecretsResponse{}), nil
	}

	out := make([]*shushv1.Secret, 0, len(bucket))
	for k, v := range bucket {
		out = append(out, &shushv1.Secret{
			ProjectId:   req.Msg.GetProjectId(),
			Environment: req.Msg.GetEnvironment(),
			Key:         k,
			Value:       v,
			Version:     1,
		})
	}
	return connect.NewResponse(&shushv1.ListSecretsResponse{Secrets: out}), nil
}
