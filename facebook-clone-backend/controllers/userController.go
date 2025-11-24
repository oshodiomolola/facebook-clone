package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"facebookapi/config"
	"facebookapi/helpers"
	"facebookapi/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var validate = validator.New()

// --------------------------- SIGNUP ---------------------------

func Signup(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	userCollection := config.OpenCollection("users")

	var user models.User

	// Parse JSON body
	if err := c.BindJSON(&user); err != nil {
		log.Println("Error parsing JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("Received JSON:", user)

	// Validate struct
	if err := validate.Struct(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email or phone exists
	count, err := userCollection.CountDocuments(ctx, bson.M{
		"$or": []bson.M{
			{"email": user.Email},
			{"phone_number": user.PhoneNumber},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email or phone number already exists"})
		return
	}

	// Hash Password
	hashedPwd, err := helpers.HashPassword(*user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
		return
	}
	user.Password = &hashedPwd

	// Set timestamps & IDs
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.ID = primitive.NewObjectID()
	user.UserID = user.ID.Hex()

	// Generate tokens
	accessToken, refreshToken, err := helpers.GenerateToken(*user.Email, user.UserID, *user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	user.Token = &accessToken
	user.RefreshToken = &refreshToken

	// Save to DB
	_, insertErr := userCollection.InsertOne(ctx, user)
	if insertErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": insertErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signup success", "UserID": user.UserID})
}

// --------------------------- LOGIN ---------------------------

func Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	userCollection := config.OpenCollection("users")

	var loginInput struct {
		Identifier string `json:"identifier"` // email or phone
		Password   string `json:"password"`
	}

	var foundUser models.User

	// Parse user input
	if err := c.BindJSON(&loginInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email OR phone number
	err := userCollection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"email": loginInput.Identifier},
			{"phone_number": loginInput.Identifier},
		},
	}).Decode(&foundUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or phone or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Compare password
	passwordValid, _ := helpers.CheckPassword(loginInput.Password, *foundUser.Password)
	if !passwordValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or phone or password"})
		return
	}

	// Generate new tokens
	accessToken, refreshToken, err := helpers.GenerateToken(*foundUser.Email, foundUser.UserID, *foundUser.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	// Update tokens in DB
	err = helpers.UpdateAllToken(accessToken, refreshToken, foundUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"user":          foundUser,
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}
