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
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	grpc "google.golang.org/grpc"
)

type accountError struct {
	message string
}

type authServer struct {
	UnimplementedAuthServer
}

// rs256 token
type JWTType struct {
	Alg string `json:"alg"` // rs256
	Typ string `json:"typ"` // JWT
	Kid string `json:"kid"` // key id
}

// rs256 payload
type JWTPayload struct {
	Sub  string `json:"sub"`  // uid
	Role string `json:"role"` // role
	Iat  int64  `json:"iat"`  // valid starting
	Exp  int64  `json:"exp"`  // expiry
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

	if val != newPass {
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
	exp := cur.Add(300 * time.Second)
	payload := JWTPayload{user, "user", cur.Unix(), exp.Unix()}
	pe, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	headString := b64.RawURLEncoding.EncodeToString([]byte(he))
	payloadString := b64.RawURLEncoding.EncodeToString([]byte(pe))
	firstTwoParts := fmt.Sprintf("%s.%s", headString, payloadString)

	h := sha256.New()
	h.Write([]byte(firstTwoParts))
	encoded := h.Sum(nil)

	priv, err := getPrivateKey()
	if err != nil {
		return "", err
	}

	ret, err := rsa.SignPKCS1v15(nil, priv, crypto.SHA256, encoded)
	if err != nil {
		return "", err
	}
	signature := b64.RawURLEncoding.EncodeToString(ret)

	return fmt.Sprintf("%s.%s", firstTwoParts, signature), nil
}

func verifyJWT(jwt string, user string, role string) error {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return &accountError{"Bad Token Split"}
	}

	head := make([]byte, b64.RawURLEncoding.DecodedLen(len(parts[0])))
	pay := make([]byte, b64.RawURLEncoding.DecodedLen(len(parts[1])))

	_, err := b64.RawURLEncoding.Decode(head, []byte(parts[0]))
	if err != nil {
		return err
	}
	_, err = b64.RawURLEncoding.Decode(pay, []byte(parts[1]))
	if err != nil {
		return err
	}

	var header JWTType
	var payload JWTPayload

	err = json.Unmarshal(head, &header)
	if err != nil {
		return err
	}

	err = json.Unmarshal(pay, &payload)
	if err != nil {
		return err
	}

	if !(header.Alg == "RS256" && header.Typ == "JWT" && header.Kid == "kid_v1") {
		return &accountError{"Incorrect JWT header"}
	}

	currTime := time.Now().Unix()
	if payload.Sub != user || payload.Role != role || payload.Iat > currTime || payload.Exp < currTime {
		return &accountError{"Incorrect JWT payload"}
	}

	firstTwoParts := fmt.Sprintf("%s.%s", parts[0], parts[1])

	h := sha256.New()
	h.Write([]byte(firstTwoParts))
	encoded := h.Sum(nil)

	pub, err := getPublicKey()
	if err != nil {
		return err
	}

	dec := make([]byte, b64.RawURLEncoding.DecodedLen(len(parts[2])))
	_, err = b64.RawURLEncoding.Decode(dec, []byte(parts[2]))
	if err != nil {
		return err
	}

	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, encoded, dec)
	if err != nil {
		return err
	}

	return nil
}

func validateAccountRequest(w http.ResponseWriter, req *http.Request, b *map[string]string) error {
	response := make(map[string]any)

	if req.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		response["error"] = "NotFound"
		response["message"] = "Only Post Allowed"
		response["status"] = 404

		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return &accountError{"NotFound"}
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response["error"] = "InternalServerError"
		response["message"] = "Error Reading Body"
		response["status"] = 500

		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return &accountError{"InternalServerError"}
	}

	err = json.Unmarshal(body, b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response["error"] = "InternalServerError"
		response["message"] = "Error Unmarshalling Payload"
		response["status"] = 500

		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return &accountError{"InternalServerError"}
	}

	if (*b)["user"] == "" || (*b)["pass"] == "" {
		w.WriteHeader(http.StatusBadRequest)
		response["error"] = "BadRequest"
		response["message"] = "Empty Username or Password"
		response["status"] = 400

		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return &accountError{"BadRequest"}
	}

	return nil
}

func loginApi(w http.ResponseWriter, req *http.Request) {
	b := make(map[string]string)
	response := make(map[string]any)
	ret := validateAccountRequest(w, req, &b)
	if ret != nil {
		return
	}

	jwt, err := login(RedisConnection, Ctx, b["user"], b["pass"])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response["error"] = "Unauthorized"
		response["message"] = "Error Logging In"
		response["status"] = 401

		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return
	}

	w.WriteHeader(http.StatusOK)
	response["auth"] = jwt
	res, _ := json.Marshal(response)
	w.Write([]byte(res))
}

func createAccountApi(w http.ResponseWriter, req *http.Request) {
	b := make(map[string]string)
	response := make(map[string]any)
	ret := validateAccountRequest(w, req, &b)
	if ret != nil {
		return
	}

	jwt, err := createAccount(RedisConnection, Ctx, b["user"], b["pass"])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response["error"] = "Unauthorized"
		response["message"] = "Error Creating Account"
		response["status"] = 401

		fmt.Println(err)
		res, _ := json.Marshal(response)
		w.Write([]byte(res))
		return
	}

	w.WriteHeader(http.StatusOK)
	response["auth"] = jwt
	res, _ := json.Marshal(response)
	w.Write([]byte(res))
}

func (s *authServer) VerifyToken(ctx context.Context, req *Token) (*Validation, error) {
	jwt := req.GetToken()
	user := req.GetUser()
	role := req.GetRole()

	err := verifyJWT(jwt, user, role)
	if err != nil {
		return &Validation{Resp: false}, err
	}

	return &Validation{Resp: true}, nil
}

func main() {
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			fmt.Println("Failed to open port")
			return
		}

		grpcServer := grpc.NewServer()
		RegisterAuthServer(grpcServer, &authServer{})

		if err := grpcServer.Serve(lis); err != nil {
			fmt.Println("Failed to bind grpc")
			return
		}
	}()

	http.HandleFunc("/signup", createAccountApi)
	http.HandleFunc("/login", loginApi)
	http.ListenAndServe(":3000", nil)
}
