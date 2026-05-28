package client

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	fakeclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

func NewFakeClient(objects ...runtime.Object) *OVNKMCPServerClientSet {
	clientSet := fakeclient.NewClientset(objects...)
	fakeDiscovery := clientSet.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.Resources = getAPIResources()
	dynamicClient := fakedynamic.NewSimpleDynamicClient(scheme.Scheme, objects...)
	memoryClient := memory.NewMemCacheClient(fakeDiscovery)
	deferredDiscoveryRESTMapper := restmapper.NewDeferredDiscoveryRESTMapper(memoryClient)

	return &OVNKMCPServerClientSet{
		clientSet:                   clientSet,
		dynamicClient:               dynamicClient,
		deferredDiscoveryRESTMapper: deferredDiscoveryRESTMapper,
		config:                      &rest.Config{},
		debugPodNamespace:           "default",
	}
}

func getAPIResources() []*metav1.APIResourceList {
	apiResourceList := []*metav1.APIResourceList{}
	allKnownTypes := scheme.Scheme.AllKnownTypes()
	gvResources := make(map[schema.GroupVersion]*metav1.APIResourceList)
	for gvk := range allKnownTypes {
		list, exists := gvResources[gvk.GroupVersion()]
		if !exists {
			list = &metav1.APIResourceList{
				GroupVersion: gvk.GroupVersion().String(),
				APIResources: []metav1.APIResource{},
			}
			gvResources[gvk.GroupVersion()] = list
			apiResourceList = append(apiResourceList, list)
		}
		gvr, _ := meta.UnsafeGuessKindToResource(gvk)
		list.APIResources = append(list.APIResources, metav1.APIResource{
			Name:    gvr.Resource,
			Kind:    gvk.Kind,
			Group:   gvk.Group,
			Version: gvk.Version,
		})
	}
	return apiResourceList
}
