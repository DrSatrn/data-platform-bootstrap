// This handler exposes on-demand dataset profiling. The endpoint is separate
// from the main catalog payload so the expensive runtime profile is fetched
// only for the asset the operator is currently inspecting.
package metadata

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/streanor/data-platform/backend/internal/shared"
)

// ProfileHandler serves asset profile requests.
type ProfileHandler struct {
	service *ProfileService
}

// NewProfileHandler constructs a profile endpoint.
func NewProfileHandler(service *ProfileService) http.Handler {
	return &ProfileHandler{service: service}
}

func (h *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assetID := r.URL.Query().Get("asset_id")
	if assetID == "" {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "asset_id is required",
		})
		return
	}

	profile, err := h.service.GenerateProfile(r.Context(), assetID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		if strings.HasPrefix(err.Error(), "asset ") {
			status = http.StatusNotFound
		}
		message := "failed to generate asset profile"
		if status == http.StatusNotFound {
			message = "asset profile not found"
		}
		shared.WriteError(w, status, message, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, profile)
}
