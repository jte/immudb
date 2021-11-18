package server

import (
	"context"
	"fmt"
	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/auth"
	"github.com/codenotary/immudb/pkg/errors"
	"github.com/codenotary/immudb/pkg/server/sessions"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/rs/xid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImmuServer) OpenSession(ctx context.Context, r *schema.OpenSessionRequest) (*schema.OpenSessionResponse, error) {
	if !s.Options.auth {
		return nil, errors.New(ErrAuthDisabled).WithCode(errors.CodProtocolViolation)
	}

	u, err := s.getValidatedUser(r.User, r.Password)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidUsernameOrPassword)
	}
	if u.Username == auth.SysAdminUsername {
		u.IsSysAdmin = true
	}

	if !u.Active {
		return nil, errors.New(ErrUserNotActive)
	}

	databaseID := sysDBIndex
	if r.DatabaseName != SystemDBName {
		databaseID = s.dbList.GetId(r.DatabaseName)
		if databaseID < 0 {
			return nil, errors.New(fmt.Sprintf("'%s' does not exist", r.DatabaseName)).WithCode(errors.CodInvalidDatabaseName)
		}
	}

	if (!u.IsSysAdmin) &&
		(!u.HasPermission(r.DatabaseName, auth.PermissionAdmin)) &&
		(!u.HasPermission(r.DatabaseName, auth.PermissionR)) &&
		(!u.HasPermission(r.DatabaseName, auth.PermissionRW)) {

		return nil, status.Errorf(codes.PermissionDenied, "Logged in user does not have permission on this database")
	}

	newSession := sessions.NewSession(u, databaseID)

	sessionID := xid.New().String()
	if s.SessManager.SessionPresent(sessionID) {
		return nil, ErrSessionAlreadyPresent
	}

	s.SessManager.AddSession(sessionID, newSession)

	return &schema.OpenSessionResponse{
		SessionID:  sessionID,
		ServerUUID: s.UUID.String(),
	}, nil
}

func (s *ImmuServer) CloseSession(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	if !s.Options.auth {
		return nil, errors.New(ErrAuthDisabled).WithCode(errors.CodProtocolViolation)
	}
	sessionID, err := sessions.GetSessionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	s.SessManager.RemoveSession(sessionID)
	s.Logger.Debugf("closing session %s", sessionID)
	return new(empty.Empty), nil
}
