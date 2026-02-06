package api

import (
	"database/sql"
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

func NewAPI(cfg *config.Config) *API {
	return &API{cfg: cfg}
}

func (a *API) HandlerCreateCapsule(w http.ResponseWriter, r *http.Request) {
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
