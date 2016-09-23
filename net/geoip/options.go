// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package geoip

import (
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/corestoreio/csfw/log"
	"github.com/corestoreio/csfw/net/mw"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/corestoreio/csfw/util/errors"
)

// IsAllowedFunc checks in middleware WithIsCountryAllowedByIP if the country is
// allowed to process the request. The StringSlice contains a list of ISO
// country names fetched from the config.ScopedGetter. Return nil to indicate
// that the request can continue.
type IsAllowedFunc func(id scope.TypeID, c *Country, allowedCountries []string) error

// WithDefaultConfig applies the default GeoIP configuration settings based for
// a specific scope. This function overwrites any previous set options.
//
// Default values are:
//		- Alternative Handler: variable DefaultAlternativeHandler
//		- Logger black hole
//		- Check allow: If allowed countries are empty, all countries are allowed
func WithDefaultConfig(id scope.TypeID) Option {
	return withDefaultConfig(id)
}

// WithAlternativeHandler sets for a scope the alternative handler
// if an IP address has been access denied.
// Only to be used with function WithIsCountryAllowedByIP()
func WithAlternativeHandler(id scope.TypeID, altHndlr mw.ErrorHandler) Option {
	return func(s *Service) error {
		s.rwmu.Lock()
		defer s.rwmu.Unlock()

		sc := s.scopeCache[id]
		if sc == nil {
			sc = optionInheritDefault(s)
		}
		sc.AlternativeHandler = altHndlr
		sc.ScopeID = id
		s.scopeCache[id] = sc
		return nil
	}
}

// WithAlternativeRedirect sets for a scope the error handler
// on a Service if an IP address has been access denied.
// Only to be used with function WithIsCountryAllowedByIP()
func WithAlternativeRedirect(id scope.TypeID, urlStr string, code int) Option {
	return WithAlternativeHandler(id, func(_ error) http.Handler {
		return http.RedirectHandler(urlStr, code)
	})
}

// WithCheckAllow sets your custom function which checks if the country of an IP
// address should access to granted, or the next middleware handler in the chain
// gets called.
// Only to be used with function WithIsCountryAllowedByIP()
func WithCheckAllow(id scope.TypeID, f IsAllowedFunc) Option {
	return func(s *Service) error {
		s.rwmu.Lock()
		defer s.rwmu.Unlock()

		sc := s.scopeCache[id]
		if sc == nil {
			sc = optionInheritDefault(s)
		}
		sc.IsAllowedFunc = f
		sc.ScopeID = id
		s.scopeCache[id] = sc
		return nil
	}
}

// WithAllowedCountryCodes sets a list of ISO countries to be validated against.
// Only to be used with function WithIsCountryAllowedByIP()
func WithAllowedCountryCodes(id scope.TypeID, isoCountryCodes ...string) Option {
	return func(s *Service) error {
		s.rwmu.Lock()
		defer s.rwmu.Unlock()

		sc := s.scopeCache[id]
		if sc == nil {
			sc = optionInheritDefault(s)
		}
		sc.AllowedCountries = isoCountryCodes
		sc.ScopeID = id
		s.scopeCache[id] = sc
		return nil
	}
}

// WithGeoIP applies a custom CountryRetriever. Sets the retriever atomically
// and only once.
func WithGeoIP(cr CountryRetriever) Option {
	return func(s *Service) error {
		if s.isGeoIPLoaded() {
			if s.Log.IsDebug() {
				s.Log.Debug("geoip.WithGeoIP.geoIPDone", log.Int("done", 1))
			}
			return nil
		}
		s.rwmu.Lock()
		defer s.rwmu.Unlock()
		if s.geoIPLoaded == 0 {
			s.geoIP = cr
			atomic.StoreUint32(&s.geoIPLoaded, 1)
			if s.Log.IsDebug() {
				s.Log.Debug("geoip.WithGeoIP.geoIPLoaded", log.Int("done", 0))
			}
		}
		return nil
	}
}

// WithGeoIP2File creates a new GeoIP2.Reader. As long as there are no other
// readers this is a mandatory argument. Error behaviour: NotFound, NotValid
func WithGeoIP2File(filename string) Option {
	return func(s *Service) error {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return errors.NewNotFoundf("[geoip] File %q not found", filename)
		}

		cr, err := newMMDBByFile(filename)
		if err != nil {
			return errors.NewNotValidf("[geoip] Maxmind Open %s with file %q", err, filename)
		}
		return WithGeoIP(cr)(s)
	}
}

// WithGeoIP2Webservice uses for each incoming a request a lookup request to the
// Maxmind Webservice http://dev.maxmind.com/geoip/geoip2/web-services/ and
// caches the result in Transcacher. Hint: use package storage/transcache. If
// the httpTimeout is lower 0 then the default 20s get applied.
func WithGeoIP2Webservice(t TransCacher, userID, licenseKey string, httpTimeout time.Duration) Option {
	if httpTimeout < 1 {
		httpTimeout = time.Second * 20
	}
	return WithGeoIP2WebserviceHTTPClient(t, userID, licenseKey, &http.Client{Timeout: httpTimeout})
}

// WithGeoIP2WebserviceHTTPClient uses for each incoming a request a lookup
// request to the Maxmind Webservice
// http://dev.maxmind.com/geoip/geoip2/web-services/ and caches the result in
// Transcacher. Hint: use package storage/transcache.
func WithGeoIP2WebserviceHTTPClient(t TransCacher, userID, licenseKey string, hc *http.Client) Option {
	return WithGeoIP(newMMWS(t, userID, licenseKey, hc))

}
