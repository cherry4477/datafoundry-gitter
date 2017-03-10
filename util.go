package main

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"fmt"
	"time"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
)

func setBaseUrl(urlStr string) string {
	// Make sure the given URL end with a slash
	if strings.HasSuffix(urlStr, "/") {
		return setBaseUrl(strings.TrimSuffix(urlStr, "/"))
	}
	return urlStr
}

func randToken() string {
	b := make([]byte, 40)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// func redirectUrl(oauthConf *oauth2.Config) string {
// 	return ""
// }

// rsa public and private keys
func generateKeyPair() (privateKey, publicKey string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}
	err = priv.Validate()
	if err != nil {
		return
	}

	priv_der := x509.MarshalPKCS1PrivateKey(priv)

	// pem.Block
	// blk pem.Block
	priv_blk := pem.Block {
		Type: "RSA PRIVATE KEY",
		Headers: nil,
		Bytes: priv_der,
	}

	// Resultant private key in PEM format.
	// priv_pem string
	privateKey = string(pem.EncodeToMemory(&priv_blk))
	//println("private:", privateKey)
	
	// ...
	
	pub := priv.PublicKey
	
	// pub.pem
	//pub_der, err := x509.MarshalPKIXPublicKey(&pub)
	//if err != nil {
	//	return
	//}
	//
	//pub_blk := pem.Block {
	//	Type: "PUBLIC KEY",
	//	Headers: nil,
	//	Bytes: pub_der,
	//}
	//publicKey = string(pem.EncodeToMemory(&pub_blk))
	//println("public:", publicKey)
	
	sshpub, err := ssh.NewPublicKey(&pub)
	if err != nil {
		return
	}
	publicKey = string(ssh.MarshalAuthorizedKey(sshpub))
	publicKey = strings.TrimRight(publicKey, "\n")
	publicKey = fmt.Sprintf("%s rsa-key-%s", publicKey, time.Now().Format("20060102"))
	//println("public:", publicKey)
	
	return
}