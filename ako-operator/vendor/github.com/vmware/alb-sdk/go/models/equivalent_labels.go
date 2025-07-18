// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// EquivalentLabels equivalent labels
// swagger:model EquivalentLabels
type EquivalentLabels struct {

	// Equivalent labels. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Labels []string `json:"labels,omitempty"`
}
