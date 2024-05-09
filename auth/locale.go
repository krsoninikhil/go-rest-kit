package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"log"

	"github.com/dghubble/sling"
	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/pkg/errors"
)

const (
	countriesFile       = "https://cdn.faithlabs.io/assets/country_list.json"
	cacheKeyCountryList = "locale_country_list"
)

type localeSvc struct {
	cache cacheClient
	sling *sling.Sling
}

func NewLocaleSvc(cache cacheClient) *localeSvc {
	return &localeSvc{
		cache: cache,
		sling: sling.New(),
	}
}

func (s *localeSvc) GetCountryInfo(ctx context.Context, locale string) (*CountryInfoSource, error) {
	val, err := s.cache.Get(cacheKeyCountryList)
	countries, ok := val.(map[string]CountryInfoSource)
	if !ok || err != nil {
		log.Printf("countries not found in cache, downloading ok= %v, err= %v", ok, err)
		countries, err = downloadCountriesFile()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if err = s.cache.Set(cacheKeyCountryList, countries, 30*24*time.Hour); err != nil {
			log.Printf("error setting countries in cache: %v", err)
		}
	}

	// countryAlphaCode
	country, ok := countries[strings.ToLower(locale)]
	if !ok {
		return nil, apperrors.NewNotFoundError("locale")
	}
	return &country, nil
}

func downloadCountriesFile() (map[string]CountryInfoSource, error) {
	var countries []CountryInfoSource
	res, err := sling.New().Get(countriesFile).ReceiveSuccess(&countries)
	if err != nil {
		return nil, apperrors.NewServerError(fmt.Errorf("error connecting to countriesFile url: %v", err))
	} else if res.StatusCode != 200 {
		return nil, apperrors.NewServerError(fmt.Errorf("error getting countries: %v", res.Status))
	}

	countryMap := make(map[string]CountryInfoSource)
	for _, country := range countries {
		countryMap[strings.ToLower(country.Code)] = country
	}
	return countryMap, nil
}
