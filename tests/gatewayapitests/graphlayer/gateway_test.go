/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package graphlayer

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"

	akogatewayapik8s "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/k8s"
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	tests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

var ctrl *akogatewayapik8s.GatewayController

const (
	DEFAULT_NAMESPACE = "default"
)

func TestMain(m *testing.M) {
	tests.KubeClient = k8sfake.NewSimpleClientset()
	tests.GatewayClient = gatewayfake.NewSimpleClientset()
	testData := tests.GetL7RuleFakeData()
	tests.DynamicClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), tests.GvrToKind, &testData)
	integrationtest.KubeClient = tests.KubeClient

	// Sets the environment variables
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("FULL_SYNC_INTERVAL", utils.AKO_DEFAULT_NS)
	os.Setenv("ENABLE_EVH", "true")
	os.Setenv("TENANT", "admin")
	os.Setenv("POD_NAME", "ako-0")

	// Set the user with prefix
	_ = lib.AKOControlConfig()
	lib.SetAKOUser(akogatewayapilib.Prefix)
	lib.SetNamePrefix(akogatewayapilib.Prefix)
	lib.AKOControlConfig().SetIsLeaderFlag(true)
	akoControlConfig := akogatewayapilib.AKOControlConfig()
	akoControlConfig.SetEventRecorder(lib.AKOGatewayEventComponent, tests.KubeClient, true)
	akogatewayapilib.SetDynamicClientSet(tests.DynamicClient)
	akogatewayapilib.NewDynamicInformers(tests.DynamicClient, false)
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.SecretInformer,
		utils.NSInformer,
	}

	registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)

	utils.AviLog.SetLevel("DEBUG")
	utils.NewInformers(utils.KubeClientIntf{ClientSet: tests.KubeClient}, registeredInformers, make(map[string]interface{}))
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	tests.KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	akoApi := integrationtest.InitializeFakeAKOAPIServer()
	defer akoApi.ShutDown()

	tests.NewAviFakeClientInstance(tests.KubeClient)
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = akogatewayapik8s.SharedGatewayController()
	ctrl.DisableSync = false
	ctrl.InitGatewayAPIInformers(tests.GatewayClient)
	akoControlConfig.SetGatewayAPIClientset(tests.GatewayClient)

	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})

	waitGroupMap := make(map[string]*sync.WaitGroup)
	wgIngestion := &sync.WaitGroup{}
	waitGroupMap["ingestion"] = wgIngestion
	wgFastRetry := &sync.WaitGroup{}
	waitGroupMap["fastretry"] = wgFastRetry
	wgSlowRetry := &sync.WaitGroup{}
	waitGroupMap["slowretry"] = wgSlowRetry
	wgGraph := &sync.WaitGroup{}
	waitGroupMap["graph"] = wgGraph
	wgStatus := &sync.WaitGroup{}
	waitGroupMap["status"] = wgStatus

	integrationtest.AddConfigMap(tests.KubeClient)
	go ctrl.InitController(k8s.K8sinformers{Cs: tests.KubeClient, DynamicClient: tests.DynamicClient}, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

/*
Positive Case
Create Gateway 1 listener (noTLS)
Create Gateway 1 listener (TLS)
*/
func TestGateway(t *testing.T) {

	gatewayName := "gateway-01"
	gatewayClassName := "gateway-class-01"
	ports := []int32{8080}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))
	// default backend response
	g.Expect(nodes[0].HttpPolicyRefs[0].RequestRules[0].Match.Path.MatchStr[0]).To(gomega.Equal("/"))
	g.Expect(*nodes[0].HttpPolicyRefs[0].RequestRules[0].SwitchingAction.StatusCode).To(gomega.Equal("HTTP_LOCAL_RESPONSE_STATUS_CODE_404"))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayWithTLS(t *testing.T) {

	gatewayName := "gateway-02"
	gatewayClassName := "gateway-class-02"
	ports := []int32{8080}

	secrets := []string{"secret-01"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false, secrets...)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[0].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
}

/*
Positive Case
Transition Gateway 1 listener (noTLS -> TLS)
Transition Gateway 1 listener (TLS -> noTLS)
*/

func TestGatewayNoTLSToTLS(t *testing.T) {

	gatewayName := "gateway-03"
	gatewayClassName := "gateway-class-03"
	ports := []int32{8080}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))

	//var tlsMode gatewayv1beta1.TLSModeType
	tlsMode := gatewayv1.TLSModeTerminate
	secrets := []string{"secret-02"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}
	certRefs := []gatewayv1.SecretObjectReference{{Name: gatewayv1.ObjectName(secrets[0])}}
	listeners[0].TLS = &gatewayv1.GatewayTLSConfig{
		Mode:            &tlsMode,
		CertificateRefs: certRefs,
	}
	tests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[0].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
}

func TestGatewayTLSToNoTLS(t *testing.T) {

	gatewayName := "gateway-04"
	gatewayClassName := "gateway-class-04"
	ports := []int32{8080}

	secrets := []string{"secret-03"}
	for _, secret := range secrets {
		integrationtest.AddSecret(secret, DEFAULT_NAMESPACE, "cert", "key")
	}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false, secrets...)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[0].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))

	listeners[0].TLS = nil
	tests.UpdateGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].PortProto[0].EnableSSL
	}, 40*time.Second).Should(gomega.Equal(false))

	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(0))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)
}

/*
Negative Case
Delete Gateway 1 listener
*/

func TestGatewayDelete(t *testing.T) {

	gatewayName := "gateway-05"
	gatewayClassName := "gateway-class-05"
	ports := []int32{8080}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		found, gwModel := objects.SharedAviGraphLister().Get(modelName)
		if found {
			return gwModel == nil
		}
		return true
	}, 25*time.Second).Should(gomega.Equal(true))

	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestSecretCreateDelete(t *testing.T) {

	gatewayName := "gateway-06"
	gatewayClassName := "gateway-class-06"
	ports := []int32{8080}
	secrets := []string{"secret-06"}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false, secrets...)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))
	// add delay
	time.Sleep(1 * time.Second)
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			g.Expect(nodes).To(gomega.HaveLen(1))
			g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
			return true
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	integrationtest.UpdateSecret(secrets[0], DEFAULT_NAMESPACE, "certnew", "keynew")

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			g.Expect(nodes).To(gomega.HaveLen(1))
			g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
			g.Expect(string(nodes[0].SSLKeyCertRefs[0].Cert)).To(gomega.Equal("certnew"))
			g.Expect(string(nodes[0].SSLKeyCertRefs[0].Key)).To(gomega.Equal("keynew"))
			return true
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestSecretCreateDeleteWithEmptyHostname(t *testing.T) {

	gatewayName := "gateway-07"
	gatewayClassName := "gateway-class-07"
	ports := []int32{8080}
	secrets := []string{"secret-07"}

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	// listener with empty hostname
	listeners := tests.GetListenersV1(ports, true, false, secrets...)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))
	// add delay
	time.Sleep(1 * time.Second)
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			g.Expect(nodes).To(gomega.HaveLen(1))
			g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
			return true
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewaywithDuplicateCertificateRefsinListener(t *testing.T) {
	/* This testcase will test one gateway having multiple duplicate certRefs in one listener
	1. Create secret
	2. Create gateway
	3. Delete secret
	4. Create secret
	5. Delete secret
	6. Delete gateway
	*/
	gatewayName := "gateway-08"
	gatewayClassName := "gateway-class-08"
	ports := []int32{8080}

	secrets := []string{"secret-08", "secret-08"}
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false, secrets...)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[0].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))

	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	//create
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			g.Expect(nodes).To(gomega.HaveLen(1))
			g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
			return true
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}
func TestMultipleGatewaywithSameCertificateRefsinListener(t *testing.T) {
	/* This testcase will test two gateways having same certRefs in their listener
	1. Create secret
	2. Create gateway1 and gateway2
	3. Delete secret
	4. Create secret
	5. Delete secret
	6. Delete gateway
	*/
	gatewayName1 := "gateway-09-a"
	gatewayName2 := "gateway-09-b"
	gatewayClassName := "gateway-class-09"
	ports1 := []int32{8080}
	ports2 := []int32{8081}

	secrets := []string{"secret-09"}
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners1 := tests.GetListenersV1(ports1, false, false, secrets...)
	tests.SetupGateway(t, gatewayName1, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners1)
	listeners2 := tests.GetListenersV1(ports2, false, false, secrets...)
	tests.SetupGateway(t, gatewayName2, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners2)

	g := gomega.NewGomegaWithT(t)
	modelName1 := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName1))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName1)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName1, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[0].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))

	modelName2 := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName2))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName2)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName2, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName2)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[0].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))

	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName1)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName2)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	//create
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName1)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			g.Expect(nodes).To(gomega.HaveLen(1))
			g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
			return true
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName2)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			g.Expect(nodes).To(gomega.HaveLen(1))
			g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
			return true
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName1)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName2)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	tests.TeardownGateway(t, gatewayName1, DEFAULT_NAMESPACE)
	tests.TeardownGateway(t, gatewayName2, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}
func TestGatewaywithSameCertificateRefsinMultipleListeners(t *testing.T) {
	/* This testcase will test one gateway having same certRefs in multiple listeners
	1. Create secret
	2. Create gateway
	3. Delete secret
	4. Create secret
	5. Delete secret
	6. Delete gateway
	*/
	gatewayName := "gateway-10"
	gatewayClassName := "gateway-class-10"
	ports := []int32{8080, 8081}

	secrets := []string{"secret-10"}
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, false, secrets...)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	g.Expect(nodes[0].PortProto[0].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[1].EnableSSL).To(gomega.Equal(true))

	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	//create
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			g.Expect(nodes).To(gomega.HaveLen(1))
			g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
			return true
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}
func TestGatewaywithSameCertificateRefsinMultipleListenersWithSameHostname(t *testing.T) {
	/* This testcase will test one gateway having same certRefs in multiple listeners sharing same hostname.
	1. Create secret
	2. Create gateway
	3. Delete secret
	4. Create secret
	5. Delete secret
	6. Delete gateway
	*/
	gatewayName := "gateway-11"
	gatewayClassName := "gateway-class-11"
	ports := []int32{8080, 8081}

	secrets := []string{"secret-11"}
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")

	tests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)
	listeners := tests.GetListenersV1(ports, false, true, secrets...)
	tests.SetupGateway(t, gatewayName, DEFAULT_NAMESPACE, gatewayClassName, nil, listeners)

	g := gomega.NewGomegaWithT(t)
	modelName := lib.GetModelName(lib.GetTenant(), akogatewayapilib.GetGatewayParentName(DEFAULT_NAMESPACE, gatewayName))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 25*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		gateway, err := tests.GatewayClient.GatewayV1().Gateways(DEFAULT_NAMESPACE).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if err != nil || gateway == nil {
			t.Logf("Couldn't get the gateway, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)) != nil
	}, 30*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	g.Expect(nodes[0].PortProto[0].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].PortProto[1].EnableSSL).To(gomega.Equal(true))
	g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))

	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	//create
	integrationtest.AddSecret(secrets[0], DEFAULT_NAMESPACE, "cert", "key")

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			g.Expect(nodes).To(gomega.HaveLen(1))
			g.Expect(nodes[0].SSLKeyCertRefs).To(gomega.HaveLen(1))
			return true
		}
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	// delete
	integrationtest.DeleteSecret(secrets[0], DEFAULT_NAMESPACE)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel == nil

	}, 30*time.Second).Should(gomega.Equal(true))

	tests.TeardownGateway(t, gatewayName, DEFAULT_NAMESPACE)
	tests.TeardownGatewayClass(t, gatewayClassName)
}
