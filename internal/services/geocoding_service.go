package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// DefaultCityCoordinates - coordenadas padrão da cidade (Cacimbas-PB)
const (
	DefaultLatitude  = -7.2000
	DefaultLongitude = -37.8000
	DefaultCity      = "Cacimbas"
	DefaultState     = "Paraíba"
	DefaultCountry   = "Brasil"
)

type GeocodingResult struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	DisplayName string  `json:"display_name"`
	Found       bool    `json:"found"`
}

type GeocodingService struct {
	client       *http.Client
	nominatimURL string
}

func NewGeocodingService() *GeocodingService {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &GeocodingService{
		client:       &http.Client{Transport: tr},
		nominatimURL: "https://nominatim.openstreetmap.org/search",
	}
}

// GeocodeAddress busca coordenadas geográficas baseadas no endereço
// Retorna as coordenadas padrão da cidade se não encontrar o endereço
func (s *GeocodingService) GeocodeAddress(address string) (*GeocodingResult, error) {
	if address == "" {
		return &GeocodingResult{
			Latitude:  DefaultLatitude,
			Longitude: DefaultLongitude,
			Found:     false,
		}, nil
	}

	reqURL, err := url.Parse(s.nominatimURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nominatim URL: %w", err)
	}

	q := reqURL.Query()
	// Adicionar cidade e estado na query para restringir à cidade padrão
	fullAddress := address + ", " + DefaultCity + " - " + DefaultState + ", " + DefaultCountry
	q.Add("q", fullAddress)
	q.Add("format", "json")
	q.Add("limit", "1")
	q.Add("addressdetails", "1")
	q.Add("countrycodes", "br")
	// Viewbox prioriza resultados na área de Cacimbas (sem bounded=1 para não restringir demais)
	q.Add("viewbox", "-38.0000,-7.0000,-37.6000,-7.4000")
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "SolucoesUrbanas-API/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return &GeocodingResult{
			Latitude:  DefaultLatitude,
			Longitude: DefaultLongitude,
			Found:     false,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &GeocodingResult{
			Latitude:  DefaultLatitude,
			Longitude: DefaultLongitude,
			Found:     false,
		}, nil
	}

	var body []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return &GeocodingResult{
			Latitude:  DefaultLatitude,
			Longitude: DefaultLongitude,
			Found:     false,
		}, nil
	}

	if len(body) == 0 {
		return &GeocodingResult{
			Latitude:  DefaultLatitude,
			Longitude: DefaultLongitude,
			Found:     false,
		}, nil
	}

	first := body[0]
	latStr, _ := first["lat"].(string)
	lonStr, _ := first["lon"].(string)
	displayName, _ := first["display_name"].(string)

	lat, _ := strconv.ParseFloat(latStr, 64)
	lon, _ := strconv.ParseFloat(lonStr, 64)

	// Se não conseguir parsear, usa o padrão
	if lat == 0 && lon == 0 {
		return &GeocodingResult{
			Latitude:  DefaultLatitude,
			Longitude: DefaultLongitude,
			Found:     false,
		}, nil
	}

	return &GeocodingResult{
		Latitude:    lat,
		Longitude:   lon,
		DisplayName: displayName,
		Found:       true,
	}, nil
}
