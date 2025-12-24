package mockserver

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	mu          sync.Mutex
	connections map[string]*connectionState
}

type connectionState struct {
	Status     string
	Note       string
	InvitedAt  time.Time
	AcceptedAt time.Time
	Messages   []string
}

func New() *Server {
	return &Server{
		connections: map[string]*connectionState{},
	}
}

func (s *Server) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/feed/", s.requireAuth(s.handleFeed))
	mux.HandleFunc("/mynetwork", s.requireAuth(s.handleMyNetwork))
	mux.HandleFunc("/search/results/people/", s.requireAuth(s.handleSearchPeople))
	mux.HandleFunc("/in/", s.requireAuth(s.handleProfile))

	mux.HandleFunc("/api/connect", s.requireAuth(s.handleAPIConnect))
	mux.HandleFunc("/api/sendInvite", s.requireAuth(s.handleAPISendInvite))
	mux.HandleFunc("/api/message", s.requireAuth(s.handleAPIMessage))

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server.ListenAndServe()
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/feed/", http.StatusFound)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, loginHTML)
}

func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, feedHTML(r.Host))
}

func (s *Server) handleMyNetwork(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, myNetworkHTML(r.Host))
}

func (s *Server) handleSearchPeople(w http.ResponseWriter, r *http.Request) {
	page := parseInt(r.URL.Query().Get("page"), 1)
	keywords := r.URL.Query().Get("keywords")
	if page < 1 {
		page = 1
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, searchHTML(r.Host, page, keywords))
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/in/")
	userID = strings.TrimSuffix(userID, "/")
	if userID == "" {
		http.NotFound(w, r)
		return
	}

	s.mu.Lock()
	state := s.getOrCreate(userID)
	status := s.resolveStatus(state)
	state.Status = status
	s.mu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, profileHTML(r.Host, userID, status))
}

func (s *Server) handleAPIConnect(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("target"))
	if target == "" {
		http.Error(w, "missing target", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	state := s.getOrCreate(target)
	state.Status = "pending"
	state.InvitedAt = time.Now()
	state.AcceptedAt = time.Now().Add(2 * time.Second)
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPISendInvite(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("target"))
	note := r.URL.Query().Get("note")
	if target == "" {
		http.Error(w, "missing target", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	state := s.getOrCreate(target)
	state.Note = note
	state.Status = "pending"
	if state.InvitedAt.IsZero() {
		state.InvitedAt = time.Now()
	}
	if state.AcceptedAt.IsZero() {
		state.AcceptedAt = time.Now().Add(2 * time.Second)
	}
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPIMessage(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("target"))
	body := strings.TrimSpace(r.URL.Query().Get("body"))
	if target == "" {
		http.Error(w, "missing target", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	state := s.getOrCreate(target)
	state.Messages = append(state.Messages, body)
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("li_at")
		if cookie == nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

func (s *Server) getOrCreate(userID string) *connectionState {
	state, ok := s.connections[userID]
	if !ok {
		state = &connectionState{Status: "none"}
		s.connections[userID] = state
	}
	return state
}

func (s *Server) resolveStatus(state *connectionState) string {
	if state.Status == "pending" && !state.AcceptedAt.IsZero() && time.Now().After(state.AcceptedAt) {
		return "accepted"
	}
	return state.Status
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return value
}

func feedHTML(host string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>Mock LinkedIn - Feed</title>
    <style>
      body { font-family: system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif; margin: 0; background: #f6f7fb; }
      header { padding: 14px 18px; background: white; border-bottom: 1px solid #e7e8ee; }
      a { color: #1a73e8; text-decoration: none; }
      .wrap { padding: 18px; max-width: 920px; margin: 0 auto; }
      .card { background: white; border: 1px solid #e7e8ee; border-radius: 12px; padding: 14px; }
      .row { display: flex; gap: 12px; flex-wrap: wrap; }
      .pill { display: inline-block; padding: 6px 10px; border-radius: 999px; font-size: 12px; background: #eef2ff; border: 1px solid #dbe3ff; color: #2b47a7; }
    </style>
  </head>
  <body>
    <header>
      <div class="row">
        <a href="/feed/">Home</a>
        <a href="/mynetwork">My Network</a>
        <a href="/search/results/people/?keywords=demo&page=1">Search People</a>
      </div>
    </header>
    <div class="wrap">
      <div class="card">
        <h1 style="font-size:16px;margin:0 0 6px 0;">Mock LinkedIn Feed</h1>
        <div class="pill">host=%s</div>
        <p style="margin:10px 0 0 0;color:#555;">This is a local mock site for demonstrating automation without using LinkedIn.</p>
      </div>
    </div>
  </body>
</html>`, html.EscapeString(host))
}

func myNetworkHTML(host string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>Mock LinkedIn - My Network</title>
  </head>
  <body style="font-family: system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif;">
    <p><a href="/feed/">Home</a></p>
    <p><strong>My Network</strong> (mock)</p>
    <p>Host: %s</p>
  </body>
</html>`, html.EscapeString(host))
}

func searchHTML(host string, page int, keywords string) string {
	escapedKeywords := html.EscapeString(keywords)
	links := []string{}
	for i := 1; i <= 6; i++ {
		user := fmt.Sprintf("demo-%d-%d", page, i)
		absolute := fmt.Sprintf("http://%s/in/%s/", host, user)
		links = append(links, fmt.Sprintf(`<li><a class="app-aware-link" href="%s">Demo User %s</a></li>`, html.EscapeString(absolute), html.EscapeString(user)))
	}
	if page > 1 {
		user := "demo-1-1"
		absolute := fmt.Sprintf("http://%s/in/%s/", host, user)
		links = append(links, fmt.Sprintf(`<li><a class="app-aware-link" href="%s">Duplicate %s</a></li>`, html.EscapeString(absolute), html.EscapeString(user)))
	}
	next := fmt.Sprintf("/search/results/people/?keywords=%s&page=%d", urlQueryEscape(keywords), page+1)
	prev := fmt.Sprintf("/search/results/people/?keywords=%s&page=%d", urlQueryEscape(keywords), max(1, page-1))
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>Mock LinkedIn - Search People</title>
    <style>
      body { font-family: system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif; margin: 0; background: #f6f7fb; }
      header { padding: 14px 18px; background: white; border-bottom: 1px solid #e7e8ee; }
      a { color: #1a73e8; text-decoration: none; }
      .wrap { padding: 18px; max-width: 920px; margin: 0 auto; }
      .card { background: white; border: 1px solid #e7e8ee; border-radius: 12px; padding: 14px; }
      ul { margin: 10px 0 0 18px; }
      .nav { display: flex; gap: 10px; margin-top: 12px; }
    </style>
  </head>
  <body>
    <header>
      <a href="/feed/">Home</a> · <a href="/mynetwork">My Network</a>
    </header>
    <div class="wrap">
      <div class="card">
        <h1 style="font-size:16px;margin:0 0 6px 0;">People Search</h1>
        <div style="color:#555;">keywords=%s · page=%d</div>
        <ul>%s</ul>
        <div class="nav">
          <a href="%s">Prev</a>
          <a href="%s">Next</a>
        </div>
      </div>
    </div>
  </body>
</html>`, escapedKeywords, page, strings.Join(links, ""), html.EscapeString(prev), html.EscapeString(next))
}

func profileHTML(host, userID, status string) string {
	name := "Demo User " + userID
	headline := "Software Engineer at Example Corp"
	actionButton := ""
	switch status {
	case "accepted":
		actionButton = `<button id="message-btn" type="button">Message</button>`
	case "pending":
		actionButton = `<button id="pending-btn" type="button">Pending</button>`
	default:
		actionButton = `<button id="connect-btn" type="button">Connect</button>`
	}
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>Mock LinkedIn - Profile</title>
    <style>
      body { font-family: system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif; margin: 0; background: #f6f7fb; }
      header { padding: 14px 18px; background: white; border-bottom: 1px solid #e7e8ee; }
      a { color: #1a73e8; text-decoration: none; }
      .wrap { padding: 18px; max-width: 920px; margin: 0 auto; }
      .card { background: white; border: 1px solid #e7e8ee; border-radius: 12px; padding: 14px; }
      button { padding: 10px 12px; border: 0; border-radius: 10px; background: #1a73e8; color: white; font-weight: 600; cursor: pointer; }
      #pending-btn { background: #666; cursor: default; }
      #connect-modal, #message-modal { margin-top: 12px; padding: 12px; border: 1px solid #e7e8ee; border-radius: 12px; background: #fafbff; display: none; }
      textarea { width: 100%%; min-height: 80px; padding: 10px 12px; border: 1px solid #cfd3e1; border-radius: 10px; }
      .msg-form__contenteditable { border: 1px solid #cfd3e1; border-radius: 10px; padding: 10px 12px; min-height: 64px; background: white; }
      .row { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
      .hint { font-size: 12px; color: #666; margin-top: 8px; }
    </style>
  </head>
  <body>
    <header>
      <a href="/feed/">Home</a> · <a href="/search/results/people/?keywords=demo&page=1">Search</a>
    </header>
    <div class="wrap">
      <div class="card">
        <h1>%s</h1>
        <div class="text-body-medium">%s</div>
        <div class="hint">status=%s · host=%s</div>
        <div class="row" style="margin-top:12px;">
          %s
        </div>

        <div id="connect-modal">
          <div class="row">
            <button id="add-note-btn" type="button" aria-label="Add a note">Add a note</button>
            <button id="send-btn" type="button">Send</button>
          </div>
          <div style="margin-top:10px;">
            <textarea name="message" id="note-text" placeholder="Personalized note (mock)"></textarea>
          </div>
          <div class="hint">Connect flow: Connect → Add a note → type note → Send.</div>
        </div>

        <div id="message-modal">
          <div class="msg-form__contenteditable" id="msg-box" role="textbox" contenteditable="true"></div>
          <div class="hint">Message flow: click Message → type → press Enter to send.</div>
        </div>
      </div>
    </div>
    <script>
      const user = %q;
      const status = %q;
      const connectBtn = document.getElementById('connect-btn');
      const messageBtn = document.getElementById('message-btn');
      const connectModal = document.getElementById('connect-modal');
      const messageModal = document.getElementById('message-modal');

      const addNote = document.getElementById('add-note-btn');
      const noteText = document.getElementById('note-text');
      const sendBtn = document.getElementById('send-btn');

      if (connectBtn) {
        connectBtn.addEventListener('click', async () => {
          connectModal.style.display = 'block';
          try { await fetch('/api/connect?target=' + encodeURIComponent(user)); } catch (e) {}
        });
      }

      if (addNote) {
        addNote.addEventListener('click', () => {
          noteText.focus();
        });
      }

      if (sendBtn) {
        sendBtn.addEventListener('click', async () => {
          const note = noteText.value || '';
          try { await fetch('/api/sendInvite?target=' + encodeURIComponent(user) + '&note=' + encodeURIComponent(note)); } catch (e) {}
          connectModal.style.display = 'none';
        });
      }

      if (messageBtn) {
        messageBtn.addEventListener('click', () => {
          messageModal.style.display = 'block';
          document.getElementById('msg-box').focus();
        });
      }

      const msgBox = document.getElementById('msg-box');
      if (msgBox) {
        msgBox.addEventListener('keydown', async (e) => {
          if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            const body = msgBox.innerText.trim();
            if (body) {
              try { await fetch('/api/message?target=' + encodeURIComponent(user) + '&body=' + encodeURIComponent(body)); } catch (e) {}
              msgBox.innerText = '';
            }
          }
        });
      }
    </script>
  </body>
</html>`, html.EscapeString(name), html.EscapeString(headline), html.EscapeString(status), html.EscapeString(host), actionButton, userID, status)
}

const loginHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>Mock LinkedIn - Login</title>
    <style>
      body { font-family: system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif; margin: 0; background: #f6f7fb; }
      .wrap { padding: 28px; max-width: 520px; margin: 0 auto; }
      .card { background: white; border: 1px solid #e7e8ee; border-radius: 12px; padding: 16px; }
      label { display:block; font-weight:600; margin: 10px 0 6px; }
      input { width: 100%; font-size: 14px; padding: 10px 12px; border: 1px solid #cfd3e1; border-radius: 10px; outline: none; }
      button { margin-top: 12px; padding: 10px 12px; border: 0; border-radius: 10px; background: #1a73e8; color: white; font-weight: 600; cursor: pointer; width: 100%; }
      .hint { font-size: 12px; color: #666; margin-top: 10px; }
    </style>
  </head>
  <body>
    <div class="wrap">
      <div class="card">
        <h1 style="font-size:16px;margin:0 0 6px 0;">Mock LinkedIn Login</h1>
        <form id="login-form">
          <label for="username">Email</label>
          <input id="username" name="username" autocomplete="username" />
          <label for="password">Password</label>
          <input id="password" name="password" type="password" autocomplete="current-password" />
          <button type="submit">Sign in</button>
        </form>
        <div class="hint">Any credentials work. On submit, a session cookie is set and you are redirected to /feed/.</div>
      </div>
    </div>
    <script>
      document.getElementById('login-form').addEventListener('submit', (e) => {
        e.preventDefault();
        document.cookie = "li_at=mock; Path=/";
        window.location = "/feed/";
      });
    </script>
  </body>
</html>`

func urlQueryEscape(raw string) string {
	replacer := strings.NewReplacer(" ", "+", "%", "%25", "&", "%26", "?", "%3F", "#", "%23", "=", "%3D", "/", "%2F")
	return replacer.Replace(raw)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
