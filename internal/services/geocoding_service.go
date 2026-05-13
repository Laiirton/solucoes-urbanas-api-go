package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
	Bairro      string  `json:"bairro"`
	Found       bool    `json:"found"`
}

type GeocodingService struct {
	client       *http.Client
	nominatimURL string
}

func NewGeocodingService() *GeocodingService {
	tr := &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    false,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &GeocodingService{
		client: &http.Client{
			Transport: tr,
			Timeout:   10 * time.Second,
		},
		nominatimURL: "https://nominatim.openstreetmap.org/search",
	}
}

// geocodeFloatValue extracts a float64 from map[string]interface{}, trying both string and float64 types
func geocodeFloatValue(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(string); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

// GeocodeAddress busca coordenadas geográficas baseadas no endereço.
// Retorna as coordenadas padrão da cidade se não encontrar o endereço.
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
	fullAddress := address + ", " + DefaultCity + " - " + DefaultState + ", " + DefaultCountry
	q.Add("q", fullAddress)
	q.Add("format", "json")
	q.Add("limit", "1")
	q.Add("addressdetails", "1")
	q.Add("countrycodes", "br")
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
	lat := geocodeFloatValue(first, "lat")
	lon := geocodeFloatValue(first, "lon")
	displayName, _ := first["display_name"].(string)

	if lat == 0 && lon == 0 {
		return &GeocodingResult{
			Latitude:  DefaultLatitude,
			Longitude: DefaultLongitude,
			Found:     false,
		}, nil
	}

	bairro := extractBairroFromNominatim(first)

	return &GeocodingResult{
		Latitude:    lat,
		Longitude:   lon,
		DisplayName: displayName,
		Bairro:      bairro,
		Found:       true,
	}, nil
}

// extractBairroFromNominatim tenta extrair o nome do bairro do objeto de
// detalhes de endereço retornado pelo Nominatim (addressdetails=1).
func extractBairroFromNominatim(result map[string]interface{}) string {
	rawAddress, ok := result["address"]
	if !ok {
		return ""
	}

	address, ok := rawAddress.(map[string]interface{})
	if !ok {
		return ""
	}

	for _, key := range []string{"suburb", "neighbourhood", "city_district", "village", "town", "municipality"} {
		if val, ok := address[key].(string); ok && val != "" {
			return val
		}
	}

	return ""
}
