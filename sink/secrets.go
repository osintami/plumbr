// Copyright © 2022 Sloan Childers
package sink

import "os"

type ISecrets interface {
	Set(key, value string)
	Find(collectorName string) string
}

type SecretsManager struct {
	secrets map[string]string
}

func NewSecretsManager(keys []string) *SecretsManager {
	secrets := make(map[string]string)
	for _, key := range keys {
		secrets[key] = os.Getenv(key)
	}
	return &SecretsManager{secrets: secrets}
}

func (x *SecretsManager) Set(key, value string) {
	x.secrets[key] = value
}

func (x *SecretsManager) Find(collectorName string) string {
	return x.secrets[collectorName]
}
