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

package element

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/corestoreio/csfw/config/cfgpath"
	"github.com/corestoreio/csfw/storage/text"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/juju/errors"
)

// ErrSectionNotFound error when a section cannot be found.
var ErrSectionNotFound = errors.New("Section not found")

// SectionSlice contains a set of Sections. Some nifty helper functions
// exists. Thread safe for reading. A section slice can be used in many
// goroutines. It must remain lock-free.
type SectionSlice []*Section

// Section defines the layout for the configuration section which contains
// groups and fields. Thread safe for reading but not for modifying.
type Section struct {
	// ID unique ID and merged with others. 1st part of the path.
	ID    cfgpath.Route
	Label text.Chars `json:",omitempty"`
	// Scopes: bit value eg: showInDefault="1" showInWebsite="1" showInStore="1"
	// Scopes can contain multiple Scope but no more than Default, Website and Store.
	Scopes    scope.Perm `json:",omitempty"`
	SortOrder int        `json:",omitempty"`
	// Resource some kind of ACL if someone is allowed for no,read or write access @todo
	Resource uint `json:",omitempty"`
	Groups   GroupSlice
}

// NewSectionSlice wrapper function, for now.
func NewSectionSlice(s ...*Section) SectionSlice {
	return SectionSlice(s)
}

// NewConfiguration creates a new validated SectionSlice with a three level configuration.
// Panics if a path is redundant.
func NewConfiguration(sections ...*Section) (SectionSlice, error) {
	ss := NewSectionSlice(sections...)
	if err := ss.Validate(); err != nil {
		if PkgLog.IsDebug() {
			PkgLog.Debug("config.NewConfiguration.Validate", "err", err)
		}
		return nil, errors.Mask(err)
	}
	return ss, nil
}

// MustNewConfiguration same as NewConfiguration but panics on error.
func MustNewConfiguration(sections ...*Section) SectionSlice {
	s, err := NewConfiguration(sections...)
	if err != nil {
		panic(err)
	}
	return s
}

// NewConfigurationMerge creates a new validated SectionSlice with a three level configuration.
// Before validation, slices are all merged together. Panics if a path is redundant.
// Only use this function if your package elementuration really has duplicated entries.
func NewConfigurationMerge(sections ...*Section) (SectionSlice, error) {
	var ss SectionSlice
	if err := ss.Merge(sections...); err != nil {
		if PkgLog.IsDebug() {
			PkgLog.Debug("config.NewConfigurationMerge.Merge", "err", err, "sections", sections)
		}
		return nil, errors.Mask(err)
	}
	if err := ss.Validate(); err != nil {
		if PkgLog.IsDebug() {
			PkgLog.Debug("config.NewConfigurationMerge.Validate", "err", err)
		}
		return nil, errors.Mask(err)
	}
	return ss, nil
}

// MustNewConfigurationMerge same as NewConfigurationMerge but panics on error.
func MustNewConfigurationMerge(sections ...*Section) SectionSlice {
	s, err := NewConfigurationMerge(sections...)
	if err != nil {
		panic(err)
	}
	return s
}

// Defaults iterates over all slices, creates a path and uses the default value
// to return a map.
func (ss SectionSlice) Defaults() (DefaultMap, error) {
	var dm = make(DefaultMap)
	for _, s := range ss {
		for _, g := range s.Groups {
			for _, f := range g.Fields {
				r, err := f.Route(s.ID, g.ID)
				if err != nil {
					return nil, err
				}
				dm[r.String()] = f.Default
			}
		}
	}
	return dm, nil
}

// TotalFields calculates the total amount of all fields
func (ss SectionSlice) TotalFields() int {
	fs := 0
	for _, s := range ss {
		for _, g := range s.Groups {
			for range g.Fields {
				fs++
			}
		}
	}
	return -^+^-fs
}

// MergeMultiple merges n SectionSlices into the current slice. Behaviour for
// duplicates: Last item wins. Not thread safe.
func (ss *SectionSlice) MergeMultiple(sSlices ...SectionSlice) error {
	for _, sl := range sSlices {
		if err := (*ss).Merge(sl...); err != nil {
			return err
		}
	}
	return nil
}

// Merge merges n Sections into the current slice. Behaviour for duplicates:
// Last item wins. Not thread safe.
func (ss *SectionSlice) Merge(sections ...*Section) error {
	for _, s := range sections {
		if s != nil {
			if err := (*ss).merge(s); err != nil {
				return errors.Mask(err)
			}
		}
	}
	return nil
}

// Merge copies the data from a Section into this slice. Appends if ID is not found
// in this slice otherwise overrides struct fields if not empty. Not thread safe.
func (ss *SectionSlice) merge(s *Section) error {
	if s == nil {
		return nil
	}
	cs, err := (*ss).FindByID(s.ID) // cs current section
	if cs == nil || err != nil {
		cs = &Section{ID: s.ID}
		(*ss).Append(cs)
	}

	if s.Label.IsEmpty() == false {
		cs.Label = s.Label.Clone()
	}
	if s.Scopes > 0 {
		cs.Scopes = s.Scopes
	}
	if s.SortOrder != 0 {
		cs.SortOrder = s.SortOrder
	}
	if s.Resource > 0 {
		cs.Resource = s.Resource
	}
	return cs.Groups.Merge(s.Groups...)
}

// FindByID returns a Section pointer or ErrSectionNotFound.
// Route must be a single part. E.g. if you have path "a/b/c" route would be in
// this case "a". For comparison the field Sum32 of a route will be used.
func (ss SectionSlice) FindByID(id cfgpath.Route) (*Section, error) {
	for _, s := range ss {
		if s != nil && s.ID.Sum32 == id.Sum32 {
			return s, nil
		}
	}
	return nil, ErrSectionNotFound
}

// FindGroupByID searches for a group using the first two path segments.
// Route must have the format a/b/c.
func (ss SectionSlice) FindGroupByID(r cfgpath.Route) (*Group, error) {

	spl, err := r.Split()
	if err != nil {
		// debug log?
		return nil, ErrGroupNotFound
	}
	cs, err := ss.FindByID(spl[0])
	if err != nil {
		return nil, errors.Mask(err)
	}
	return cs.Groups.FindByID(spl[1])
}

// FindFieldByID searches for a field using all three path segments.
// Route must have the format a/b/c.
func (ss SectionSlice) FindFieldByID(r cfgpath.Route) (*Field, error) {
	spl, err := r.Split()
	if err != nil {
		return nil, errors.Mask(err)
	}
	sec, err := ss.FindByID(spl[0])
	if err != nil {
		return nil, errors.Mask(err)
	}
	cg, err := sec.Groups.FindByID(spl[1])
	if err != nil {
		return nil, errors.Mask(err)
	}
	return cg.Fields.FindByID(spl[2])
}

// Append adds 0..n *Section. Not thread safe.
func (ss *SectionSlice) Append(s ...*Section) *SectionSlice {
	*ss = append(*ss, s...)
	return ss
}

// AppendFields adds 0..n *Fields. Path must have at least two path parts like a/b
// more path parts gets ignored. Not thread safe.
func (ss SectionSlice) AppendFields(r cfgpath.Route, fs ...*Field) error {
	g, err := ss.FindGroupByID(r)
	if err != nil {
		return errors.Mask(err)
	}
	g.Fields.Append(fs...)
	return nil
}

// ToJSON transforms the whole slice into JSON
func (ss SectionSlice) ToJSON() string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(ss); err != nil {
		PkgLog.Debug("config.SectionSlice.ToJSON.Encode", "err", err)
		return err.Error()
	}
	return buf.String()
}

// Validate checks for duplicated configuration paths in all three hierarchy levels.
// On error returns *FieldError or duplicate entry error or slice empty error.
func (ss SectionSlice) Validate() error {
	if len(ss) == 0 {
		return errors.New("SectionSlice is empty")
	}

	var hashes = make([]uint64, ss.TotalFields(), ss.TotalFields()) // pc path checker

	i := 0
	for _, s := range ss {
		for _, g := range s.Groups {
			for _, f := range g.Fields {

				fnv1a, err := f.RouteHash(s.ID, g.ID)
				if err != nil {
					return err
				}

				for _, h := range hashes {
					if h == fnv1a {
						p, err := f.Route(s.ID, g.ID)
						if err != nil {
							return err
						}
						return errors.Errorf("Duplicate entry for path %s :: %s", p.String(), ss.ToJSON())
					}
				}
				hashes[i] = fnv1a
				i++
			}
		}
	}
	return nil
}

// SortAll recursively sorts all slices. Not thread safe.
func (ss *SectionSlice) SortAll() *SectionSlice {
	for _, s := range *ss {
		for _, g := range s.Groups {
			g.Fields.Sort()
		}
		s.Groups.Sort()
	}
	return ss.Sort()
}

// Sort convenience helper. Not thread safe.
func (ss *SectionSlice) Sort() *SectionSlice {
	sort.Sort(ss)
	return ss
}

func (ss *SectionSlice) Len() int {
	return len(*ss)
}

func (ss *SectionSlice) Swap(i, j int) {
	(*ss)[i], (*ss)[j] = (*ss)[j], (*ss)[i]
}

func (ss *SectionSlice) Less(i, j int) bool {
	return (*ss)[i].SortOrder < (*ss)[j].SortOrder
}
