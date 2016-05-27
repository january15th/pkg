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

package backendgeoip_test

import (
	"testing"

	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/corestoreio/csfw/config/cfgmock"
	"github.com/corestoreio/csfw/config/cfgpath"
	"github.com/corestoreio/csfw/net/geoip"
	"github.com/corestoreio/csfw/net/geoip/backendgeoip"
	"github.com/corestoreio/csfw/store"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/corestoreio/csfw/store/storemock"
	"github.com/corestoreio/csfw/util/errors"
	"github.com/stretchr/testify/assert"
	"os"
)

func mustToPath(t *testing.T, f func(s scope.Scope, scopeID int64) (cfgpath.Path, error), s scope.Scope, scopeID int64) string {
	p, err := f(s, scopeID)
	if err != nil {
		t.Fatal(errors.PrintLoc(err))
	}
	return p.String()
}
func mustGetTestService(opts ...geoip.Option) *geoip.Service {
	maxMindDB := filepath.Join("../", "testdata", "GeoIP2-Country-Test.mmdb")
	return geoip.MustNew(append(opts, geoip.WithGeoIP2File(maxMindDB))...)
}

func TestBackend_WithGeoIP2Webservice_Redis(t *testing.T) {
	redConURL := os.Getenv("CS_REDIS_TEST") // redis://127.0.0.1:6379/3
	if redConURL == "" {
		t.Skip(`Skipping live test because environment CS_REDIS_TEST variable not found.
	export CS_REDIS_TEST="redis://127.0.0.1:6379/3"
		`)
	}

	// clear all redis keys

	// test if we get the correct country and if the country has
	// been successfully stored in redis and can be retrieved.

	scpFnc := backendgeoip.Default()
	cfgSrv := cfgmock.NewService(cfgmock.WithPV(cfgmock.PathValue{
		// @see structure.go for the limitation to scope.Default
		mustToPath(t, backend.NetGeoipMaxmindWebserviceUserID.ToPath, scope.Default, 0):   `TestUserID`,
		mustToPath(t, backend.NetGeoipMaxmindWebserviceLicense.ToPath, scope.Default, 0):  `TestLicense`,
		mustToPath(t, backend.NetGeoipMaxmindWebserviceTimeout.ToPath, scope.Default, 0):  `21s`,
		mustToPath(t, backend.NetGeoipMaxmindWebserviceRedisURL.ToPath, scope.Default, 0): redConURL,
	}))
	cfgScp := cfgSrv.NewScoped(1, 2) // Website ID 2 == euro / Store ID == 2 Austria ==> here doesn't matter

	geoSrv := mustGetTestService()
	if err := geoSrv.Options(scpFnc(cfgScp)...); err != nil {
		t.Fatal(errors.PrintLoc(err))
	}

	req := func() *http.Request {
		req, _ := http.NewRequest("GET", "http://corestore.io", nil)
		req.RemoteAddr = "2a02:d180::" // Germany
		return req
	}()

	geoSrv.WithCountryByIP()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {

		cty, err := geoip.FromContextCountry(r.Context())
		t.Logf("%#v", cty)
		if err != nil {
			t.Error(errors.PrintLoc(err))
		}
		assert.Exactly(t, "DE", cty.Country.IsoCode)

	})).ServeHTTP(nil, req)

}

func TestBackend_WithAlternativeRedirect(t *testing.T) {

	scpFnc := backendgeoip.Default()
	cfgSrv := cfgmock.NewService(cfgmock.WithPV(cfgmock.PathValue{
		// @see structure.go why scope.Store and scope.Website can be used.
		mustToPath(t, backend.NetGeoipAlternativeRedirect.ToPath, scope.Store, 2):       `https://byebye.de.io`,
		mustToPath(t, backend.NetGeoipAlternativeRedirectCode.ToPath, scope.Website, 1): 307,
	}))
	cfgScp := cfgSrv.NewScoped(1, 2) // Website ID 2 == euro / Store ID == 2 Austria

	geoSrv := mustGetTestService(geoip.WithOptionFactory(scpFnc(cfgScp)))
	if err := geoSrv.Options(geoip.WithAllowedCountryCodes(scope.Store, 2, "AT", "CH")); err != nil {
		t.Fatal(errors.PrintLoc(err))
	}

	// Germany is not allowed and must be redirected to https://byebye.de.io with code 307
	req := func() *http.Request {
		o, err := scope.SetByCode(scope.Website, "euro")
		if err != nil {
			t.Fatal(err)
		}
		storeSrv := storemock.NewEurozzyService(o)
		req, _ := http.NewRequest("GET", "http://corestore.io", nil)
		req.RemoteAddr = "2a02:d180::"
		st, err := storeSrv.Store(scope.MockID(2)) // Austria Store
		if err != nil {
			t.Fatal(errors.PrintLoc(err))
		}
		st.Config = cfgmock.NewService().NewScoped(1, 2)
		return req.WithContext(store.WithContextRequestedStore(req.Context(), st))
	}()

	rec := httptest.NewRecorder()
	geoSrv.WithIsCountryAllowedByIP()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		panic("Should not be called")

	})).ServeHTTP(rec, req)

	assert.Exactly(t, `https://byebye.de.io`, rec.Header().Get("Location"))
	assert.Exactly(t, 307, rec.Code)
}

func TestBackend_Path_Errors(t *testing.T) {

	tests := []struct {
		toPath func(s scope.Scope, scopeID int64) (cfgpath.Path, error)
		val    interface{}
		errBhf errors.BehaviourFunc
	}{
		{backend.NetGeoipAllowedCountries.ToPath, struct{}{}, errors.IsNotValid},
		{backend.NetGeoipAlternativeRedirect.ToPath, struct{}{}, errors.IsNotValid},
		{backend.NetGeoipAlternativeRedirectCode.ToPath, struct{}{}, errors.IsNotValid},
		{backend.NetGeoipMaxmindLocalFile.ToPath, "fileNotFound.txt", errors.IsNotFound},
		{backend.NetGeoipMaxmindLocalFile.ToPath, struct{}{}, errors.IsNotValid},
		{backend.NetGeoipMaxmindWebserviceUserID.ToPath, struct{}{}, errors.IsNotValid},
		{backend.NetGeoipMaxmindWebserviceLicense.ToPath, struct{}{}, errors.IsNotValid},
		{backend.NetGeoipMaxmindWebserviceTimeout.ToPath, struct{}{}, errors.IsNotValid},
		{backend.NetGeoipMaxmindWebserviceRedisURL.ToPath, struct{}{}, errors.IsNotValid},
	}
	for i, test := range tests {

		scpFnc := backendgeoip.Default()
		cfgSrv := cfgmock.NewService(cfgmock.WithPV(cfgmock.PathValue{
			mustToPath(t, test.toPath, scope.Default, 0): test.val,
		}))
		cfgScp := cfgSrv.NewScoped(2, 0)

		_, err := geoip.New(scpFnc(cfgScp)...)
		assert.True(t, test.errBhf(err), "Index %d Error: %s", i, err)
	}
}
