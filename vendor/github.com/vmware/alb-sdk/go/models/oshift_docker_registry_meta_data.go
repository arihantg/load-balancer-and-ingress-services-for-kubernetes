// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OshiftDockerRegistryMetaData oshift docker registry meta data
// swagger:model OshiftDockerRegistryMetaData
type OshiftDockerRegistryMetaData struct {

	// Namespace for the ServiceEngine image to be hosted in Openshift Integrated registry. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RegistryNamespace *string `json:"registry_namespace,omitempty"`

	// Name of the Integrated registry Service in Openshift. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RegistryService *string `json:"registry_service,omitempty"`

	// Static VIP for 'docker-registry' service in Openshift if Avi is proxying for this service.This VIP should be outside the cluster IP subnet in Kubernetes and within the subnet configured (but outside the available pool of IPs) in the East West IPAM profile configuration for this Cloud. For example, if kubernetes cluster VIP range is 172.30.0.0/16 and subnet configured in East West IPAM profile is 172.50.0.0/16, then 172.50.0.2 can be used for this vip and IP pool can start from 172.50.0.3 onwards. Use this static VIP in '--insecure-registry <this-vip> 5000' docker config if using an insecure registry or add this to the list of IPs/hostnames when generating certificates if using a secure TLS registry. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RegistryVip *IPAddr `json:"registry_vip,omitempty"`
}
