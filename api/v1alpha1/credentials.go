package v1alpha1

import (
	"context"
	"encoding/base64"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type (
	Credential interface {
		GetUserPassword() (*UserPassword, error)
		MutateTemplate(tpl *corev1.PodTemplateSpec)
	}

	SecretCredential struct {
		Spec      *SecretSpec
		Namespace string
	}

	VaultCredential struct {
		Spec *VaultSpec
	}

	UserPassword struct {
		User     string
		Password string
	}
)

func (migration *Migration) GetCredentials() Credential {
	if migration.Spec.DB.Secret != (SecretSpec{}) {
		return SecretCredential{Spec: &migration.Spec.DB.Secret, Namespace: migration.ObjectMeta.Namespace}
	} else if migration.Spec.DB.Vault != (VaultSpec{}) {
		return VaultCredential{Spec: &migration.Spec.DB.Vault}
	}
	return nil
}

func (s SecretCredential) GetUserPassword() (*UserPassword, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}

	creds := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: s.Spec.Name}}
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: s.Namespace, Name: s.Spec.Name}, &creds); err != nil {
		return nil, err
	}
	var user, password []byte
	_, err = base64.StdEncoding.Decode(creds.Data["username"], user)
	if err != nil {
		return nil, err
	}
	_, err = base64.StdEncoding.Decode(creds.Data["password"], password)
	if err != nil {
		return nil, err
	}

	return &UserPassword{User: string(user), Password: string(password)}, nil
}

func (s SecretCredential) MutateTemplate(tpl *corev1.PodTemplateSpec) {
	extraEnvs := []corev1.EnvVar{
		corev1.EnvVar{
			Name: "FLYWAY_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: s.Spec.Name},
					Key:                  s.Spec.UserKey,
				},
			},
		},
		corev1.EnvVar{
			Name: "FLYWAY_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: s.Spec.Name},
					Key:                  s.Spec.PasswordKey,
				},
			},
		},
	}
	tpl.Spec.Containers[0].Env = append(tpl.Spec.Containers[0].Env, extraEnvs...)
}

func (v VaultCredential) GetUserPassword() (*UserPassword, error) {

	return nil, nil
}

func (s VaultCredential) MutateTemplate(tpl *corev1.PodTemplateSpec) {

}
