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

package store_test

import (
	"testing"

	"github.com/corestoreio/csfw/config/cfgmock"
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/corestoreio/csfw/store"
	"github.com/stretchr/testify/assert"
)

func TestGroupSlice_Map_Each(t *testing.T) {
	gs := store.GroupSlice{
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 1, WebsiteID: 1, Name: "DACH Group", RootCategoryID: 2, DefaultStoreID: 2},
			&store.TableWebsite{WebsiteID: 1, Code: dbr.NewNullString("euro"), Name: dbr.NewNullString("Europe"), SortOrder: 0, DefaultGroupID: 1, IsDefault: dbr.NewNullBool(true)},
			store.TableStoreSlice{
				&store.TableStore{StoreID: 1, Code: dbr.NewNullString("de"), WebsiteID: 1, GroupID: 1, Name: "Germany", SortOrder: 10, IsActive: true},
			},
		),
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 2, WebsiteID: 2, Name: "DACH2 Group", RootCategoryID: 2, DefaultStoreID: 2},
			&store.TableWebsite{WebsiteID: 2, Code: dbr.NewNullString("euro2"), Name: dbr.NewNullString("Europe"), SortOrder: 0, DefaultGroupID: 2, IsDefault: dbr.NewNullBool(true)},
			store.TableStoreSlice{
				&store.TableStore{StoreID: 2, Code: dbr.NewNullString("de2"), WebsiteID: 2, GroupID: 1, Name: "Germany", SortOrder: 10, IsActive: true},
			},
		),
	}

	gs.
		Map(func(g *store.Group) {
			g.Data.GroupID = 4
			g.Website.Data.Name.String = "Gopher"
		}).
		Each(func(g store.Group) {
			assert.Exactly(t, "Gopher", g.Website.Name())
		})
	assert.Exactly(t, []int64{4, 4}, gs.IDs())
}

func TestGroupSlice_Sort(t *testing.T) {
	gs := store.GroupSlice{
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 2, WebsiteID: 1, RootCategoryID: 2, DefaultStoreID: 2},
			nil,
			nil,
		),
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 1, WebsiteID: 2, RootCategoryID: 2, DefaultStoreID: 2},
			nil,
			nil,
		),
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 3, WebsiteID: 2, RootCategoryID: 2, DefaultStoreID: 2},
			nil,
			nil,
		),
	}
	gs.Sort()
	assert.Exactly(t, []int64{1, 2, 3}, gs.IDs())
}

func TestGroupSlice_IDs(t *testing.T) {
	gs := store.GroupSlice{
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 2, WebsiteID: 1, RootCategoryID: 2, DefaultStoreID: 2},
			nil,
			nil,
		),
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 1, WebsiteID: 2, RootCategoryID: 2, DefaultStoreID: 2},
			nil,
			nil,
		),
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 3, WebsiteID: 2, RootCategoryID: 2, DefaultStoreID: 2},
			nil,
			nil,
		),
	}
	assert.Exactly(t, []int64{2, 1, 3}, gs.IDs())
	assert.Nil(t, store.GroupSlice{}.IDs())
}

func TestGroupSlice_FindByID(t *testing.T) {
	gs := store.GroupSlice{
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 2, WebsiteID: 1, RootCategoryID: 2, DefaultStoreID: 2},
			nil,
			nil,
		),
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 1, WebsiteID: 2, RootCategoryID: 2, DefaultStoreID: 2},
			nil,
			nil,
		),
		store.MustNewGroup(
			cfgmock.NewService(),
			&store.TableGroup{GroupID: 3, WebsiteID: 2, RootCategoryID: 2, DefaultStoreID: 2},
			nil,
			nil,
		),
	}

	g, gOK := gs.FindByID(1)
	assert.True(t, gOK)
	assert.Exactly(t, int64(1), g.ID())
	g, gOK = gs.FindByID(44)
	assert.Nil(t, g.Data)
	assert.False(t, gOK)
}
