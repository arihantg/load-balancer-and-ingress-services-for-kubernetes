// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// VIMgrNWRuntimeClient is a client for avi VIMgrNWRuntime resource
type VIMgrNWRuntimeClient struct {
	aviSession *session.AviSession
}

// NewVIMgrNWRuntimeClient creates a new client for VIMgrNWRuntime resource
func NewVIMgrNWRuntimeClient(aviSession *session.AviSession) *VIMgrNWRuntimeClient {
	return &VIMgrNWRuntimeClient{aviSession: aviSession}
}

func (client *VIMgrNWRuntimeClient) getAPIPath(uuid string) string {
	path := "api/vimgrnwruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIMgrNWRuntime objects
func (client *VIMgrNWRuntimeClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VIMgrNWRuntime, error) {
	var plist []*models.VIMgrNWRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VIMgrNWRuntime by uuid
func (client *VIMgrNWRuntimeClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VIMgrNWRuntime, error) {
	var obj *models.VIMgrNWRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VIMgrNWRuntime by name
func (client *VIMgrNWRuntimeClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VIMgrNWRuntime, error) {
	var obj *models.VIMgrNWRuntime
	err := client.aviSession.GetObjectByName("vimgrnwruntime", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VIMgrNWRuntime by filters like name, cloud, tenant
// Api creates VIMgrNWRuntime object with every call.
func (client *VIMgrNWRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.VIMgrNWRuntime, error) {
	var obj *models.VIMgrNWRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vimgrnwruntime", newOptions...)
	return obj, err
}

// Create a new VIMgrNWRuntime object
func (client *VIMgrNWRuntimeClient) Create(obj *models.VIMgrNWRuntime, options ...session.ApiOptionsParams) (*models.VIMgrNWRuntime, error) {
	var robj *models.VIMgrNWRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VIMgrNWRuntime object
func (client *VIMgrNWRuntimeClient) Update(obj *models.VIMgrNWRuntime, options ...session.ApiOptionsParams) (*models.VIMgrNWRuntime, error) {
	var robj *models.VIMgrNWRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VIMgrNWRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIMgrNWRuntime
// or it should be json compatible of form map[string]interface{}
func (client *VIMgrNWRuntimeClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VIMgrNWRuntime, error) {
	var robj *models.VIMgrNWRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VIMgrNWRuntime object with a given UUID
func (client *VIMgrNWRuntimeClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VIMgrNWRuntime object with a given name
func (client *VIMgrNWRuntimeClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VIMgrNWRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
