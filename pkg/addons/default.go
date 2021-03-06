package addons

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/openshift/openshift-azure/pkg/jsonpath"
)

func defaultContainerSpec(obj map[string]interface{}) {
	jsonpath.MustCompile("$.livenessProbe.failureThreshold").DeleteIfMatch(obj, int64(3))
	jsonpath.MustCompile("$.livenessProbe.periodSeconds").DeleteIfMatch(obj, int64(10))
	jsonpath.MustCompile("$.livenessProbe.successThreshold").DeleteIfMatch(obj, int64(1))
	jsonpath.MustCompile("$.livenessProbe.timeoutSeconds").DeleteIfMatch(obj, int64(1))
	jsonpath.MustCompile("$.ports.*.protocol").DeleteIfMatch(obj, "TCP")
	jsonpath.MustCompile("$.readinessProbe.failureThreshold").DeleteIfMatch(obj, int64(3))
	jsonpath.MustCompile("$.readinessProbe.periodSeconds").DeleteIfMatch(obj, int64(10))
	jsonpath.MustCompile("$.readinessProbe.successThreshold").DeleteIfMatch(obj, int64(1))
	jsonpath.MustCompile("$.readinessProbe.timeoutSeconds").DeleteIfMatch(obj, int64(1))
	jsonpath.MustCompile("$.resources").DeleteIfMatch(obj, map[string]interface{}{})
	jsonpath.MustCompile("$.terminationMessagePath").DeleteIfMatch(obj, "/dev/termination-log")
	jsonpath.MustCompile("$.terminationMessagePolicy").DeleteIfMatch(obj, "File")
}

func defaultPodSpec(obj map[string]interface{}) {
	for _, c := range jsonpath.MustCompile("$.containers.*").Get(obj) {
		defaultContainerSpec(c.(map[string]interface{}))
	}
	for _, c := range jsonpath.MustCompile("$.initContainers.*").Get(obj) {
		defaultContainerSpec(c.(map[string]interface{}))
	}
	jsonpath.MustCompile("$.dnsPolicy").DeleteIfMatch(obj, "ClusterFirst")
	jsonpath.MustCompile("$.restartPolicy").DeleteIfMatch(obj, "Always")
	jsonpath.MustCompile("$.schedulerName").DeleteIfMatch(obj, "default-scheduler")
	jsonpath.MustCompile("$.securityContext").DeleteIfMatch(obj, map[string]interface{}{})
	jsonpath.MustCompile("$.serviceAccount").Delete(obj) // deprecated alias of serviceAccountName
	jsonpath.MustCompile("$.terminationGracePeriodSeconds").DeleteIfMatch(obj, int64(30))
	jsonpath.MustCompile("$.volumes.*.configMap.defaultMode").DeleteIfMatch(obj, int64(0644))
	jsonpath.MustCompile("$.volumes.*.hostPath.type").DeleteIfMatch(obj, "")
	jsonpath.MustCompile("$.volumes.*.secret.defaultMode").DeleteIfMatch(obj, int64(0644))
}

func Default(o unstructured.Unstructured) {
	gk := o.GroupVersionKind().GroupKind()

	switch gk.String() {
	case "DaemonSet.apps":
		jsonpath.MustCompile("$.spec.revisionHistoryLimit").DeleteIfMatch(o.Object, int64(10))

		for _, c := range jsonpath.MustCompile("$.spec.template.spec").Get(o.Object) {
			defaultPodSpec(c.(map[string]interface{}))
		}

		defaultUpdateStrategy := map[string]interface{}{
			"rollingUpdate": map[string]interface{}{
				"maxUnavailable": int64(1),
			},
			"type": "RollingUpdate",
		}
		jsonpath.MustCompile("$.spec.updateStrategy").DeleteIfMatch(o.Object, defaultUpdateStrategy)

	case "Deployment.apps":
		jsonpath.MustCompile("$.spec.progressDeadlineSeconds").DeleteIfMatch(o.Object, int64(600))
		jsonpath.MustCompile("$.spec.revisionHistoryLimit").DeleteIfMatch(o.Object, int64(10))

		defaultStrategy := map[string]interface{}{
			"rollingUpdate": map[string]interface{}{
				"maxSurge":       "25%",
				"maxUnavailable": "25%",
			},
			"type": "RollingUpdate",
		}
		jsonpath.MustCompile("$.spec.strategy").DeleteIfMatch(o.Object, defaultStrategy)

		for _, c := range jsonpath.MustCompile("$.spec.template.spec").Get(o.Object) {
			defaultPodSpec(c.(map[string]interface{}))
		}

	case "Secret":
		jsonpath.MustCompile("$.type").DeleteIfMatch(o.Object, "Opaque")

	case "Service":
		jsonpath.MustCompile("$.spec.ports.*.protocol").DeleteIfMatch(o.Object, "TCP")
		jsonpath.MustCompile("$.spec.sessionAffinity").DeleteIfMatch(o.Object, "None")
		jsonpath.MustCompile("$.spec.type").DeleteIfMatch(o.Object, "ClusterIP")

		for _, p := range jsonpath.MustCompile("$.spec.ports.*").Get(o.Object) {
			jsonpath.MustCompile("$.targetPort").DeleteIfMatch(p, jsonpath.MustCompile("$.port").Get(p)[0].(int64))
		}

	case "StatefulSet.apps":
		jsonpath.MustCompile("$.spec.revisionHistoryLimit").DeleteIfMatch(o.Object, int64(10))
		jsonpath.MustCompile("$.spec.serviceName").DeleteIfMatch(o.Object, "")

		for _, c := range jsonpath.MustCompile("$.spec.template.spec").Get(o.Object) {
			defaultPodSpec(c.(map[string]interface{}))
		}
	}
}
