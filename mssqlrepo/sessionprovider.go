package mssqlrepo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/satori/go.uuid"

	"time"

	"log"

	"github.com/syllabix/juno"
)

//NewSessionProvider is a factory constructor used to create a useful instance of SessionProvider
func NewSessionProvider(db *sql.DB, cookieProvider juno.CookieProvider, duration ...time.Duration) *SessionProvider {

	var dur time.Duration
	if len(duration) < 1 {
		dur = time.Minute * 30
	} else {
		dur = duration[0]
	}

	g, err := db.Prepare(getsession)
	if err != nil {
		log.Fatal("Session Provider failed to instantiate:", err)
	}

	i, err := db.Prepare(insertsession)
	if err != nil {
		log.Fatal("Session Provider failed to instantiate:", err)
	}

	return &SessionProvider{
		db:         db,
		cookie:     cookieProvider,
		duration:   dur,
		getStmt:    g,
		insertStmt: i,
	}
}

//SessionProvider is an implementation of juno.SessionProvider using mssql as it's backing store
type SessionProvider struct {
	db         *sql.DB
	cookie     juno.CookieProvider
	duration   time.Duration
	getStmt    *sql.Stmt
	insertStmt *sql.Stmt
}

const getsession = `
    SELECT cast(GUID as char(36)), Expiration, ContentsJSON FROM dbo.UserSessions
    WHERE GUID = ?
		AND Expiration > SYSDATETIMEOFFSET()
`

//GetSession tries to retrieve an existing session, if it failes, it creates one. If session creation failed, it returns an error
func (sp *SessionProvider) GetSession(req *http.Request) (juno.Session, error) {
	baseSession, err := sp.cookie.Read(req)
	if err != nil {
		session := juno.NewStdSession(sp.duration)
		err = sp.SetSession(session)
		return session, err
	}

	var (
		guid         string
		expiration   time.Time
		contentsJSON sql.NullString
	)

	qID, err := uuid.FromString(baseSession.SessionID())
	if err != nil {
		return nil, fmt.Errorf("Invalid GUID: %v", baseSession.SessionID())
	}

	err = sp.getStmt.QueryRow(qID).Scan(&guid, &expiration, &contentsJSON)

	if err == sql.ErrNoRows {
		session := juno.NewStdSession(sp.duration)
		err = sp.SetSession(session)
		return session, err
	} else if err != nil {
		return nil, err
	}

	sessionID, err := uuid.FromString(guid)

	if err != nil {
		return nil, fmt.Errorf("Invalid GUID: %v", guid)
	}

	session := new(juno.StdSession)
	session.ID = sessionID
	session.Expiration = expiration

	if contentsJSON.Valid {
		var store map[string]interface{}
		err := json.Unmarshal([]byte(contentsJSON.String), &store)
		if err != nil {
			return session, err
		}
		session.ReplaceStore(store)
	}

	return session, nil
}

const insertsession = `INSERT INTO dbo.UserSessions (GUID, Expiration) VALUES (?, ?)`

//SetSession creates a new session and stores it in the database
func (sp *SessionProvider) SetSession(s juno.Session) error {
	exp := time.Now().Add(sp.duration)
	_, err := sp.insertStmt.Exec(s.SessionID(), exp)
	return err
}

const updatesessionDirty = `
    UPDATE dbo.UserSessions
    SET Expiration = ?, ContentsJSON = ?
    WHERE GUID = ?`

const updatesessionClean = `
    UPDATE dbo.UserSessions
    SET Expiration = ?
    WHERE GUID = ?`

//UpdateSession updates the session expiration and contents if dirty
func (sp *SessionProvider) UpdateSession(s juno.Session) error {
	exp := time.Now().Add(sp.duration)
	if s.StoreDirty() {
		contentsJSON, err := json.Marshal(s.Store())
		if err != nil {
			return err
		}

		_, err = sp.db.Exec(updatesessionDirty, exp, string(contentsJSON), s.SessionID())
		return err
	} else {
		_, err := sp.db.Exec(updatesessionClean, exp, s.SessionID())
		return err
	}
}

const deletesession = `DELETE FROM dbo.UserSessions WHERE GUID = ?`

//EndSession terminates a session be removing it from the database and invlaidating the cookie
func (sp *SessionProvider) EndSession(w http.ResponseWriter, s juno.Session) error {
	sp.cookie.Invalidate(w)
	_, err := sp.db.Exec(deletesession, s.SessionID())
	return err
}

//SaveSession sets the session id on the cookie
func (sp *SessionProvider) WriteCookie(w http.ResponseWriter, s juno.Session) error {
	return sp.cookie.Set(w, s)
}
