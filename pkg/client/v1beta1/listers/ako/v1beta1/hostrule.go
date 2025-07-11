/*
Copyright The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// HostRuleLister helps list HostRules.
// All objects returned here must be treated as read-only.
type HostRuleLister interface {
	// List lists all HostRules in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*akov1beta1.HostRule, err error)
	// HostRules returns an object that can list and get HostRules.
	HostRules(namespace string) HostRuleNamespaceLister
	HostRuleListerExpansion
}

// hostRuleLister implements the HostRuleLister interface.
type hostRuleLister struct {
	listers.ResourceIndexer[*akov1beta1.HostRule]
}

// NewHostRuleLister returns a new HostRuleLister.
func NewHostRuleLister(indexer cache.Indexer) HostRuleLister {
	return &hostRuleLister{listers.New[*akov1beta1.HostRule](indexer, akov1beta1.Resource("hostrule"))}
}

// HostRules returns an object that can list and get HostRules.
func (s *hostRuleLister) HostRules(namespace string) HostRuleNamespaceLister {
	return hostRuleNamespaceLister{listers.NewNamespaced[*akov1beta1.HostRule](s.ResourceIndexer, namespace)}
}

// HostRuleNamespaceLister helps list and get HostRules.
// All objects returned here must be treated as read-only.
type HostRuleNamespaceLister interface {
	// List lists all HostRules in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*akov1beta1.HostRule, err error)
	// Get retrieves the HostRule from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*akov1beta1.HostRule, error)
	HostRuleNamespaceListerExpansion
}

// hostRuleNamespaceLister implements the HostRuleNamespaceLister
// interface.
type hostRuleNamespaceLister struct {
	listers.ResourceIndexer[*akov1beta1.HostRule]
}
