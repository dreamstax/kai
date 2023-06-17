/*
Copyright 2023 The Kai Authors.

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

package credentials

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// AWS
	AWSAccessKeyID         = "AWS_ACCESS_KEY_ID"
	AWSAccessKeyIDName     = "awsAccessKeyID"
	AWSSecretAccessKey     = "AWS_SECRET_ACCESS_KEY"
	AWSSecretAccessKeyName = "awsSecretAccessKey"

	// GCS
	GCSCredentialFileName        = "gcloud-application-credentials.json"
	GCSCredentialVolumeName      = "user-gcp-sa"
	GCSCredentialVolumeMountPath = "/var/secrets/"
	GCSCredentialEnvKey          = "GOOGLE_APPLICATION_CREDENTIALS"

	// Azure
	AzureStorageAccessKey = "AZURE_STORAGE_ACCESS_KEY"
)

type Client struct {
	client client.Client
	config *Config
}

type Config struct {
	S3Config    *S3Config
	GCSConfig   *GCSConfig
	AzureConfig *AzureConfig
}

type S3Config struct {
	S3AccessKeyID         string
	S3AccessKeyIDName     string
	S3SecretAccessKey     string
	S3SecretAccessKeyName string
}

type GCSConfig struct {
	GCSCredentialFileName        string
	GCSCredentialVolumeName      string
	GCSCredentialVolumeMountPath string
	GCSCredentialEnvKey          string
}

type AzureConfig struct {
	AzureStorageAccessKey string
}

func NewDefaultCredentialBuilder(client client.Client) *Client {
	return &Client{
		client: client,
		config: newDefaultConfig(),
	}
}

func newDefaultConfig() *Config {
	return &Config{
		S3Config:    newDefaultS3Config(),
		GCSConfig:   newDefaultGCSConfig(),
		AzureConfig: newDefaultAzureConfig(),
	}
}

func newDefaultS3Config() *S3Config {
	return &S3Config{
		S3AccessKeyID:         AWSAccessKeyID,
		S3AccessKeyIDName:     AWSAccessKeyIDName,
		S3SecretAccessKeyName: AWSSecretAccessKeyName,
	}
}

func newDefaultGCSConfig() *GCSConfig {
	return &GCSConfig{
		GCSCredentialFileName:        GCSCredentialFileName,
		GCSCredentialVolumeName:      GCSCredentialVolumeName,
		GCSCredentialVolumeMountPath: GCSCredentialVolumeMountPath,
		GCSCredentialEnvKey:          GCSCredentialEnvKey,
	}
}

func newDefaultAzureConfig() *AzureConfig {
	return &AzureConfig{
		AzureStorageAccessKey: AzureStorageAccessKey,
	}
}

func (c *Client) BuildCredentials(ctx context.Context, name types.NamespacedName, container *v1.Container, volumes *[]v1.Volume) error {
	serviceAccount := &v1.ServiceAccount{}
	err := c.client.Get(ctx, name, serviceAccount)
	if err != nil {
		return err
	}

	for _, ref := range serviceAccount.Secrets {
		secret := &v1.Secret{}
		err := c.client.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: name.Namespace}, secret)
		if err != nil {
			return err
		}

		// AWS
		_, ok := secret.Data[c.config.S3Config.S3SecretAccessKeyName]
		if ok {
			envs := c.buildS3Credentials(secret)
			container.Env = append(container.Env, envs...)
			continue
		}

		// GCS
		_, ok = secret.Data[c.config.GCSConfig.GCSCredentialFileName]
		if ok {
			volume, volumeMount := c.buildGCSCredentials(secret)
			*volumes = append(*volumes, volume)
			container.VolumeMounts = append(container.VolumeMounts, volumeMount)
			container.Env = append(container.Env, v1.EnvVar{
				Name:  c.config.GCSConfig.GCSCredentialEnvKey,
				Value: c.config.GCSConfig.GCSCredentialVolumeMountPath + c.config.GCSConfig.GCSCredentialFileName,
			})
			continue
		}

		// Azure
		_, ok = secret.Data[c.config.AzureConfig.AzureStorageAccessKey]
		if ok {
			env := c.buildAzureCredentials(secret)
			container.Env = append(container.Env, env)
		}
	}

	return nil
}

func (c *Client) buildS3Credentials(secret *v1.Secret) []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name: c.config.S3Config.S3AccessKeyID,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: secret.Name,
					},
					Key: c.config.S3Config.S3AccessKeyIDName,
				},
			},
		},
		{
			Name: c.config.S3Config.S3SecretAccessKey,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: secret.Name,
					},
					Key: c.config.S3Config.S3SecretAccessKeyName,
				},
			},
		},
	}
}

func (c *Client) buildGCSCredentials(secret *v1.Secret) (v1.Volume, v1.VolumeMount) {
	volume := v1.Volume{
		Name: c.config.GCSConfig.GCSCredentialVolumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secret.Name,
			},
		},
	}

	volumeMount := v1.VolumeMount{
		MountPath: c.config.GCSConfig.GCSCredentialVolumeMountPath,
		Name:      c.config.GCSConfig.GCSCredentialVolumeName,
		ReadOnly:  true,
	}

	return volume, volumeMount
}

func (c *Client) buildAzureCredentials(secret *v1.Secret) v1.EnvVar {
	return v1.EnvVar{
		Name: c.config.AzureConfig.AzureStorageAccessKey,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secret.Name,
				},
				Key: c.config.AzureConfig.AzureStorageAccessKey,
			},
		},
	}
}
