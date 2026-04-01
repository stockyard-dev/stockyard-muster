package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/stockyard-dev/stockyard-muster/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	port   int
	limits Limits
}

func New(db *store.DB, port int, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), port: port, limits: limits}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /api/members", s.handleCreateMember)
	s.mux.HandleFunc("GET /api/members", s.handleListMembers)
	s.mux.HandleFunc("DELETE /api/members/{id}", s.handleDeleteMember)

	s.mux.HandleFunc("POST /api/schedules", s.handleCreateSchedule)
	s.mux.HandleFunc("GET /api/schedules", s.handleListSchedules)
	s.mux.HandleFunc("GET /api/oncall", s.handleCurrentOnCall)
	s.mux.HandleFunc("DELETE /api/schedules/{id}", s.handleDeleteSchedule)

	s.mux.HandleFunc("POST /api/incidents", s.handleCreateIncident)
	s.mux.HandleFunc("GET /api/incidents", s.handleListIncidents)
	s.mux.HandleFunc("GET /api/incidents/{id}", s.handleGetIncident)
	s.mux.HandleFunc("POST /api/incidents/{id}/resolve", s.handleResolveIncident)
	s.mux.HandleFunc("DELETE /api/incidents/{id}", s.handleDeleteIncident)

	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /ui", s.handleUI)
	s.mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"product": "stockyard-muster", "version": "0.1.0"})
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("[muster] listening on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) handleCreateMember(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "name is required"})
		return
	}
	m, err := s.db.CreateMember(req.Name, req.Email, req.Phone)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"member": m})
}

func (s *Server) handleListMembers(w http.ResponseWriter, r *http.Request) {
	members, _ := s.db.ListMembers()
	if members == nil {
		members = []store.Member{}
	}
	writeJSON(w, 200, map[string]any{"members": members, "count": len(members)})
}

func (s *Server) handleDeleteMember(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteMember(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Server) handleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MemberID  string `json:"member_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Notes     string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MemberID == "" || req.StartDate == "" || req.EndDate == "" {
		writeJSON(w, 400, map[string]string{"error": "member_id, start_date, end_date required"})
		return
	}
	sch, err := s.db.CreateSchedule(req.MemberID, req.StartDate, req.EndDate, req.Notes)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"schedule": sch})
}

func (s *Server) handleListSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, _ := s.db.ListSchedules()
	if schedules == nil {
		schedules = []store.Schedule{}
	}
	writeJSON(w, 200, map[string]any{"schedules": schedules, "count": len(schedules)})
}

func (s *Server) handleCurrentOnCall(w http.ResponseWriter, r *http.Request) {
	sch, err := s.db.CurrentOnCall()
	if err != nil {
		writeJSON(w, 200, map[string]any{"on_call": nil, "message": "no one currently on call"})
		return
	}
	writeJSON(w, 200, map[string]any{"on_call": sch})
}

func (s *Server) handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteSchedule(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Server) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Severity    string `json:"severity"`
		Description string `json:"description"`
		ResponderID string `json:"responder_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		writeJSON(w, 400, map[string]string{"error": "title is required"})
		return
	}
	inc, err := s.db.CreateIncident(req.Title, req.Severity, req.Description, req.ResponderID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"incident": inc})
}

func (s *Server) handleListIncidents(w http.ResponseWriter, r *http.Request) {
	incidents, _ := s.db.ListIncidents(r.URL.Query().Get("status"))
	if incidents == nil {
		incidents = []store.Incident{}
	}
	writeJSON(w, 200, map[string]any{"incidents": incidents, "count": len(incidents)})
}

func (s *Server) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	inc, err := s.db.GetIncident(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "incident not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"incident": inc})
}

func (s *Server) handleResolveIncident(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Resolution string `json:"resolution"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	s.db.ResolveIncident(r.PathValue("id"), req.Resolution)
	inc, _ := s.db.GetIncident(r.PathValue("id"))
	writeJSON(w, 200, map[string]any{"incident": inc})
}

func (s *Server) handleDeleteIncident(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteIncident(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.db.Stats())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
