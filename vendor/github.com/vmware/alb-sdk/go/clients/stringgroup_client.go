// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// StringGroupClient is a client for avi StringGroup resource
type StringGroupClient struct {
	aviSession *session.AviSession
}

// NewStringGroupClient creates a new client for StringGroup resource
func NewStringGroupClient(aviSession *session.AviSession) *StringGroupClient {
	return &StringGroupClient{aviSession: aviSession}
}

func (client *StringGroupClient) getAPIPath(uuid string) string {
	path := "api/stringgroup"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of StringGroup objects
func (client *StringGroupClient) GetAll(options ...session.ApiOptionsParams) ([]*models.StringGroup, error) {
	var plist []*models.StringGroup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing StringGroup by uuid
func (client *StringGroupClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.StringGroup, error) {
	var obj *models.StringGroup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing StringGroup by name
func (client *StringGroupClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.StringGroup, error) {
	var obj *models.StringGroup
	err := client.aviSession.GetObjectByName("stringgroup", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing StringGroup by filters like name, cloud, tenant
// Api creates StringGroup object with every call.
func (client *StringGroupClient) GetObject(options ...session.ApiOptionsParams) (*models.StringGroup, error) {
	var obj *models.StringGroup
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("stringgroup", newOptions...)
	return obj, err
}

// Create a new StringGroup object
func (client *StringGroupClient) Create(obj *models.StringGroup, options ...session.ApiOptionsParams) (*models.StringGroup, error) {
	var robj *models.StringGroup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing StringGroup object
func (client *StringGroupClient) Update(obj *models.StringGroup, options ...session.ApiOptionsParams) (*models.StringGroup, error) {
	var robj *models.StringGroup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing StringGroup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.StringGroup
// or it should be json compatible of form map[string]interface{}
func (client *StringGroupClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.StringGroup, error) {
	var robj *models.StringGroup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing StringGroup object with a given UUID
func (client *StringGroupClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing StringGroup object with a given name
func (client *StringGroupClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *StringGroupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
