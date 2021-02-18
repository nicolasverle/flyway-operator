package v1alpha1

import corev1 "k8s.io/api/core/v1"

type (
	ScriptsLocation interface {
		MutateTemplate(tpl *corev1.PodTemplateSpec)
	}

	GitLocation struct {
		Spec *GitMigrationSpec
	}

	VolumeLocation struct {
		Name string
	}
)

const gitMountName = "git-key"

func GetScriptsLocation(spec *SQLSpec) ScriptsLocation {
	if spec.Git != (GitMigrationSpec{}) {
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
			Image: "k8s.gcr.io/git-sync:v3.1.5",
			Env: []corev1.EnvVar{
				corev1.EnvVar{Name: "GIT_SYNC_REPO", Value: git.Spec.CheckoutURL},
				corev1.EnvVar{Name: "GIT_SYNC_BRANCH", Value: git.Spec.Branch},
				corev1.EnvVar{Name: "GIT_SYNC_ROOT", Value: "/opt/sources/"},
				corev1.EnvVar{Name: "GIT_SYNC_DEST", Value: git.Spec.Branch},
				corev1.EnvVar{Name: "GIT_SYNC_SSH", Value: "true"},
				corev1.EnvVar{Name: "GIT_SSH_KEY_FILE", Value: "/etc/git-secret/id_rsa"},
				corev1.EnvVar{Name: "GIT_KNOWN_HOSTS", Value: "false"},
				corev1.EnvVar{Name: "GIT_SYNC_ONE_TIME", Value: "true"},
			},
			VolumeMounts: []corev1.VolumeMount{
				corev1.VolumeMount{Name: SQLVolumeName, MountPath: "/opt/sources/"},
				corev1.VolumeMount{Name: gitMountName, MountPath: "/etc/git-secret"},
			},
		},
	}
	tpl.Spec.Volumes = []corev1.Volume{
		corev1.Volume{Name: SQLVolumeName, VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
		corev1.Volume{Name: gitMountName, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: git.Spec.Secret}}},
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
