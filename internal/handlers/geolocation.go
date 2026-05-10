package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

type GeolocationHandler struct{}

func NewGeolocationHandler() *GeolocationHandler {
	return &GeolocationHandler{}
}

func (h *GeolocationHandler) Search(w http.ResponseWriter, r *http.Request) {
	street := r.URL.Query().Get("street")
	if len(street) < 2 {
		respondError(w, http.StatusBadRequest, "street parameter is required and must have at least 2 characters")
		return
	}

	nominatimURL := "https://nominatim.openstreetmap.org/search"
	reqURL, err := url.Parse(nominatimURL)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	q := reqURL.Query()
	q.Add("q", street)
	q.Add("format", "json")
	q.Add("limit", "1")
	q.Add("addressdetails", "1")
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, reqURL.String(), nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	req.Header.Set("User-Agent", "Laravel-Geolocation/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		respondError(w, http.StatusBadGateway, "Erro ao consultar serviço de geolocalização")
		return
	}
	defer resp.Body.Close()

	var body []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao processar resposta da geolocalização")
		return
	}

	if len(body) == 0 {
		respondError(w, http.StatusNotFound, "Endereço não encontrado")
		return
	}

	first := body[0]
	latStr, _ := first["lat"].(string)
	lonStr, _ := first["lon"].(string)

	lat, _ := strconv.ParseFloat(latStr, 64)
	lon, _ := strconv.ParseFloat(lonStr, 64)

	response := map[string]interface{}{
		"query":        street,
		"latitude":     lat,
		"longitude":    lon,
		"display_name": first["display_name"],
	}

	respondJSON(w, http.StatusOK, response)
}
