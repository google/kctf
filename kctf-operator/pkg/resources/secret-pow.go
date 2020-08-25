package resources

import (
	"io/ioutil"
	"os/exec"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func extractKey(address string) []byte {
	// Where we will store the key
	key, err := ioutil.ReadFile(address)
	if err != nil {
		log.Error(err, "Failed opening the key file")
	}

	return key
}

func generateKeys() {
	// Generate the keys
	// Generate first pow-bypass-key.pem
	cmd := exec.Command("openssl", "ecparam", "-name", "prime256v1",
		"-genkey", "-noout", "-out", "pow-bypass-key.pem")
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "cmd.Run() failed\n")
	}

	// Then generate first pow-bypass-key-pub.pem
	cmd = exec.Command("openssl", "ec", "-in", "pow-bypass-key.pem",
		"-pubout", "-out", "pow-bypass-key-pub.pem")
	_, err = cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "cmd.Run() failed\n")
	}
}

func secret(name string, nameMap string) *corev1.Secret {
	// We extract the key
	key := extractKey(nameMap)
	data := map[string][]byte{nameMap: key}
	// Then we create the secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kube-system",
		},
		Data: data,
	}

	return secret
}

func NewSecretPowBypass() runtime.Object {
	return secret("pow-bypass", "pow-bypass-key.pem")
}

func NewSecretPowBypassPub() runtime.Object {
	return secret("pow-bypass-pub", "pow-bypass-key-pub.pem")
}
