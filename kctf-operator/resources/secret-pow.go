package resources

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var privateKey *ecdsa.PrivateKey = generateKey()

func generateKey() *ecdsa.PrivateKey {
	// Generate the public key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// Check error
	if err != nil {
		log.Error(err, "Failed to generate private key")
	}

	return privateKey
}

func secret(name string, nameMap string, private bool) *corev1.Secret {
	var err error
	var der []byte
	var block pem.Block

	// We take the right key
	if private == true {
		der, err = x509.MarshalPKCS8PrivateKey(privateKey)
		block.Type = "EC PRIVATE KEY"
	} else {
		der, err = x509.MarshalPKIXPublicKey(privateKey.Public())
		block.Type = "PUBLIC KEY"
	}

	// Check error
	if err != nil {
		log.Error(err, "Couldn't get DER form")
	}

	block.Bytes = der

	// Transform in bytes
	pem := pem.EncodeToMemory(&block)

	data := map[string][]byte{nameMap: pem}
	// Then we create the secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kctf-system",
		},
		Data: data,
	}

	return secret
}

func NewSecretPowBypass() client.Object {
	return secret("pow-bypass", "pow-bypass-key.pem", true)
}

func NewSecretPowBypassPub() client.Object {
	return secret("pow-bypass-pub", "pow-bypass-key-pub.pem", false)
}

func NewSecretTls() client.Object {
	// Generate empty secret so ingress works
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-cert",
			Namespace: "kctf-system",
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte{},
			corev1.TLSPrivateKeyKey: []byte{},
		},
	}
	return secret
}
