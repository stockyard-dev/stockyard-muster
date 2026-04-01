package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ conn *sql.DB }

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	conn, err := sql.Open("sqlite", filepath.Join(dataDir, "muster.db"))
	if err != nil {
		return nil, err
	}
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA busy_timeout=5000")
	conn.SetMaxOpenConns(4)
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error { return db.conn.Close() }

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS members (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT DEFAULT '',
    phone TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS schedules (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    notes TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_sched_dates ON schedules(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_sched_member ON schedules(member_id);

CREATE TABLE IF NOT EXISTS incidents (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    severity TEXT DEFAULT 'medium',
    status TEXT DEFAULT 'open',
    description TEXT DEFAULT '',
    responder_id TEXT DEFAULT '',
    started_at TEXT DEFAULT (datetime('now')),
    resolved_at TEXT DEFAULT '',
    resolution TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_inc_status ON incidents(status);
`)
	return err
}

// --- Members ---

type Member struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
}

func (db *DB) CreateMember(name, email, phone string) (*Member, error) {
	id := "mem_" + genID(6)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO members (id,name,email,phone,created_at) VALUES (?,?,?,?,?)", id, name, email, phone, now)
	if err != nil {
		return nil, err
	}
	return &Member{ID: id, Name: name, Email: email, Phone: phone, CreatedAt: now}, nil
}

func (db *DB) ListMembers() ([]Member, error) {
	rows, err := db.conn.Query("SELECT id,name,email,phone,created_at FROM members ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Member
	for rows.Next() {
		var m Member
		rows.Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.CreatedAt)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (db *DB) GetMember(id string) (*Member, error) {
	var m Member
	err := db.conn.QueryRow("SELECT id,name,email,phone,created_at FROM members WHERE id=?", id).
		Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.CreatedAt)
	return &m, err
}

func (db *DB) DeleteMember(id string) error {
	_, err := db.conn.Exec("DELETE FROM members WHERE id=?", id)
	return err
}

// --- Schedules ---

type Schedule struct {
	ID         string `json:"id"`
	MemberID   string `json:"member_id"`
	MemberName string `json:"member_name,omitempty"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Notes      string `json:"notes"`
	CreatedAt  string `json:"created_at"`
}

func (db *DB) CreateSchedule(memberID, startDate, endDate, notes string) (*Schedule, error) {
	id := "sch_" + genID(6)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO schedules (id,member_id,start_date,end_date,notes,created_at) VALUES (?,?,?,?,?,?)",
		id, memberID, startDate, endDate, notes, now)
	if err != nil {
		return nil, err
	}
	return &Schedule{ID: id, MemberID: memberID, StartDate: startDate, EndDate: endDate, Notes: notes, CreatedAt: now}, nil
}

func (db *DB) ListSchedules() ([]Schedule, error) {
	rows, err := db.conn.Query(`SELECT s.id, s.member_id, COALESCE(m.name,''), s.start_date, s.end_date, s.notes, s.created_at
		FROM schedules s LEFT JOIN members m ON m.id=s.member_id ORDER BY s.start_date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Schedule
	for rows.Next() {
		var s Schedule
		rows.Scan(&s.ID, &s.MemberID, &s.MemberName, &s.StartDate, &s.EndDate, &s.Notes, &s.CreatedAt)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (db *DB) CurrentOnCall() (*Schedule, error) {
	today := time.Now().Format("2006-01-02")
	var s Schedule
	err := db.conn.QueryRow(`SELECT s.id, s.member_id, COALESCE(m.name,''), s.start_date, s.end_date, s.notes, s.created_at
		FROM schedules s LEFT JOIN members m ON m.id=s.member_id
		WHERE s.start_date<=? AND s.end_date>=? LIMIT 1`, today, today).
		Scan(&s.ID, &s.MemberID, &s.MemberName, &s.StartDate, &s.EndDate, &s.Notes, &s.CreatedAt)
	return &s, err
}

func (db *DB) DeleteSchedule(id string) error {
	_, err := db.conn.Exec("DELETE FROM schedules WHERE id=?", id)
	return err
}

// --- Incidents ---

type Incident struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Status      string `json:"status"`
	Description string `json:"description"`
	ResponderID string `json:"responder_id"`
	StartedAt   string `json:"started_at"`
	ResolvedAt  string `json:"resolved_at,omitempty"`
	Resolution  string `json:"resolution,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func (db *DB) CreateIncident(title, severity, description, responderID string) (*Incident, error) {
	id := "inc_" + genID(8)
	now := time.Now().UTC().Format(time.RFC3339)
	if severity == "" {
		severity = "medium"
	}
	_, err := db.conn.Exec("INSERT INTO incidents (id,title,severity,description,responder_id,started_at,created_at) VALUES (?,?,?,?,?,?,?)",
		id, title, severity, description, responderID, now, now)
	if err != nil {
		return nil, err
	}
	return &Incident{ID: id, Title: title, Severity: severity, Status: "open",
		Description: description, ResponderID: responderID, StartedAt: now, CreatedAt: now}, nil
}

func (db *DB) ListIncidents(statusFilter string) ([]Incident, error) {
	var rows *sql.Rows
	var err error
	if statusFilter != "" {
		rows, err = db.conn.Query("SELECT id,title,severity,status,description,responder_id,started_at,resolved_at,resolution,created_at FROM incidents WHERE status=? ORDER BY created_at DESC", statusFilter)
	} else {
		rows, err = db.conn.Query("SELECT id,title,severity,status,description,responder_id,started_at,resolved_at,resolution,created_at FROM incidents ORDER BY created_at DESC LIMIT 50")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Incident
	for rows.Next() {
		var i Incident
		rows.Scan(&i.ID, &i.Title, &i.Severity, &i.Status, &i.Description, &i.ResponderID, &i.StartedAt, &i.ResolvedAt, &i.Resolution, &i.CreatedAt)
		out = append(out, i)
	}
	return out, rows.Err()
}

func (db *DB) GetIncident(id string) (*Incident, error) {
	var i Incident
	err := db.conn.QueryRow("SELECT id,title,severity,status,description,responder_id,started_at,resolved_at,resolution,created_at FROM incidents WHERE id=?", id).
		Scan(&i.ID, &i.Title, &i.Severity, &i.Status, &i.Description, &i.ResponderID, &i.StartedAt, &i.ResolvedAt, &i.Resolution, &i.CreatedAt)
	return &i, err
}

func (db *DB) ResolveIncident(id, resolution string) {
	now := time.Now().UTC().Format(time.RFC3339)
	db.conn.Exec("UPDATE incidents SET status='resolved', resolved_at=?, resolution=? WHERE id=?", now, resolution, id)
}

func (db *DB) DeleteIncident(id string) error {
	_, err := db.conn.Exec("DELETE FROM incidents WHERE id=?", id)
	return err
}

// --- Stats ---

func (db *DB) Stats() map[string]any {
	var members, schedules, openInc, resolvedInc int
	db.conn.QueryRow("SELECT COUNT(*) FROM members").Scan(&members)
	db.conn.QueryRow("SELECT COUNT(*) FROM schedules").Scan(&schedules)
	db.conn.QueryRow("SELECT COUNT(*) FROM incidents WHERE status='open'").Scan(&openInc)
	db.conn.QueryRow("SELECT COUNT(*) FROM incidents WHERE status='resolved'").Scan(&resolvedInc)
	return map[string]any{"members": members, "schedules": schedules, "open_incidents": openInc, "resolved_incidents": resolvedInc}
}

func genID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
