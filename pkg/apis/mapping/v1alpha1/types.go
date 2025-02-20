package v1alpha1

import (
	"io/ioutil"

	"github.com/jenkins-x/jx-api/v4/pkg/util"

	"github.com/pkg/errors"
	"gopkg.in/validator.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	SecretMappingFileName = "secret-mappings.yaml"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SecretMapping represents a collection of mappings of Secrets to destinations in the underlying secret store (e.g. Vault keys)
//
// +k8s:openapi-gen=true
type SecretMapping struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata"`

	// Spec the definition of the secret mappings
	Spec SecretMappingSpec `json:"spec"`
}

// SecretMappingSpec defines the desired state of SecretMapping.
type SecretMappingSpec struct {
	// Secrets rules for each secret
	Secrets []SecretRule `json:"secrets,omitempty"`

	Defaults `json:"defaults,omitempty" validate:"nonzero"`
}

// Defaults contains default mapping configuration for any Kubernetes secrets to External Secrets
type Defaults struct {
	// DefaultBackendType the default back end to use if there's no specific mapping
	BackendType BackendType `json:"backendType,omitempty" validate:"nonzero"`

	// RoleArn is used for some back ends like AWS and Alicloud
	RoleArn string `json:"roleArn,omitempty"`

	// Region is used for some back ends like AWS
	Region string `json:"region,omitempty"`

	// VersionStage the default version stage to use which is used on some back ends like AWS and Alicloud
	VersionStage string `json:"versionStage,omitempty"`

	// AzureKeyVault config
	AzureKeyVaultConfig *AzureKeyVaultConfig `json:"azureKeyVault,omitempty"`

	// GcpSecretsManager config
	GcpSecretsManager *GcpSecretsManager `json:"gcpSecretsManager,omitempty"`

	// AwsSecretsManager config
	AwsSecretsManager *AwsSecretsManager `json:"secretsManager,omitempty"`
}

// SecretMappingList contains a list of SecretMapping
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SecretMappingList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretMapping `json:"items"`
}

// SecretRule the rules for a specific Secret
type SecretRule struct {
	// Name name of the secret
	Name string `json:"name,omitempty"`
	// Namespace name of the secret
	Namespace string `json:"namespace,omitempty"`
	// BackendType for the secret
	BackendType BackendType `json:"backendType"`
	// Mappings one more mappings
	Mappings []Mapping `json:"mappings,omitempty"`
	// Unsecured represent a list of a secret's keys that will remain as plain secrets rather than undergoing conversion
	Unsecured []string `json:"unsecured,omitempty"`
	// RoleArn is used for some back ends like AWS and Alicloud
	RoleArn string `json:"roleArn,omitempty"`
	// Region is used for some back ends like AWS
	Region string `json:"region,omitempty"`
	// AzureKeyVaultConfig config
	AzureKeyVaultConfig *AzureKeyVaultConfig `json:"azureKeyVault,omitempty"`
	// GcpSecretsManager config
	GcpSecretsManager *GcpSecretsManager `json:"gcpSecretsManager,omitempty"`
	// AwsSecretsManager config
	AwsSecretsManager *AwsSecretsManager `json:"secretsManager,omitempty"`
}

// BackendType describes a secrets backend
type BackendType string

const (
	// BackendTypeAlicloud Alicloud KMS Secret Manager as the Backed service
	BackendTypeAlicloud BackendType = "alicloudSecretsManager"
	// BackendTypeAWSSecretsManager AWS Secrets Manager as the Backed service
	BackendTypeAWSSecretsManager BackendType = "secretsManager"
	// BackendTypeAWSParameterStore AWS SSM Parameter Store as the Backed service
	BackendTypeAWSParameterStore BackendType = "systemManager"
	// BackendTypeAzure Azure Key Vault as the Backed service
	BackendTypeAzure BackendType = "azureKeyVault"
	// BackendTypeGSM Google Secrets Manager is the Backed service
	BackendTypeGSM BackendType = "gcpSecretsManager"
	// BackendTypeIBMSecretsManager IBM Secrets Manager is the Backed service
	BackendTypeIBMSecretsManager BackendType = "ibmcloudSecretsManager"
	// BackendTypeLocal local secrets - i.e. vanilla k8s Secrets
	BackendTypeLocal BackendType = "local"
	// BackendTypeVault Vault is the Backed service
	BackendTypeVault BackendType = "vault"
	// BackendTypeNone if none is configured
	BackendTypeNone BackendType = ""
)

// GcpSecretsManager stores default config when using GSM for secret storage
type GcpSecretsManager struct {
	// Version of the referenced secret
	Version string `json:"version,omitempty"`
	// ProjectID for the secret, defaults to the current GCP project
	ProjectID string `json:"projectId,omitempty"`
	// UniquePrefix needs to be a unique prefix in the GCP project where the secret resides, defaults to cluster name
	UniquePrefix string `json:"uniquePrefix,omitempty"`
}

// AwsSecretsManager stores default config when using AWS Secret Manager for secret storage
type AwsSecretsManager struct {
	RoleArn      string `json:"roleArn,omitempty"`
	Region       string `json:"region,omitempty"`
	VersionStage string `json:"versionStage,omitempty"`
}

// AzureKeyVaultConfig stores default config when using Azure Key Vault for secret storage
type AzureKeyVaultConfig struct {
	KeyVaultName string `json:"keyVaultName,omitempty"`
}

// Mapping the predicates which must be true to invoke the associated tasks/pipelines
type Mapping struct {
	// Name the secret entry name which maps to the Key of the Secret.Data map
	Name string `json:"name,omitempty"`

	// Key the Vault key to load the secret value
	// +optional
	Key string `json:"key,omitempty"`

	// Property the Vault property on the key to load the secret value
	// +optional
	Property string `json:"property,omitempty"`

	// VersionStage the version of the secret value
	// +optional
	VersionStage string `json:"versionStage,omitempty"`

	// IsBinary to indicate a binary secret
	// +optional
	IsBinary bool `json:"isBinary,omitempty"`
}

// FindRule finds a secret rule for the given secret name
func (c *SecretMapping) FindRule(namespace, secretName string) *SecretRule {
	for i := range c.Spec.Secrets {
		m := &c.Spec.Secrets[i]
		if m.Name == secretName && (m.Namespace == "" || m.Namespace == namespace) {
			return &c.Spec.Secrets[i]
		}
	}
	return &SecretRule{
		BackendType: c.Spec.Defaults.BackendType,
	}
}

// Find finds a secret rule for the given secret name
func (c *SecretMapping) Find(secretName, dataKey string) *Mapping {
	for i := range c.Spec.Secrets {
		m := &c.Spec.Secrets[i]
		if m.Name == secretName {
			return c.Spec.Secrets[i].Find(dataKey)
		}
	}
	return nil
}

// Find finds a secret rule for the given secret name
func (c *SecretMapping) FindSecret(secretName string) *SecretRule {
	for i := range c.Spec.Secrets {
		m := &c.Spec.Secrets[i]
		if m.Name == secretName {
			return &c.Spec.Secrets[i]
		}
	}
	return nil
}

func (c *SecretMapping) IsSecretKeyUnsecured(secretName, keyName string) bool {
	secret := c.FindSecret(secretName)
	if secret == nil {
		return false
	}
	for _, u := range secret.Unsecured {
		if u == keyName {
			return true
		}
	}
	return false
}

// Find finds a mapping for the given data name
func (r *SecretRule) Find(dataKey string) *Mapping {
	for i := range r.Mappings {
		m := &r.Mappings[i]
		if m.Name == dataKey {
			return &r.Mappings[i]
		}
	}
	return nil
}

// validate the secrete mapping fields
func (c *SecretMapping) Validate() error {
	return validator.Validate(c)
}

// SaveConfig saves the configuration file to the given project directory
func (c *SecretMapping) SaveConfig(fileName string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, data, util.DefaultWritePermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to save file %s", fileName)
	}

	return nil
}

// DestinationString returns a unique string for where the entry will be stored so that we can find
// secrets using the same storage location.
func (c *SecretMapping) DestinationString(rule *SecretRule, mapping *Mapping) string {
	defaults := c.Spec.Defaults
	backend := rule.BackendType
	if backend == BackendTypeNone {
		backend = defaults.BackendType
	}

	switch backend {
	case BackendTypeGSM:
		if rule.GcpSecretsManager == nil {
			rule.GcpSecretsManager = &GcpSecretsManager{}
		}
		projectID := rule.GcpSecretsManager.ProjectID
		if projectID == "" {
			projectID = defaults.GcpSecretsManager.ProjectID
		}
		prefix := rule.GcpSecretsManager.UniquePrefix
		if prefix == "" {
			prefix = c.Spec.GcpSecretsManager.UniquePrefix
		}
		return projectID + "/" + prefix + "-" + mapping.Key + "/" + mapping.Property
	default:
		return mapping.Key + "/" + mapping.Property
	}
}
