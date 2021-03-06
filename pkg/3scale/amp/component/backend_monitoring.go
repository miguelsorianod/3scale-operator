package component

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/assets"
	"github.com/3scale/3scale-operator/pkg/common"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (backend *Backend) BackendListenerPodMonitor() *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-listener",
			Labels: backend.Options.CommonListenerLabels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{{
				Port:   "metrics",
				Path:   "/metrics",
				Scheme: "http",
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: backend.Options.CommonListenerLabels,
			},
		},
	}
}

func (backend *Backend) BackendWorkerPodMonitor() *monitoringv1.PodMonitor {
	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "backend-worker",
			Labels: backend.Options.CommonWorkerLabels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{{
				Port:   "metrics",
				Path:   "/metrics",
				Scheme: "http",
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: backend.Options.CommonWorkerLabels,
			},
		},
	}
}

func BackendGrafanaDashboard(ns string) *grafanav1alpha1.GrafanaDashboard {
	data := &struct {
		Namespace string
	}{
		ns,
	}
	return &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend",
			Labels: map[string]string{
				"monitoring-key": common.MonitoringKey,
			},
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: assets.TemplateAsset("monitoring/backend-grafana-dashboard-1.json.tpl", data),
			Name: fmt.Sprintf("%s/backend-grafana-dashboard-1.json", ns),
		},
	}
}

func BackendWorkerPrometheusRules(ns string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-worker",
			Labels: map[string]string{
				"prometheus": "application-monitoring",
				"role":       "alert-rules",
			},
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/backend-worker.rules", ns),
					Rules: []monitoringv1.Rule{
						{
							Alert: "ThreescaleBackendWorkerJobsCountRunningHigh",
							Annotations: map[string]string{
								"summary":     "{{$labels.container_name}} replica controller on {{$labels.namespace}}: Has more than 10000 jobs processed in the last 5 minutes",
								"description": "{{$labels.container_name}} replica controller on {{$labels.namespace}} project: Has more than 1000 jobs processed in the last 5 minutes",
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum(avg_over_time(apisonator_worker_job_count{job=~"backend.*",namespace="%s"} [5m])) by (namespace,job) > 10000`, ns)),
							For:  "5m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
						{
							Alert: "ThreescaleBackendWorkerJobDown",
							Annotations: map[string]string{
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
							},
							Expr: intstr.FromString(fmt.Sprintf(`up{job=~".*backend-worker.*",namespace="%s"} == 0`, ns)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
					},
				},
			},
		},
	}
}

func BackendListenerPrometheusRules(ns string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-listener",
			Labels: map[string]string{
				"prometheus": "application-monitoring",
				"role":       "alert-rules",
			},
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: fmt.Sprintf("%s/backend-listener.rules", ns),
					Rules: []monitoringv1.Rule{
						{
							Alert: "ThreescaleBackendListener5XXRequestsHigh",
							Annotations: map[string]string{
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} has more than 5000 HTTP 5xx requests in the last 5 minutes",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} has more than 5000 HTTP 5xx requests in the last 5 minutes",
							},
							Expr: intstr.FromString(fmt.Sprintf(`sum(rate(apisonator_listener_response_codes{job=~"backend.*",namespace="%s",resp_code="5xx"}[5m])) by (namespace,job,resp_code) > 5000`, ns)),
							For:  "5m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
						{
							Alert: "ThreescaleBackendListenerJobDown",
							Annotations: map[string]string{
								"summary":     "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
								"description": "Job {{ $labels.job }} on {{ $labels.namespace }} is DOWN",
							},
							Expr: intstr.FromString(fmt.Sprintf(`up{job=~".*backend-listener.*",namespace="%s"} == 0`, ns)),
							For:  "1m",
							Labels: map[string]string{
								"severity": "critical",
							},
						},
					},
				},
			},
		},
	}
}
