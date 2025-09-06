package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/raynine/go-chatroom/models"
)

var INT_ZERO_VALUE = 0

func EncodeErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	var statusCode int
	message := ""

	errBody, ok := err.(*models.CustomError)
	if !ok {
		statusCode = http.StatusInternalServerError
		message = err.Error()
	} else {
		statusCode = errBody.Code
		message = errBody.Message

	}

	if statusCode == INT_ZERO_VALUE {
		statusCode = http.StatusInternalServerError
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
		"context": errBody.AppContext,
	})
}

func EncodeResponse(w http.ResponseWriter, model models.ServerResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(model.Code)
	json.NewEncoder(w).Encode(model.Data)
}

func DecodePayload(r *http.Request, model any) error {
	return json.NewDecoder(r.Body).Decode(&model)
}

func GetUserDataFromContext(ctx context.Context) (int, string, error) {
	appContext := "GetUserDataFromContext"

	userID, ok := ctx.Value("user_id").(int)
	if !ok {
		return -1, "", &models.CustomError{
			Message:    "User not provided",
			Code:       http.StatusBadRequest,
			AppContext: appContext,
		}
	}

	userName, ok := ctx.Value("user_user_name").(string)
	if !ok {
		return -1, "", &models.CustomError{
			Message:    "User not provided",
			Code:       http.StatusBadRequest,
			AppContext: appContext,
		}
	}
	return userID, userName, nil
}

func CreateJWTToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":        user.Id,
		"user_email":     user.Email,
		"user_user_name": user.Username,
		"exp":            time.Now().Add(time.Minute * 30).Unix(),
	}

	secretKey := os.Getenv("SECRET_KEY")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token := strings.Replace(authorization, "Bearer ", "", 1)

		authToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}

			secretKey := os.Getenv("SECRET_KEY")

			return []byte(secretKey), nil
		})
		if err != nil {
			log.Println("Error parsing JWT: ", err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !authToken.Valid {
			log.Println("Auth token is not valid")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, ok := authToken.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("Auth token is not a valid JWT Claims struct")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claimUserId, ok := claims["user_id"].(float64)
		if !ok {
			log.Println("Claim do not include user_id")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userId := int(claimUserId)

		userName, ok := claims["user_user_name"].(string)
		if !ok {
			log.Println("Claim do not include user_name")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userId)
		ctx = context.WithValue(ctx, "user_user_name", userName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
