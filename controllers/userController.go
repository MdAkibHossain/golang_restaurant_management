package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	database "restaurant-management/database"
	"restaurant-management/helpers"
	"restaurant-management/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		recordePerPage, err := strconv.Atoi(c.Query("recordePerPage"))
		if err != nil || recordePerPage < 1 {
			recordePerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}
		startIndex := (page - 1) * recordePerPage

		startIndex, err = strconv.Atoi(c.Query("startIndex"))
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}

		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordePerPage}}}},
			}}}

		cursor, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, projectStage,
		})
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("error occured while listing user items")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		var allUsers []bson.M
		if err := cursor.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allUsers[0])

	}
}
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User

		userId := c.Param("user_id")

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)

		defer cancel()

		if err != nil {
			msg := fmt.Sprintf("error occured while listing user items")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
func Sugnup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		//convert the JSON data coming from postman to something that golang understands
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//validate the data based on user struct
		if valivatorErr := validate.Struct(&user); valivatorErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": valivatorErr.Error()})
			return
		}

		//you'll check if the email has already been used by another user
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}

		//hash password
		password := HashPassword(*user.Password)
		user.Password = &password

		//you'll also check if the phone no. has already been used by another user
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number alresdy exist"})
			return
		}

		//create some extra details for the user object - created_at, updated_at, ID
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		//generate token and refersh token (generate all tokens function from helper)
		token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.Phone)
		user.Token = &token
		user.Refresh_Token = &refreshToken

		//if all ok, then you insert this new user into the user collection
		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()

		//return status OK and send the result back
		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		//convert the login data from postman which is in JSON to golang readable format
		defer cancel()
		if err := c.BindJSON(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//find a user with that email and see if that user even exists
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "This email id not correct"})
			return
		}

		//then you will verify the password
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		//if all goes well, then you'll generate tokens
		token, refreshToken, err := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		//update tokens - token and refersh token
		helpers.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		//return statusOK
		c.JSON(http.StatusOK, foundUser)

	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("paassword is incorrect")
		check = false
	}
	return check, msg

}
