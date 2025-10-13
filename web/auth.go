package web

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bluesky-social/indigo/atproto/crypto"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/golang-jwt/jwt/v5"
	"github.com/orthanc/feedgenerator/following"
)

var Hostanme = os.Getenv("FEEDGEN_HOSTNAME")
var ServiceDid = fmt.Sprintf("did:web:%s", Hostanme)
var PublisherDid = os.Getenv("FEEDGEN_PUBLISHER_DID")

type AtProtoSigningMethod struct {
	alg string
}

func (m *AtProtoSigningMethod) Alg() string {
	return m.alg
}

func (m *AtProtoSigningMethod) Verify(signingString string, signature []byte, key interface{}) error {
	return key.(crypto.PublicKey).HashAndVerifyLenient([]byte(signingString), signature)
}

func (m *AtProtoSigningMethod) Sign(signingString string, key interface{}) ([]byte, error) {
	return nil, fmt.Errorf("unimplemented")
}

func init() {
	ES256K := AtProtoSigningMethod{alg: "ES256K"}
	jwt.RegisterSigningMethod(ES256K.Alg(), func() jwt.SigningMethod {
		return &ES256K
	})

	ES256 := AtProtoSigningMethod{alg: "ES256"}
	jwt.RegisterSigningMethod(ES256.Alg(), func() jwt.SigningMethod {
		return &ES256
	})
}

func validateAuth(r *http.Request) (string, error) {
	headerValues := r.Header["Authorization"]
	if len(headerValues) != 1 {
		return "", fmt.Errorf("missing authorization header")
	}
	token := strings.TrimSpace(strings.Replace(headerValues[0], "Bearer ", "", 1))

	nsid := strings.Replace(r.URL.Path, "/xrpc/", "", 1)

	parsedToken, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		did := syntax.DID(token.Claims.(jwt.MapClaims)["iss"].(string))
		identity, err := following.DidDirectory.LookupDID(r.Context(), did)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve did %s: %s", did, err)
		}
		key, err := identity.PublicKey()
		if err != nil {
			return nil, fmt.Errorf("signing key not found for did %s: %s", did, err)
		}
		return key, nil
	})
	if err != nil {
		return "", fmt.Errorf("invalid token: %s", err)
	}

	claims := parsedToken.Claims.(jwt.MapClaims)
	if claims["lxm"] != nsid {
		return "", fmt.Errorf("bad jwt lexicon method (\"lxm\"). must match: %s", nsid)
	}
	return claims["iss"].(string), nil
}
