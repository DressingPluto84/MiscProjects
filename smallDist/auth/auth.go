package main

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type accountError struct {
	message string
}

// rs256 token
type JWTType struct {
	alg string // rs256
	typ string // JWT
	kid string // key id
}

// rs256 payload
type JWTPayload struct {
	sub  string // uid
	role string // role
	iat  int64  // valid starting
	exp  int64  // expiry
}

type loginPayload struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

const SALT = "ianwslhhihyesamsa"

var RedisConnection = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password
	DB:       0,  // use default DB
	Protocol: 2,
})
var Ctx = context.Background()

func getSaltedPass(pass string) string {
	var s strings.Builder
	hasher := sha256.New()

	s.WriteString(pass)
	s.WriteString(":")
	s.WriteString(SALT)

	hasher.Write([]byte(s.String()))

	return hex.EncodeToString(hasher.Sum(nil))
}

func (e *accountError) Error() string {
	return e.message
}

func login(rdb *redis.Client, ctx context.Context, user string, pass string) (string, error) {
	newPass := getSaltedPass(pass)

	val, err := rdb.Get(ctx, user).Result()
	if err != nil {
		return "", err
	}

	actualSalted := getSaltedPass(val)
	if actualSalted != newPass {
		return "", &accountError{"Incorrect Password"}
	}

	jwtUser := fmt.Sprintf("auth:%s", user)
	val, _ = rdb.Get(ctx, jwtUser).Result()
	if err != nil {
		return "", err
	}
	if val != "" {
		return val, nil
	}

	jwt, err := getJWT(user)
	if err != nil {
		return "", err
	}

	err = rdb.Set(ctx, jwtUser, jwt, 5*time.Minute).Err()
	if err != nil {
		return "", err
	}

	return jwt, nil
}

func createAccount(rdb *redis.Client, ctx context.Context, user string, pass string) (string, error) {
	newPass := getSaltedPass(pass)

	_, err := rdb.Get(ctx, user).Result()
	if err != redis.Nil {
		return "", &accountError{"Account Already Exists"}
	}

	err = rdb.Set(ctx, user, newPass, 2*time.Hour).Err()
	if err != nil {
		return "", err
	}

	jwt, err := login(rdb, ctx, user, pass)
	if err != nil {
		return "", err
	}

	return jwt, nil
}

func getJWT(user string) (string, error) {
	header := JWTType{"RS256", "JWT", "kid_v1"}
	he, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	cur := time.Now()
	exp := cur.Add(600 * time.Second)
	payload := JWTPayload{user, "user", cur.Unix(), exp.Unix()}
	pe, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	headString := b64.StdEncoding.EncodeToString([]byte(he))
	payloadString := b64.StdEncoding.EncodeToString([]byte(pe))
	firstTwoParts := fmt.Sprint("%s.%s", headString, payloadString)

	h := sha256.New()
	h.Write([]byte(firstTwoParts))
	encoded := h.Sum(nil)

	priv, err := getKeys()
	if err != nil {
		return "", err
	}

	ret, err := rsa.SignPKCS1v15(nil, priv, crypto.SHA256, encoded)
	if err != nil {
		return "", err
	}
	signature := b64.StdEncoding.EncodeToString(ret)

	return fmt.Sprintf("%s.%s", firstTwoParts, signature), nil
}

func loginApi(w http.ResponseWriter, req *http.Request) {
	response := make(map[string]any)

	if req.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		response["error"] = "NotFound"
		response["message"] = "Only Post Allowed"
		response["status"] = 404

		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response["error"] = "InternalServerError"
		response["message"] = "Error Reading Body"
		response["status"] = 500

		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return
	}

	b := make(map[string]string)
	err = json.Unmarshal(body, &b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response["error"] = "InternalServerError"
		response["message"] = "Error Unmarshalling Payload"
		response["status"] = 500

		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return
	}

	if b["user"] == "" || b["pass"] == "" {
		w.WriteHeader(http.StatusBadRequest)
		response["error"] = "BadRequest"
		response["message"] = "Empty Username or Password"
		response["status"] = 400

		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return
	}

	jwt, err := login(RedisConnection, Ctx, b["user"], b["pass"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response["error"] = "BadRequest"
		response["message"] = "Invalid Username or Password"
		response["status"] = 400

		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return
	}

	w.WriteHeader(http.StatusOK)
	response["auth"] = jwt
	res, _ := json.Marshal(response)
	w.Write([]byte(res))
}

func main() {
	http.HandleFunc("/login", loginApi)
	http.ListenAndServe(":3000", nil)
}
