package main

import (
	"database/sql"
	json "encoding/json"
	fmt "fmt"
	"net/http"
	time "time"

	uuid "github.com/google/uuid"

	"github.com/mnhsh/time-capsule/internal/auth"
	"github.com/mnhsh/time-capsule/internal/config"
	"github.com/mnhsh/time-capsule/internal/database"
	response "github.com/mnhsh/time-capsule/internal/response"
)

type API struct {
	cfg *config.Config
}

func newAPI(cfg *config.Config) *API {
	return &API{cfg: cfg}
}

func (a *API) handlerCreateCapsule(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(uuid.UUID)
	if !ok {
		response.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	const maxMemory = 1 << 30
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "file too large", err)
		return
	}
	file, _, err := r.FormFile("capsule_file")
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "error retrieving file", err)
		return
	}
	defer file.Close()
	title := r.FormValue("title")
	unlockAtStr := r.FormValue("unlock_at")

	unlockAt, err := time.Parse(time.RFC3339, unlockAtStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid date format", err)
		return
	}

	s3Key := fmt.Sprintf("%s/%s", userID, uuid.New().String())
	err = a.cfg.Storage.Upload(r.Context(), s3Key, file)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "failed to upload file", err)
		return
	}
	capsuleID := uuid.New()
	err = a.cfg.DB.CreateCapsuleWithOutbox(r.Context(), database.CreateCapsuleParams{
		ID:         capsuleID,
		UserID:     userID,
		Title:      sql.NullString{String: title, Valid: title != ""},
		CreatedAt:  time.Now().UTC(),
		S3key:      s3Key,
		UnlockAt:   unlockAt.UTC(),
		IsUnlocked: sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "failed to save capsule metadata", err)
		return
	}
	response.RespondWithJSON(w, http.StatusCreated, map[string]uuid.UUID{
		"id": capsuleID,
	})
}

func (a *API) handlerGetCapsule(w http.ResponseWriter, r *http.Request) {
	type Capsule struct {
		ID         string    `json:"id"`
		Title      string    `json:"title"`
		CreatedAt  time.Time `json:"created_at"`
		UnlockAt   time.Time `json:"unlock_at"`
		IsUnlocked bool      `json:"is_unlocked"`
	}

	type CapsulesResponse struct {
		Capsules []Capsule `json:"capsules"`
	}

	userID, ok := r.Context().Value(auth.UserIDKey).(uuid.UUID)
	if !ok {
		response.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	dbCapsules, err := a.cfg.DB.GetCapsulesByUserID(r.Context(), userID)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "couldn't get user's capsules", err)
		return
	}
	capsules := make([]Capsule, 0, len(dbCapsules))
	for _, c := range dbCapsules {
		capsules = append(capsules, Capsule{
			ID:         c.ID.String(),
			Title:      c.Title.String,
			CreatedAt:  c.CreatedAt,
			UnlockAt:   c.UnlockAt,
			IsUnlocked: c.IsUnlocked.Bool,
		})
	}
	response.RespondWithJSON(w, http.StatusOK, CapsulesResponse{
		Capsules: capsules,
	})
}

func (a *API) handlerUsers(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	req := request{}
	err := decoder.Decode(&req)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "couldn't decode request", err)
		return
	}

	if req.Password == "" || req.Email == "" {
		response.RespondWithError(w, http.StatusBadRequest, "email and password are required", nil)
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "couldn't hash password", err)
		return
	}

	user, err := a.cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "couldn't create user", err)
		return
	}
	response.RespondWithJSON(w, http.StatusCreated, user)
}

func (a *API) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type res struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	decoder := json.NewDecoder(r.Body)
	req := request{}
	err := decoder.Decode(&req)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "couldn't decode request", err)
		return
	}
	user, err := a.cfg.DB.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "incorrect email or password", err)
		return
	}
	match, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}
	if !match {
		response.RespondWithError(w, http.StatusUnauthorized, "incorrect email or password", nil)
		return
	}
	accessToken, err := auth.MakeJwt(
		user.ID,
		a.cfg.JWTSecret,
		time.Minute*15,
	)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "couldn't create access JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "couldn't create refresh JWT", err)
		return
	}

	_, err = a.cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "couldn't save refresh token", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, res{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	})
}

func (a *API) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type res struct {
		AccessToken string `json:"access_token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	user, err := a.cfg.DB.GetUserByRefreshToken(r.Context(), refreshToken)
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}

	accessToken, err := auth.MakeJwt(
		user.ID,
		a.cfg.JWTSecret,
		time.Minute*15,
	)
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, res{
		AccessToken: accessToken,
	})
}

func (a *API) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	err = a.cfg.DB.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
