// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IcapProfile icap profile
// swagger:model IcapProfile
type IcapProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Allow ICAP server to send 204 response as described in RFC 3507 section 4.5. Service Engine will buffer the complete request if alllow_204 is enabled. If disabled, preview_size request body will be buffered if enable_preview is set to true, and rest of the request body will be streamed to the ICAP server. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Allow204 *bool `json:"allow_204,omitempty"`

	// The maximum buffer size for the HTTP request body. If the request body exceeds this size, the request will not be checked by the ICAP server. In this case, the configured action will be executed and a significant log entry will be generated. Allowed values are 1-51200. Field introduced in 20.1.1. Unit is KB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BufferSize *uint32 `json:"buffer_size,omitempty"`

	// Decide what should happen if the request body size exceeds the configured buffer size. If this is set to Fail Open, the request will not be checked by the ICAP server. If this is set to Fail Closed, the request will be rejected with 413 status code. Enum options - ICAP_FAIL_OPEN, ICAP_FAIL_CLOSED. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BufferSizeExceedAction *string `json:"buffer_size_exceed_action,omitempty"`

	// The cloud where this object belongs to. This must match the cloud referenced in the pool group below. It is a reference to an object of type Cloud. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// A description for this ICAP profile. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Use the ICAP preview feature as described in RFC 3507 section 4.5. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnablePreview *bool `json:"enable_preview,omitempty"`

	// Decide what should happen if there is a problem with the ICAP server like communication timeout, protocol error, pool error, etc. If the ICAP server responds with 4xx-5xx error code the configured fail action is performed. If this is set to Fail Open, the request will continue, but will create a significant log entry. If this is set to Fail Closed, the request will be rejected with a 500 status code. Enum options - ICAP_FAIL_OPEN, ICAP_FAIL_CLOSED. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FailAction *string `json:"fail_action,omitempty"`

	// Name of the ICAP profile. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// NSXDefender specific ICAP configurations. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NsxDefenderConfig *IcapNsxDefenderConfig `json:"nsx_defender_config,omitempty"`

	// The pool group which is used to connect to ICAP servers. It is a reference to an object of type PoolGroup. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	PoolGroupRef *string `json:"pool_group_ref"`

	// The ICAP preview size as described in RFC 3507 section 4.5. This should not exceed the size supported by the ICAP server. If this is set to 0, only the HTTP header will be sent to the ICAP server as a preview. To disable preview completely, set the enable-preview option to false.If vendor is LASTLINE, recommended preview size is 1000 bytes,minimum preview size is 10 bytes. Allowed values are 0-5000. Field introduced in 20.1.1. Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreviewSize *uint32 `json:"preview_size,omitempty"`

	// Maximum time, client's request will be paused for ICAP processing. If this timeout is exceeded, the request to the ICAP server will be aborted and the configured fail action is executed. Allowed values are 50-3600000. Field introduced in 20.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResponseTimeout *uint32 `json:"response_timeout,omitempty"`

	// The path and query component of the ICAP URL. Host name and port will be taken from the pool. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServiceURI *string `json:"service_uri"`

	// If the ICAP request takes longer than this value, this request will generate a significant log entry. Allowed values are 50-3600000. Field introduced in 20.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SlowResponseWarningThreshold *uint32 `json:"slow_response_warning_threshold,omitempty"`

	// Tenant which this object belongs to. It is a reference to an object of type Tenant. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the ICAP profile. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// The vendor of the ICAP server. Enum options - ICAP_VENDOR_GENERIC, ICAP_VENDOR_OPSWAT, ICAP_VENDOR_LASTLINE. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vendor *string `json:"vendor,omitempty"`
}
