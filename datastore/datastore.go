/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package datastore implements an interface on top of the mgo mongo client
package datastore

import (
	"errors"

	"gopkg.in/mgo.v2"
)

// Config configures the database connection
type Config struct {
	URL      string
	Database string
}

// Session is an interface for a MongoDB session
type Session interface {
	DB() (Database, func())
}

// Database is an interface for accessing a MongoDB database
type Database interface {
	C(name string) Collection
}

// Collection is an interface for accessing a MongoDB collection
type Collection interface {
	Find(query interface{}) Query
	FindId(id interface{}) Query
	Count() (n int, err error)
	Insert(docs ...interface{}) error
	Remove(selector interface{}) error
	UpdateId(id, update interface{}) error
	Upsert(selector, update interface{}) (*mgo.ChangeInfo, error)
	UpsertId(id, update interface{}) (*mgo.ChangeInfo, error)
}

// Query is an interface for querying a MongoDB collection
type Query interface {
	All(result interface{}) error
	One(result interface{}) error
}

// mgoSession wraps an mgo.Session and implements Session
type mgoSession struct {
	conf Config
	*mgo.Session
}

func (s *mgoSession) DB() (Database, func()) {
	copy := s.Session.Copy()
	closer := func() { copy.Close() }
	return &mgoDatabase{copy.DB(s.conf.Database)}, closer
}

// mgoDatabase wraps an mgo.Database and implements Database
type mgoDatabase struct {
	*mgo.Database
}

func (d *mgoDatabase) C(name string) Collection {
	return &mgoCollection{d.Database.C(name)}
}

// mgoCollection wraps an mgo.Collection and implements Collection
type mgoCollection struct {
	*mgo.Collection
}

func (c *mgoCollection) Find(query interface{}) Query {
	return c.Collection.Find(query)
}

func (c *mgoCollection) FindId(id interface{}) Query {
	return c.Collection.FindId(id)
}

// NewSession initializes a MongoDB connection to the given host
func NewSession(conf Config) (Session, error) {
	session, err := mgo.Dial(conf.URL)
	if err != nil {
		return nil, errors.New("unable to connect to MongoDB")
	}
	session.SetMode(mgo.Monotonic, true)
	return &mgoSession{conf, session}, nil
}
