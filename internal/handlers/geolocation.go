package handlers

import (
	"crypto/tls"
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
		http.Error(w, "street parameter is required and must have at least 2 characters", http.StatusBadRequest)
		return
	}

	nominatimURL := "https://nominatim.openstreetmap.org/search"
	reqURL, err := url.Parse(nominatimURL)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("User-Agent", "Laravel-Geolocation/1.0")

	// Skipping cert check (since the Laravel app uses a specific local cert for verify, using default or skipping)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Erro ao consultar serviço de geolocalização", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var body []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		http.Error(w, "Erro ao processar resposta da geolocalização", http.StatusInternalServerError)
		return
	}

	if len(body) == 0 {
		http.Error(w, "Endereço não encontrado", http.StatusNotFound)
		return
	}

	first := body[0]
	latStr, _ := first["lat"].(string)
	lonStr, _ := first["lon"].(string)

	lat, _ := strconv.ParseFloat(latStr, 64)
	lon, _ := strconv.ParseFloat(lonStr, 64)
	displayName, _ := first["display_name"].(string)

	response := map[string]interface{}{
		"query":        street,
		"latitude":     lat,
		"longitude":    lon,
		"display_name": displayName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
