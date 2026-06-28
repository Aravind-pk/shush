// Package server contains the ConnectRPC service implementation for ShushService.

package server

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	shushv1 "github.com/Aravind-pk/shush/backend/gen/shush/v1"
	"github.com/Aravind-pk/shush/backend/gen/shush/v1/shushv1connect"
	"github.com/Aravind-pk/shush/backend/internal/store"
)

// Server is the concrete ConnectRPC service implementation.
//
// Embedding UnimplementedShushServiceHandler gives us default "Unimplemented"
// stubs for every RPC we haven't written yet, which keeps the file compilable
// as the proto evolves. We only override the methods we actually support.
type Server struct {
	shushv1connect.UnimplementedShushServiceHandler

	store *store.Store

	// demoUserID is used as secrets.created_by until auth supplies a real
	// caller identity. Temporary scaffolding, removed once Register lands.
	demoUserID string
}

// New returns a Server backed by the given store. demoUserID is the user id
// recorded as the author of writes until the auth layer exists.
func New(st *store.Store, demoUserID string) *Server {
	return &Server{store: st, demoUserID: demoUserID}
}

// ListSecrets returns every secret for a given project + environment, fetched
// from PostgreSQL and decrypted via the store's envelope encryption.
//
// Still TODO: auth check on the caller's project access, and an audit entry.
func (s *Server) ListSecrets(ctx context.Context, req *connect.Request[shushv1.ListSecretsRequest]) (*connect.Response[shushv1.ListSecretsResponse], error) {
	projectID := req.Msg.GetProjectId()
	env := req.Msg.GetEnvironment()
	if projectID == "" || env == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("project_id and environment are required"))
	}

	secrets, err := s.store.ListSecrets(ctx, projectID, env)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	out := make([]*shushv1.Secret, 0, len(secrets))
	for _, sec := range secrets {
		out = append(out, &shushv1.Secret{
			ProjectId:   projectID,
			Environment: env,
			Key:         sec.Key,
			Value:       sec.Value,
			Version:     sec.Version,
		})
	}
	return connect.NewResponse(&shushv1.ListSecretsResponse{Secrets: out}), nil
}

// PutSecret encrypts and upserts a secret, returning its id and new version.
//
// Still TODO: auth check + audit entry; created_by is the demo user for now.
func (s *Server) PutSecret(ctx context.Context, req *connect.Request[shushv1.PutSecretRequest]) (*connect.Response[shushv1.PutSecretResponse], error) {
	m := req.Msg
	if m.GetProjectId() == "" || m.GetEnvironment() == "" || m.GetKey() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("project_id, environment, and key are required"))
	}

	id, version, err := s.store.PutSecret(ctx, m.GetProjectId(), m.GetEnvironment(), m.GetKey(), m.GetValue(), s.demoUserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&shushv1.PutSecretResponse{Id: id, Version: version}), nil
}
