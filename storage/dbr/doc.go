// Copyright 2015-2017, Cyrill @ Schumacher.fm and the CoreStore contributors
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

// Package dbr has additions to Go's database/sql for super fast performance and
// type safety and convenience.
//
// Aim: Allow a developer to easily modify a SQL query without type assertion of
// parts of the query. This package gets extended during csfw development.
//
// Abbreviations
//
// DML (https://en.wikipedia.org/wiki/Data_manipulation_language) Select,
// Insert, Update and Delete.
//
// DDL (https://en.wikipedia.org/wiki/Data_definition_language) Create, Drop,
// Alter, and Rename.
//
// DCL (https://en.wikipedia.org/wiki/Data_control_language) Grant and Revoke.
//
// CRUD (https://en.wikipedia.org/wiki/Create,_read,_update_and_delete) Create,
// Read, Update and Delete.
//
// https://mariadb.com/kb/en/mariadb/documentation/
//
// TODO(CyS) Add named parameter from GO1.8 to each query builder
package dbr
