package helpers

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"facebookapi/config"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

var jwtKey []byte

// SetJWTKey stores the secret key for signing tokens
func SetJWTKey(key string) {
	jwtKey = []byte(key)
}

// GetJWTKey returns the JWT secret
func GetJWTKey() []byte {
	return jwtKey
}

// ValidateToken parses and validates a JWT token string
func ValidateToken(tokenString string) (*Claims, error) {
	secretKey := GetJWTKey()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateToken creates access + refresh token
func GenerateToken(email, userID, userType string) (string, string, error) {
	accessExpiry := time.Now().Add(24 * time.Hour).Unix()
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour).Unix()

	claims := &Claims{
		Email:  email,
		UserID: userID,
		Role:   userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: accessExpiry,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedAccess, err := accessToken.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	refreshClaims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: refreshExpiry,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	return signedAccess, signedRefresh, nil
}

// UpdateAllToken updates the token fields in the MongoDB user document
func UpdateAllToken(signedToken, signedRefreshToken, userid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	userCollection := config.OpenCollection("users")

	updateObj := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "token", Value: signedToken},
			{Key: "refresh_token", Value: signedRefreshToken},
			{Key: "updated_at", Value: time.Now()},
		}},
	}

	filter := bson.M{"user_id": userid}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		updateObj,
		options.Update().SetUpsert(true),
	)

	if err != nil {
		log.Printf("Error updating tokens for user %s: %v", userid, err)
	}
	return err
}

// Hash password safely
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Check password against hash
func CheckPassword(password, hashed string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil, err
}
