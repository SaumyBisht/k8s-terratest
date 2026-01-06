package test

import (
	"fmt"
	"testing"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesHelloWorldDeployment(t *testing.T) {
	t.Parallel()

	kubectlOptions := k8s.NewKubectlOptions("", "", "default")

	deploymentPath := "../manifests/k8s-hello-world/deployment.yaml"
	servicePath := "../manifests/k8s-hello-world/service.yaml"

	// -------------------------------
	// Apply Kubernetes manifests
	// -------------------------------
	k8s.KubectlApply(t, kubectlOptions, deploymentPath)
	k8s.KubectlApply(t, kubectlOptions, servicePath)

	// Cleanup
	defer k8s.KubectlDelete(t, kubectlOptions, deploymentPath)
	defer k8s.KubectlDelete(t, kubectlOptions, servicePath)

	// --------------------------------
	// TEST CASE 1: Deployment exists
	// --------------------------------
	deployment := k8s.GetDeployment(
		t,
		kubectlOptions,
		"hello-world-deploymnet",
	)

	assert.Equal(t, int32(1), *deployment.Spec.Replicas)

	// Label Selector
	listOptions := metav1.ListOptions{
		LabelSelector: "app=hello-world",
	}

	// --------------------------------
	// Test Case 2:  Pod is created
	k8s.WaitUntilNumPodsCreated(
		t,
		kubectlOptions,
		listOptions,
		1,
		60,
		5*time.Second,
	)

	pods := k8s.ListPods(t, kubectlOptions, listOptions)
	assert.Equal(t, 1, len(pods))

	//--------------------------------
	// Test Case 3: Pod becomes Ready
	//--------------------------------
	k8s.WaitUntilPodAvailable(
		t,
		kubectlOptions,
		pods[0].Name,
		60,
		5*time.Second,
	)

	//-------------------------------
	// Test Case 4: Service Exists
	//-------------------------------
	service := k8s.GetService(
		t,
		kubectlOptions,
		"hello-world-service",
	)
	assert.Equal(t, "LoadBalancer", string(service.Spec.Type))
	assert.Equal(t, int32(5000), service.Spec.Ports[0].Port)

	//--------------------------------
	//  Test Case 5: HTTP endpoint works
	//---------------------------------

	// NodePort assigned by kubernetes
	nodePort := service.Spec.Ports[0].NodePort

	// Get Node Internal IP
	nodes := k8s.GetNodes(t, kubectlOptions)

	var nodeIP string

	for _, addr := range nodes[0].Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			nodeIP = addr.Address
			break
		}
	}

	assert.NotEmpty(t, nodeIP, "NodeInternalIP should not be empty")
	url := fmt.Sprintf("http://%s:%d", nodeIP, nodePort)

	http_helper.HttpGetWithRetry(
		t,
		url,
		nil,
		200,
		"hello world",
		30,
		3*time.Second,
	)
}
