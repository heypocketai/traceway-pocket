package services

import (
	"database/sql"
	"encoding/base32"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/repositories"
)

// dbSessionStore is a gorilla/sessions.Store backed by the main transactional
// DB. The cookie holds only a signed session ID; the session data (which can
// exceed the 4096-byte cookie limit for large OIDC ID tokens) lives in the
// oauth_sessions table.
type dbSessionStore struct {
	codecs  []securecookie.Codec
	options *sessions.Options
}

func newDBSessionStore(secret string, options *sessions.Options) *dbSessionStore {
	codecs := securecookie.CodecsFromPairs([]byte(secret))
	for _, c := range codecs {
		if sc, ok := c.(*securecookie.SecureCookie); ok {
			sc.MaxLength(0)
		}
	}
	return &dbSessionStore{codecs: codecs, options: options}
}

func (s *dbSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *dbSessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.options
	session.Options = &opts
	session.IsNew = true

	cookie, err := r.Cookie(name)
	if err != nil {
		return session, nil
	}

	var id string
	if err := securecookie.DecodeMulti(name, cookie.Value, &id, s.codecs...); err != nil {
		return session, nil
	}

	data, err := db.ExecuteTransaction(func(tx *sql.Tx) ([]byte, error) {
		return repositories.OAuthSessionRepository.Get(tx, id)
	})
	if err != nil || data == nil {
		return session, nil
	}

	if err := securecookie.DecodeMulti(name, string(data), &session.Values, s.codecs...); err != nil {
		return session, nil
	}
	session.ID = id
	session.IsNew = false
	return session, nil
}

func (s *dbSessionStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options != nil && session.Options.MaxAge < 0 {
		if session.ID != "" {
			if _, err := db.ExecuteTransaction(func(tx *sql.Tx) (struct{}, error) {
				return struct{}{}, repositories.OAuthSessionRepository.Delete(tx, session.ID)
			}); err != nil {
				return err
			}
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)),
			"=",
		)
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, s.codecs...)
	if err != nil {
		return err
	}

	expires := time.Now().UTC().Add(time.Duration(session.Options.MaxAge) * time.Second)
	if _, err := db.ExecuteTransaction(func(tx *sql.Tx) (struct{}, error) {
		return struct{}{}, repositories.OAuthSessionRepository.Save(tx, session.ID, []byte(encoded), expires)
	}); err != nil {
		return err
	}

	cookieValue, err := securecookie.EncodeMulti(session.Name(), session.ID, s.codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), cookieValue, session.Options))
	return nil
}
