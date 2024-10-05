package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/crypto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/golang-jwt/jwt/v5"
)

var Hostanme = os.Getenv("FEEDGEN_HOSTNAME")
var ServiceDid = fmt.Sprintf("did:web:%s", Hostanme)
var PublisherDid = os.Getenv("FEEDGEN_PUBLISHER_DID")
var directory = identity.DefaultDirectory();

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

func wellKnownDidHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(map[string]any{
		"@context": []string{"https://www.w3.org/ns/did/v1"},
		"id":       ServiceDid,
		"service": []any{
			map[string]any{
				"id":              "#bsky_fg",
				"type":            "BskyFeedGenerator",
				"serviceEndpoint": fmt.Sprintf("https://%s", Hostanme),
			},
		},
	})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func describeFeedGenerator(w http.ResponseWriter, r *http.Request) {
	// TODO Build feeds dynamically based on algorithms
	feeds := []*bsky.FeedDescribeFeedGenerator_Feed{
		{Uri: fmt.Sprintf("at://%s/app.bsky.feed.generator/%s", PublisherDid, "replies-foll")},
	}
	data, err := json.Marshal(bsky.FeedDescribeFeedGenerator_Output{
		Did:   ServiceDid,
		Feeds: feeds,
	})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func validateAuth(r *http.Request) (string, error) {
	headerValues := r.Header["Authorization"]
	if len(headerValues) != 1 {
		return "", fmt.Errorf("missing authorization header")
	}
	token := strings.TrimSpace(strings.Replace(headerValues[0], "Bearer ",  "", 1))

	nsid := strings.Replace(r.URL.Path, "/xrpc/", "", 1)
	fmt.Println(nsid)

	parsedToken, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		did := syntax.DID(token.Claims.(jwt.MapClaims)["iss"].(string))
		identity, err := directory.LookupDID(r.Context(), did)
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

func getFeedSkeleton(w http.ResponseWriter, r *http.Request) {
	subjectDid, err := validateAuth(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(401)
		return
	}
	fmt.Println(subjectDid)

	// TODO actually load feed data from database
	data, err := json.Marshal(bsky.FeedGetFeedSkeleton_Output{
		Feed: []*bsky.FeedDefs_SkeletonFeedPost{
			{
				Post: "at://did:plc:difjsauz26vnv7c5woktj4ju/app.bsky.feed.post/3l5pu5bnmrd2c",
			},
			{
				Post: "at://did:plc:l5ykap4c5bmtdodwpikl24u3/app.bsky.feed.post/3l5pu5ervzl2y",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func StartServer() {
	http.HandleFunc("GET /.well-known/did.json", wellKnownDidHandler)
	http.HandleFunc("GET /xrpc/app.bsky.feed.describeFeedGenerator", describeFeedGenerator)
	http.HandleFunc("GET /xrpc/app.bsky.feed.getFeedSkeleton", getFeedSkeleton)

	fmt.Printf("Starting server on %s:%s\n", os.Getenv("FEEDGEN_LISTENHOST"), os.Getenv("FEEDGEN_PORT"))
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("FEEDGEN_LISTENHOST"), os.Getenv("FEEDGEN_PORT")), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Closing server")
}
