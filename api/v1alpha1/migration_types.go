/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MigrationSpec defines the desired state of Migration
type MigrationSpec struct {
	DB  DBSpec  `json:"db"`
	SQL SQLSpec `json:"sql"`
}

type DBSpec struct {
	Host   string     `json:"host"`
	Port   int32      `json:"port"`
	DBName string     `json:"dbName"`
	Secret SecretSpec `json:"secret,omitempty"`
	Vault  VaultSpec  `json:"vault,omitempty"`
	Driver string     `json:"driver"`
}

type SecretSpec struct {
	Name        string `json:"name"`
	UserKey     string `json:"userKey"`
	PasswordKey string `json:"passwordKey"`
}

type VaultSpec struct {
}

type SQLSpec struct {
	Git         GitMigrationSpec `json:"fromGit,omitempty"`
	VolumeClaim string           `json:"fromVolumeClaim,omitempty"`
	Path        string           `json:"path"`
}

type GitMigrationSpec struct {
	CheckoutURL string `json:"checkoutUrl"`
	Branch      string `json:"branch"`
	Secret      string `json:"secret"`
}

// MigrationStatus defines the observed state of Migration
type MigrationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Migration is the Schema for the migrations API
type Migration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MigrationSpec   `json:"spec,omitempty"`
	Status MigrationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MigrationList contains a list of Migration
type MigrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Migration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Migration{}, &MigrationList{})
}
