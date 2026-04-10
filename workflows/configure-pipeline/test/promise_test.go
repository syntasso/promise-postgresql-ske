package test_test

import (
	"context"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	crdTimeout        = 120 * time.Second
	deploymentTimeout = 120 * time.Second
	podReadyTimeout   = 300 * time.Second
	healthTimeout     = 300 * time.Second
	pollInterval      = 5 * time.Second
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func newClientSet(contextName string) (*kubernetes.Clientset, error) {
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: contextName},
	)
	rest, err := cfg.ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(rest)
}

func newAPIExtensionsClient(contextName string) (*apiextensionsclient.Clientset, error) {
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: contextName},
	)
	rest, err := cfg.ClientConfig()
	if err != nil {
		return nil, err
	}
	return apiextensionsclient.NewForConfig(rest)
}

func newDynamicClient(contextName string) (dynamic.Interface, error) {
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: contextName},
	)
	rest, err := cfg.ClientConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(rest)
}

func kubectlApply(contextName, path string) error {
	cmd := exec.Command("kubectl", "--context="+contextName, "apply", "-f", path)
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	return cmd.Run()
}

var _ = Describe("PostgreSQL Promise", Ordered, func() {
	var (
		workerCS    *kubernetes.Clientset
		workerExtCS *apiextensionsclient.Clientset
		workerDyn   dynamic.Interface
		platformDyn dynamic.Interface
		ctx         context.Context
	)

	workerCtx := getEnv("WORKER_CONTEXT", "kind-worker")
	platformCtx := getEnv("PLATFORM_CONTEXT", "kind-platform")
	promiseYAML := getEnv("PROMISE_YAML", "../../../promise.yaml")
	resourceRequestYAML := getEnv("RESOURCE_REQUEST_YAML", "../../../resource-request.yaml")

	BeforeAll(func() {
		ctx = context.Background()

		var err error
		workerCS, err = newClientSet(workerCtx)
		Expect(err).NotTo(HaveOccurred())

		workerExtCS, err = newAPIExtensionsClient(workerCtx)
		Expect(err).NotTo(HaveOccurred())

		workerDyn, err = newDynamicClient(workerCtx)
		Expect(err).NotTo(HaveOccurred())

		platformDyn, err = newDynamicClient(platformCtx)
		Expect(err).NotTo(HaveOccurred())

		By("Applying promise.yaml")
		Expect(kubectlApply(platformCtx, promiseYAML)).To(Succeed())
	})

	Context("Promise installation", func() {
		It("propagates the acid.zalan.do CRD to the worker cluster", func() {
			Eventually(func(g Gomega) {
				crd, err := workerExtCS.ApiextensionsV1().CustomResourceDefinitions().Get(
					ctx, "postgresqls.acid.zalan.do", metav1.GetOptions{},
				)
				g.Expect(err).NotTo(HaveOccurred())
				var established bool
				for _, cond := range crd.Status.Conditions {
					if cond.Type == apiextensionsv1.Established && cond.Status == apiextensionsv1.ConditionTrue {
						established = true
					}
				}
				g.Expect(established).To(BeTrue())
			}).WithTimeout(crdTimeout).WithPolling(pollInterval).Should(Succeed())
		})

		It("deploys the postgres-operator on the worker cluster", func() {
			Eventually(func(g Gomega) {
				deploy, err := workerCS.AppsV1().Deployments("default").Get(
					ctx, "postgres-operator", metav1.GetOptions{},
				)
				g.Expect(err).NotTo(HaveOccurred())
				var available bool
				for _, cond := range deploy.Status.Conditions {
					if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
						available = true
					}
				}
				g.Expect(available).To(BeTrue())
			}).WithTimeout(deploymentTimeout).WithPolling(pollInterval).Should(Succeed())
		})
	})

	Context("Resource request provisioning", func() {
		BeforeAll(func() {
			By("Applying resource-request.yaml")
			Expect(kubectlApply(platformCtx, resourceRequestYAML)).To(Succeed())
		})

		It("creates the postgresql resource on the worker cluster", func() {
			postgresqlGVR := schema.GroupVersionResource{
				Group:    "acid.zalan.do",
				Version:  "v1",
				Resource: "postgresqls",
			}
			Eventually(func(g Gomega) {
				_, err := workerDyn.Resource(postgresqlGVR).Namespace("default").Get(
					ctx, "acme-org-team-a-example-postgresql", metav1.GetOptions{},
				)
				g.Expect(err).NotTo(HaveOccurred())
			}).WithTimeout(crdTimeout).WithPolling(pollInterval).Should(Succeed())
		})

		It("brings the spilo master pod to Ready", func() {
			Eventually(func(g Gomega) {
				pods, err := workerCS.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
					LabelSelector: "spilo-role=master",
				})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pods.Items).NotTo(BeEmpty())
				var ready bool
				for _, cond := range pods.Items[0].Status.Conditions {
					if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
						ready = true
					}
				}
				g.Expect(ready).To(BeTrue())
			}).WithTimeout(podReadyTimeout).WithPolling(pollInterval).Should(Succeed())
		})

		It("reports a healthy Health Status on the resource", func() {
			skePostgresqlGVR := schema.GroupVersionResource{
				Group:    "marketplace.kratix.io",
				Version:  "v1alpha2",
				Resource: "ske-postgresqls",
			}
			Eventually(func(g Gomega) {
				obj, err := platformDyn.Resource(skePostgresqlGVR).Namespace("default").Get(
					ctx, "example", metav1.GetOptions{},
				)
				g.Expect(err).NotTo(HaveOccurred())
				status, ok := obj.Object["status"].(map[string]any)
				g.Expect(ok).To(BeTrue(), "status field missing")
				healthStatus, ok := status["healthStatus"].(map[string]any)
				g.Expect(ok).To(BeTrue(), "healthStatus field missing")
				state, ok := healthStatus["state"].(string)
				g.Expect(ok).To(BeTrue(), "healthStatus.state field missing")
				g.Expect(state).To(Equal("healthy"))
			}).WithTimeout(healthTimeout).WithPolling(pollInterval).Should(Succeed())
		})
	})
})
