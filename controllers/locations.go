package controllers

import (
	migrationsv1alpha1 "flyway-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

type (
	ScriptsLocation interface {
		MutateTemplate(tpl *corev1.PodTemplateSpec)
	}

	GitLocation struct {
		Spec *migrationsv1alpha1.GitMigrationSpec
	}

	VolumeLocation struct {
		Name string
	}
)

const (
	gitMountName = "git-key"
	// SQLVolumeName that sets the name of volume for sql scripts
	SQLVolumeName = "sql-scripts"
)

func GetScriptsLocation(spec *migrationsv1alpha1.SQLSpec) ScriptsLocation {
	if spec.Git != (migrationsv1alpha1.GitMigrationSpec{}) {
		return GitLocation{Spec: &spec.Git}
	} else if spec.VolumeClaim != "" {
		return VolumeLocation{Name: spec.VolumeClaim}
	}
	return nil
}

func (git GitLocation) MutateTemplate(tpl *corev1.PodTemplateSpec) {
	tpl.Spec.InitContainers = []corev1.Container{
		corev1.Container{
			Name:  "git",
			Image: "alpine/git:1.0.2",
			Env: []corev1.EnvVar{
				corev1.EnvVar{Name: "GIT_SSH_COMMAND", Value: "ssh -o StrictHostKeyChecking=no -i /etc/git-secret/id_rsa"},
			},
			Args: []string{"clone", "--branch", git.Spec.Branch, git.Spec.CheckoutURL, "/opt/sources"},
			VolumeMounts: []corev1.VolumeMount{
				corev1.VolumeMount{Name: SQLVolumeName, MountPath: "/opt/sources/"},
				corev1.VolumeMount{Name: gitMountName, MountPath: "/etc/git-secret"},
			},
		},
	}
	mode := int32(256)
	tpl.Spec.Volumes = []corev1.Volume{
		corev1.Volume{
			Name: SQLVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: gitMountName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  git.Spec.Secret,
					DefaultMode: &mode,
				},
			},
		},
	}
}

func (vol VolumeLocation) MutateTemplate(tpl *corev1.PodTemplateSpec) {
	tpl.Spec.Volumes = []corev1.Volume{
		corev1.Volume{
			Name: SQLVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: vol.Name},
				},
			},
		},
	}
}
