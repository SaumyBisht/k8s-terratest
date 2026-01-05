package test

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesDeployment(t *testing.T) {
	t.Parallel()

	kubectlOptions := k8s.NewKubectlOptions("", "", "default")

	// Apply manifests
	k8s.KubectlApply(t, kubectlOptions, "../manifests/deployment.yaml")
	k8s.KubectlApply(t, kubectlOptions, "../manifests/service.yaml")

	// Cleanup
	defer k8s.KubectlDelete(t, kubectlOptions, "../manifests/deployment.yaml")
	defer k8s.KubectlDelete(t, kubectlOptions, "../manifests/service.yaml")

	// --------------------------------
	// TEST CASE 1: Deployment exists
	// --------------------------------
	deployment := k8s.GetDeployment(t, kubectlOptions, "nginx-deployment")
	assert.Equal(t, int32(2), *deployment.Spec.Replicas)

	// Label selector (THIS IS THE FIX)
	listOptions := metav1.ListOptions{
		LabelSelector: "app=nginx",
	}

	// --------------------------------
	// TEST CASE 2: Pods are created
	// --------------------------------
	k8s.WaitUntilNumPodsCreated(
		t,
		kubectlOptions,
		listOptions,
		2,
		60,
		5*time.Second,
	)

	pods := k8s.ListPods(t, kubectlOptions, listOptions)
	assert.Equal(t, 2, len(pods))

	// --------------------------------
	// TEST CASE 3: Pods become Ready
	// --------------------------------
	for _, pod := range pods {
		k8s.WaitUntilPodAvailable(
			t,
			kubectlOptions,
			pod.Name,
			60,
			5*time.Second,
		)
	}

	// --------------------------------
	// TEST CASE 4: Service exists
	// --------------------------------
	service := k8s.GetService(t, kubectlOptions, "nginx-service")
	assert.Equal(t, "NodePort", string(service.Spec.Type))
	assert.Equal(t, int32(80), service.Spec.Ports[0].Port)
}
