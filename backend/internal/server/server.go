// Package server contains the ConnectRPC service implementation for ShushService.

package server

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	shushv1 "github.com/Aravind-pk/shush/backend/gen/shush/v1"
	"github.com/Aravind-pk/shush/backend/gen/shush/v1/shushv1connect"
	"github.com/Aravind-pk/shush/backend/internal/auth"
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
}

// New returns a Server backed by the given store. Identity comes from the
// Clerk auth interceptor on each request, so no user is injected here.
func New(st *store.Store) *Server {
	return &Server{store: st}
}

// callerUserID resolves the authenticated Clerk user (set by the interceptor)
// to our internal user UUID, creating the user row on first sight.
func (s *Server) callerUserID(ctx context.Context) (string, error) {
	clerkUserID, ok := auth.UserID(ctx)
	if !ok {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("not authenticated"))
	}
	userID, err := s.store.EnsureUser(ctx, clerkUserID)
	if err != nil {
		return "", connect.NewError(connect.CodeInternal, err)
	}
	return userID, nil
}

// CreateProject creates a project owned by the authenticated user.
func (s *Server) CreateProject(ctx context.Context, req *connect.Request[shushv1.CreateProjectRequest]) (*connect.Response[shushv1.CreateProjectResponse], error) {
	userID, err := s.callerUserID(ctx)
	if err != nil {
		return nil, err
	}
	if req.Msg.GetName() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name is required"))
	}

	p, err := s.store.CreateProject(ctx, userID, req.Msg.GetName())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&shushv1.CreateProjectResponse{Project: toProtoProject(p)}), nil
}

// ListProjects returns the authenticated user's projects.
func (s *Server) ListProjects(ctx context.Context, _ *connect.Request[shushv1.ListProjectsRequest]) (*connect.Response[shushv1.ListProjectsResponse], error) {
	userID, err := s.callerUserID(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := s.store.ListProjects(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := make([]*shushv1.Project, 0, len(projects))
	for _, p := range projects {
		out = append(out, toProtoProject(p))
	}
	return connect.NewResponse(&shushv1.ListProjectsResponse{Projects: out}), nil
}

// ListSecrets returns metadata (key + version) for a project + environment.
// Plaintext values are never included — clients call GetSecret to reveal one
// value at a time.
//
// Still TODO: verify the project belongs to the caller.
func (s *Server) ListSecrets(ctx context.Context, req *connect.Request[shushv1.ListSecretsRequest]) (*connect.Response[shushv1.ListSecretsResponse], error) {
	if _, err := s.callerUserID(ctx); err != nil {
		return nil, err
	}
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
			Version:     sec.Version,
		})
	}
	return connect.NewResponse(&shushv1.ListSecretsResponse{Secrets: out}), nil
}

// GetSecret decrypts and returns one secret's value. This is the only RPC that
// exposes plaintext, so reveals are explicit and one-at-a-time rather than the
// whole project being decrypted on list.
//
// Still TODO: verify the project belongs to the caller, and write a `read`
// audit entry (the point at which plaintext is disclosed).
func (s *Server) GetSecret(ctx context.Context, req *connect.Request[shushv1.GetSecretRequest]) (*connect.Response[shushv1.GetSecretResponse], error) {
	if _, err := s.callerUserID(ctx); err != nil {
		return nil, err
	}
	m := req.Msg
	if m.GetProjectId() == "" || m.GetEnvironment() == "" || m.GetKey() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("project_id, environment, and key are required"))
	}

	sec, err := s.store.GetSecret(ctx, m.GetProjectId(), m.GetEnvironment(), m.GetKey())
	if errors.Is(err, store.ErrNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("secret not found"))
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&shushv1.GetSecretResponse{
		Secret: &shushv1.Secret{
			ProjectId:   m.GetProjectId(),
			Environment: m.GetEnvironment(),
			Key:         sec.Key,
			Value:       sec.Value,
			Version:     sec.Version,
		},
	}), nil
}

// PutSecret encrypts and upserts a secret, attributed to the caller.
//
// Still TODO: verify the project belongs to the caller, and write an audit entry.
func (s *Server) PutSecret(ctx context.Context, req *connect.Request[shushv1.PutSecretRequest]) (*connect.Response[shushv1.PutSecretResponse], error) {
	userID, err := s.callerUserID(ctx)
	if err != nil {
		return nil, err
	}
	m := req.Msg
	if m.GetProjectId() == "" || m.GetEnvironment() == "" || m.GetKey() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("project_id, environment, and key are required"))
	}

	id, version, err := s.store.PutSecret(ctx, m.GetProjectId(), m.GetEnvironment(), m.GetKey(), m.GetValue(), userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&shushv1.PutSecretResponse{Id: id, Version: version}), nil
}

func toProtoProject(p store.Project) *shushv1.Project {
	return &shushv1.Project{
		Id:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
	}
}
