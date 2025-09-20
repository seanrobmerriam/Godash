package main

import (
    "encoding/json"
    "html/template"
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
)

// Session store for authentication
var store = sessions.NewCookieStore([]byte("very-secret-key"))

// In-memory user for demo purposes
var adminUser = struct {
    Username string
    Password string
}{"admin", "password"}

// CaddyInstance represents a managed Caddy server
type CaddyInstance struct {
    ID      string
    Name    string
    Running bool
    Logs    []string
    Mutex   sync.Mutex
}

// Simulated instances
var instances = []*CaddyInstance{
    {ID: "1", Name: "Main Site", Running: true, Logs: []string{"Started", "Serving on :80"}},
    {ID: "2", Name: "API Server", Running: false, Logs: []string{"Stopped"}},
}

// Templates
var tmpl = template.Must(template.ParseGlob("templates/*.html"))

// Middleware for authentication
func authRequired(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session")
        if session.Values["authenticated"] != true {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Home/dashboard view
func HomeHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/instances", http.StatusFound)
}

// Login page and logic
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        tmpl.ExecuteTemplate(w, "login.html", nil)
        return
    }
    // POST: process login
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form", http.StatusBadRequest)
        return
    }
    username := r.FormValue("username")
    password := r.FormValue("password")
    if username == adminUser.Username && password == adminUser.Password {
        session, _ := store.Get(r, "session")
        session.Values["authenticated"] = true
        session.Save(r, w)
        http.Redirect(w, r, "/instances", http.StatusFound)
        return
    }
    tmpl.ExecuteTemplate(w, "login.html", "Invalid credentials")
}

// List all Caddy instances
func InstancesHandler(w http.ResponseWriter, r *http.Request) {
    tmpl.ExecuteTemplate(w, "instances.html", struct {
        Instances []*CaddyInstance
    }{instances})
}

// Start a Caddy instance
func StartInstanceHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    for _, inst := range instances {
        if inst.ID == vars["id"] {
            inst.Mutex.Lock()
            if !inst.Running {
                inst.Running = true
                inst.Logs = append(inst.Logs, "Started")
            }
            inst.Mutex.Unlock()
            break
        }
    }
    http.Redirect(w, r, "/instances", http.StatusFound)
}

// Stop a Caddy instance
func StopInstanceHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    for _, inst := range instances {
        if inst.ID == vars["id"] {
            inst.Mutex.Lock()
            if inst.Running {
                inst.Running = false
                inst.Logs = append(inst.Logs, "Stopped")
            }
            inst.Mutex.Unlock()
            break
        }
    }
    http.Redirect(w, r, "/instances", http.StatusFound)
}

// Retrieve logs for a Caddy instance (JSON)
func LogsHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    for _, inst := range instances {
        if inst.ID == vars["id"] {
            inst.Mutex.Lock()
            logs := inst.Logs
            inst.Mutex.Unlock()
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(logs)
            return
        }
    }
    http.Error(w, "Instance not found", http.StatusNotFound)
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/", HomeHandler)
    r.HandleFunc("/login", LoginHandler)
    r.Handle("/instances", authRequired(http.HandlerFunc(InstancesHandler)))
    r.Handle("/instance/{id}/start", authRequired(http.HandlerFunc(StartInstanceHandler))).Methods("POST")
    r.Handle("/instance/{id}/stop", authRequired(http.HandlerFunc(StopInstanceHandler))).Methods("POST")
    r.Handle("/instance/{id}/logs", authRequired(http.HandlerFunc(LogsHandler)))

    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

    log.Println("CaddyDash running on :8080")
    http.ListenAndServe(":8080", r)
}
