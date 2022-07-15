package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewExternalDnsClusterRole() client.Object {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "external-dns-viewer",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "create", "update", "patch", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services", "endpoints", "pods", "nodes"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"extensions", "networking.k8s.io"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}
	return clusterRole
}

func NewExternalDnsClusterRoleBinding() client.Object {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "external-dns-sa:external-dns-viewer",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "external-dns-viewer",
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      "external-dns-sa",
			Namespace: "kctf-system",
		}},
	}
	return clusterRoleBinding
}

func NewExternalDnsDeployment() client.Object {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "external-dns",
			Namespace: "kctf-system",
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "external-dns"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "external-dns"},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "external-dns-sa",
					Containers: []corev1.Container{
						{
							Image: "us.gcr.io/k8s-artifacts-prod/external-dns/external-dns:v0.10.0",
							Name:  "external-dns",
							Env: []corev1.EnvVar{{
								Name: "DOMAIN_NAME",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "external-dns",
										},
										Key: "DOMAIN_NAME",
									},
								},
							}, {
								Name: "CLUSTER_NAME",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "external-dns",
										},
										Key: "CLUSTER_NAME",
									},
								},
							}, },
							Args: []string{"--log-level=debug", "--source=service", "--source=ingress",
								"--provider=google", "--domain-filter=$(DOMAIN_NAME)", "--registry=txt",
								"--txt-owner-id=kctf-cloud-dns-$(CLUSTER_NAME)"},
						},
						{
							Image: DOCKER_CERTBOT_IMAGE,
							Name:  "certbot",
							Env: []corev1.EnvVar{
								{
									Name: "DOMAIN",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "external-dns",
											},
											Key: "DOMAIN_NAME",
										},
									},
								},
								{
									Name:  "SECRET",
									Value: "tls-cert",
								},
								{
									Name: "EMAIL",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "external-dns",
											},
											Key: "EMAIL_ADDRESS",
										},
									},
								},
								{
									Name:  "PROD",
									Value: "true",
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}
