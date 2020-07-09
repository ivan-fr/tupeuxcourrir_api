package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"reflect"
	"strings"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"

	"github.com/dgrijalva/jwt-go"
)

type (
	// JWTConfig defines the config for JWT middleware.
	JWTConfig struct {
		// SuccessHandler defines a function which is executed for a valid token.
		SuccessHandler JWTSuccessHandler

		// Signing key to validate token. Used as fallback if SigningKeys has length 0.
		// Required. This or SigningKeys.
		SigningKey interface{}

		// Map of signing keys to validate token with kid field usage.
		// Required. This or SigningKey.
		SigningKeys map[string]interface{}

		// Signing method, used to check token signing method.
		// Optional. Default value HS256.
		SigningMethod string

		// Context key to store user information from the token into context.
		// Optional. Default value "user".
		ContextKey string

		// Claims are extendable claims data defining token content.
		// Optional. Default value jwt.MapClaims
		Claims jwt.Claims

		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "param:<name>"
		// - "cookie:<name>"
		TokenLookup string

		// AuthScheme to be used in the Authorization header.
		// Optional. Default value "Bearer".
		AuthScheme string

		keyFunc jwt.Keyfunc
	}

	JwtUserCustomClaims struct {
		UserID int `json:"id"`
		jwt.StandardClaims
	}

	// JWTSuccessHandler defines a function which is executed for a valid token.
	JWTSuccessHandler func(ctx context.Context) context.Context

	jwtExtractor func(*http.Request) string

	ImplementJWTUser struct {
		AddInitiatedThread bool
		AddRoles           bool
		AddReceivedThread  bool
		GiveMeSQB          bool
		Subject            string
	}
)

var MyJWTUserConfig = JWTConfig{
	Claims:     &JwtUserCustomClaims{},
	SigningKey: []byte(config.JWTSecret),
	ContextKey: "JWTContextUser",
}

// Algorithms
const (
	AlgorithmHS256 = "HS256"
)

var (
	// DefaultJWTConfig is the default JWT auth middleware config.
	DefaultJWTConfig = JWTConfig{
		SigningMethod: AlgorithmHS256,
		ContextKey:    "user",
		TokenLookup:   "header:" + config.HeaderAuthorization,
		AuthScheme:    "Bearer",
		Claims:        jwt.MapClaims{},
	}
)

// JWTWithConfig returns a JWT auth middleware with config.
// See: `JWT()`.
func JWTWithConfig(config JWTConfig) mux.MiddlewareFunc {
	// Defaults
	if config.SigningKey == nil && len(config.SigningKeys) == 0 {
		panic("echo: jwt middleware requires signing key")
	}
	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}
	if config.Claims == nil {
		config.Claims = DefaultJWTConfig.Claims
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}
	if config.AuthScheme == "" {
		config.AuthScheme = DefaultJWTConfig.AuthScheme
	}
	config.keyFunc = func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != config.SigningMethod {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		if len(config.SigningKeys) > 0 {
			if kid, ok := t.Header["kid"].(string); ok {
				if key, ok := config.SigningKeys[kid]; ok {
					return key, nil
				}
			}
			return nil, fmt.Errorf("unexpected jwt key id=%v", t.Header["kid"])
		}

		return config.SigningKey, nil
	}

	// Initialize
	parts := strings.Split(config.TokenLookup, ":")
	extractor := jwtFromHeader(parts[1], config.AuthScheme)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
			auth := extractor(r)
			if auth == "" {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(errors.New("forbidden")))
			}
			token := new(jwt.Token)
			// Issue #647, #656
			if _, ok := config.Claims.(jwt.MapClaims); ok {
				token, err = jwt.Parse(auth, config.keyFunc)
			} else {
				t := reflect.ValueOf(config.Claims).Type().Elem()
				claims := reflect.New(t).Interface().(jwt.Claims)
				token, err = jwt.ParseWithClaims(auth, claims, config.keyFunc)
			}
			if err == nil && token.Valid {
				// Store user information from token into context.
				ctx := context.WithValue(r.Context(), config.ContextKey, token)
				if config.SuccessHandler != nil {
					ctx = config.SuccessHandler(ctx)
				}
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(errors.New("invalid or expired jwt")))
			}
		})
	}
}

// jwtFromHeader returns a `jwtExtractor` that extracts token from the request header.
func jwtFromHeader(header string, authScheme string) jwtExtractor {
	return func(r *http.Request) string {
		auth := r.Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:]
		}
		return ""
	}
}

func (jCC *JwtUserCustomClaims) GetToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jCC)

	// Generate encoded token and send it as response.
	stringToken, err := token.SignedString([]byte(config.JWTSecret))

	if err != nil {
		panic(err)
	}

	return stringToken
}

func ImplementUserFromJwtSuccessHandler(iJU *ImplementJWTUser) JWTSuccessHandler {
	return func(ctx context.Context) context.Context {
		JWTContext := ctx.Value("JWTContextUser").(*jwt.Token)
		claims := JWTContext.Claims.(*JwtUserCustomClaims)

		if claims.Subject == iJU.Subject {
			sQB := orm.GetSelectQueryBuilder(models.NewUser())

			if iJU.AddInitiatedThread {
				sQB = sQB.Consider("InitiatedThreads")
			}

			if iJU.AddReceivedThread {
				sQB = sQB.Consider("ReceivedThreads")
			}

			if iJU.AddRoles {
				sQB = sQB.Consider("Roles")
			}

			sQB = sQB.Where(orm.And(orm.H{"IdUser": claims.UserID}))

			if iJU.GiveMeSQB {
				return context.WithValue(ctx, "uSQB", sQB)
			} else {
				var err error
				err = sQB.ApplyQueryRow()

				if err != nil {
					panic(err)
				}
				return context.WithValue(ctx, "user", sQB.EffectiveModel)
			}
		} else {
			if iJU.GiveMeSQB {
				return context.WithValue(ctx, "uSQB", nil)
			}

			return context.WithValue(ctx, "user", nil)
		}
	}
}
