package helpers

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/joho/godotenv"
)

// EnvVar function's purpose is to read env variables from .env file or use a default value instead
func EnvVar(key string, defaultVal string) string {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	value := os.Getenv(key)

	if len(value) == 0 {
		return defaultVal
	}
	return value
}

// SignedDetails
type SignedDetails struct {
	ID    string
	Email string
	jwt.StandardClaims
}

// GenerateTokens generates detailed token
func GenerateToken(userType, userId, email string) (signedToken string, err error) {
	var SECRET_KEY string
	if userType == "ADMIN" {
		SECRET_KEY = EnvVar("ADMIN_SECRET_KEY", "")
	} else {
		SECRET_KEY = EnvVar("SECRET_KEY", "")
	}
	claims := &SignedDetails{
		ID:    userId,
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	if err != nil {
		log.Panic(err)
		return
	}

	return token, err
}

//ValidateToken validates the jwt token
func ValidateToken(userType, signedToken string) (claims *SignedDetails, msg string) {
	var SECRET_KEY string
	if userType == "ADMIN" {
		SECRET_KEY = EnvVar("ADMIN_SECRET_KEY", "")
	} else {
		SECRET_KEY = EnvVar("SECRET_KEY", "")
	}
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "the token is invalid"
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "token is expired"
		msg = err.Error()
		return
	}

	return claims, msg
}

func IPInfo(ip string) string {
	client := ipinfo.NewClient(nil, nil, EnvVar("IPINFO_TOKEN", ""))
	info, err := client.GetIPCountryName(net.ParseIP(ip))
	if err != nil || info == "" {
		info = "unknown"
	}

	fmt.Println(info)

	return info
}
