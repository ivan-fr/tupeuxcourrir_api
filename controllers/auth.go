package controllers

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
	"tupeuxcourrir_api/forms"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"
)

func jsonErrorFormat(err error) gin.H {
	var sliceStr []string
	if _, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range err.(validator.ValidationErrors) {
			sliceStr = append(sliceStr, fmt.Sprint(&utils.FieldError{Err: fieldErr}))
		}
	}

	if sliceStr == nil {
		return gin.H{"error": err.Error()}
	}

	return gin.H{"error": sliceStr}
}

func SignUp(context *gin.Context) {
	var form forms.SignUpForm
	var user models.User

	if err := context.ShouldBind(&form); err != nil {
		context.JSON(http.StatusBadRequest, jsonErrorFormat(err))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(form.EncryptedPassword), bcrypt.MinCost)
	if err != nil {
		context.JSON(http.StatusBadRequest, jsonErrorFormat(err))
		return
	}

	form.EncryptedPassword = string(hash)

	orm.BindForm(&user, &form)
	iQB := orm.GetInsertQueryBuilder(models.NewUser(), &user)

	if _, err := iQB.ApplyInsert(); err != nil {
		context.JSON(http.StatusBadRequest, jsonErrorFormat(err))
		return
	}

	context.JSON(http.StatusOK, gin.H{})
	return
}

func Login(ctx *gin.Context) {
	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"Email": ctx.PostForm("email")}))

	mapUser, err := sQB.ApplyQueryRow()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, jsonErrorFormat(err))
		return
	}

	user := mapUser["User"].(*models.User)

	if err = bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword),
		[]byte(ctx.PostForm("password"))); err != nil {
		ctx.JSON(http.StatusBadRequest, jsonErrorFormat(err))
		return
	}

	expirationTime := time.Now().Add(1 * time.Hour)

	claims := jwt.MapClaims{"userId": user.IdUser, "exp": expirationTime.Unix()}
	instantiateClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, errToken := instantiateClaims.SignedString([]byte("mySecret"))

	if errToken != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		log.Println(errToken)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}
