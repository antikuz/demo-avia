package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/antikuz/demo-avia/internal/models"
	"github.com/antikuz/demo-avia/internal/processors"
	"github.com/antikuz/demo-avia/pkg/logging"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

const (
	rootURL        = "/"
	searchURL      = "/search"
	buyURL         = "/buy/:id"
	signinURL      = "/signin"
	userProfileURL = "/profile"
	buyStatusURL   = "/buystatus"
	registerURL    = "/register"
)

type handler struct {
	templates *template.Template
	processor *processors.StorageProcessor
	sessions  map[string]*models.Session
	logger    *logging.Logger
}

func NewHandler(templates *template.Template, processor *processors.StorageProcessor, sessions map[string]*models.Session, logger *logging.Logger) *handler {
	return &handler{
		templates: templates,
		logger:    logger,
		sessions:  sessions,
		processor: processor,
	}
}

func (h *handler) isAuthorized(w http.ResponseWriter, r *http.Request) bool {
	token, err := r.Cookie("session_token")
	if err != nil {
		return false
	}
	_, ok := h.sessions[token.Value]
	if !ok {
		return false
	}

	return true
}

func (h *handler) getUser(r *http.Request) string {
	token, err := r.Cookie("session_token")
	if err != nil {
		h.logger.Errorf("Error getUser, due to err: %v", err)
	}
	return h.sessions[token.Value].Username
}

func (h *handler) Register(router *httprouter.Router) {
	router.GET(rootURL, h.GetMain)
	router.GET(buyURL, h.BuyTicket)
	router.POST(searchURL, h.GetFlights)
	router.GET(signinURL, h.SignIn)
	router.POST(signinURL, h.SignIn)
	router.POST(buyStatusURL, h.BuyStatus)
	router.GET(registerURL, h.RegisterUser)
	router.POST(registerURL, h.RegisterUser)
	router.GET(userProfileURL, h.UserProfile)
}

func (h *handler) GetMain(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	templateValues := map[string]bool{
		"auth": false,
	}
	if h.isAuthorized(w, r) {
		templateValues["auth"] = true
	}
	if err := h.templates.ExecuteTemplate(w, "main.html", templateValues); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) GetFlights(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	r.ParseForm()
	postFormValues := r.PostForm
	result := models.SearchResult{
		SearchValues:  postFormValues,
		SearchResults: h.processor.List(postFormValues),
	}
	if err := h.templates.ExecuteTemplate(w, "search.html", result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) BuyTicket(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !h.isAuthorized(w, r) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	} else {
		r.ParseForm()
		postFormValues := r.PostForm
		id := params.ByName("id")
		result := models.BuyFlightID{
			SearchValues:  postFormValues,
			SearchResults: h.processor.GetFlight(id),
		}
		if err := h.templates.ExecuteTemplate(w, "buy-ticket.html", result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *handler) SignIn(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if r.Method == "GET" {
		if err := h.templates.ExecuteTemplate(w, "signin.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		r.ParseForm()
		creds := r.PostForm
		user := h.processor.GetUser(creds["username"][0])
		if user.Password == "" || user.Password != creds["password"][0] {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		sessionToken := uuid.NewString()
		expiresAt := time.Now().Add(6000 * time.Second)

		h.sessions[sessionToken] = &models.Session{
			Username: creds["username"][0],
			Expiry:   expiresAt,
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: expiresAt,
		})

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}

func (h *handler) UserProfile(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !h.isAuthorized(w, r) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	} else {
		user := h.getUser(r)
		flights := h.processor.GetUserFlights(user)
		if err := h.templates.ExecuteTemplate(w, "profile.html", flights); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *handler) BuyStatus(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !h.isAuthorized(w, r) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	} else {
		r.ParseForm()
		postFormValues := r.PostForm
		BuySuccess := h.processor.BuyTicket(postFormValues)
		if BuySuccess {
			fmt.Fprintf(w, "Покупка успешная")
		} else {
			fmt.Fprintf(w, "Покупка неудалась")
		}
	}
}

func (h *handler) RegisterUser(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if r.Method == "GET" {
		if err := h.templates.ExecuteTemplate(w, "register.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		r.ParseForm()
		formValues := r.PostForm

		userRegistered := h.processor.RegisterUser(formValues)
		if !userRegistered {
			return
		}
		sessionToken := uuid.NewString()
		expiresAt := time.Now().Add(3600 * time.Second)

		h.sessions[sessionToken] = &models.Session{
			Username: formValues["username"][0],
			Expiry:   expiresAt,
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: expiresAt,
		})

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}
