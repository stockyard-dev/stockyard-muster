package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Attendee struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	EventName string `json:"event_name"`
	TicketType string `json:"ticket_type"`
	Status string `json:"status"`
	CheckedIn int `json:"checked_in"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"muster.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS attendees(id TEXT PRIMARY KEY,name TEXT NOT NULL,email TEXT DEFAULT '',event_name TEXT DEFAULT '',ticket_type TEXT DEFAULT 'general',status TEXT DEFAULT 'registered',checked_in INTEGER DEFAULT 0,notes TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Attendee)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO attendees(id,name,email,event_name,ticket_type,status,checked_in,notes,created_at)VALUES(?,?,?,?,?,?,?,?,?)`,e.ID,e.Name,e.Email,e.EventName,e.TicketType,e.Status,e.CheckedIn,e.Notes,e.CreatedAt);return err}
func(d *DB)Get(id string)*Attendee{var e Attendee;if d.db.QueryRow(`SELECT id,name,email,event_name,ticket_type,status,checked_in,notes,created_at FROM attendees WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.Email,&e.EventName,&e.TicketType,&e.Status,&e.CheckedIn,&e.Notes,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Attendee{rows,_:=d.db.Query(`SELECT id,name,email,event_name,ticket_type,status,checked_in,notes,created_at FROM attendees ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Attendee;for rows.Next(){var e Attendee;rows.Scan(&e.ID,&e.Name,&e.Email,&e.EventName,&e.TicketType,&e.Status,&e.CheckedIn,&e.Notes,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Update(e *Attendee)error{_,err:=d.db.Exec(`UPDATE attendees SET name=?,email=?,event_name=?,ticket_type=?,status=?,checked_in=?,notes=? WHERE id=?`,e.Name,e.Email,e.EventName,e.TicketType,e.Status,e.CheckedIn,e.Notes,e.ID);return err}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM attendees WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM attendees`).Scan(&n);return n}

func(d *DB)Search(q string, filters map[string]string)[]Attendee{
    where:="1=1"
    args:=[]any{}
    if q!=""{
        where+=" AND (name LIKE ? OR email LIKE ?)"
        args=append(args,"%"+q+"%");args=append(args,"%"+q+"%");
    }
    if v,ok:=filters["status"];ok&&v!=""{where+=" AND status=?";args=append(args,v)}
    rows,_:=d.db.Query(`SELECT id,name,email,event_name,ticket_type,status,checked_in,notes,created_at FROM attendees WHERE `+where+` ORDER BY created_at DESC`,args...)
    if rows==nil{return nil};defer rows.Close()
    var o []Attendee;for rows.Next(){var e Attendee;rows.Scan(&e.ID,&e.Name,&e.Email,&e.EventName,&e.TicketType,&e.Status,&e.CheckedIn,&e.Notes,&e.CreatedAt);o=append(o,e)};return o
}

func(d *DB)Stats()map[string]any{
    m:=map[string]any{"total":d.Count()}
    rows,_:=d.db.Query(`SELECT status,COUNT(*) FROM attendees GROUP BY status`)
    if rows!=nil{defer rows.Close();by:=map[string]int{};for rows.Next(){var s string;var c int;rows.Scan(&s,&c);by[s]=c};m["by_status"]=by}
    return m
}
