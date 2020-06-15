package controllers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
	"tupeuxcourrir_api/forms"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"
)

func SignUp(context *gin.Context) {
	var form forms.SignUpForm
	var user models.User

	if err := context.ShouldBind(&form); err != nil {
		context.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(form.EncryptedPassword), bcrypt.MinCost)
	if err != nil {
		context.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
		return
	}

	form.EncryptedPassword = string(hash)

	orm.BindForm(&user, &form)
	iQB := orm.GetInsertQueryBuilder(models.NewUser(), &user)

	if _, err := iQB.ApplyInsert(); err != nil {
		context.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
		return
	}

	context.JSON(http.StatusOK, gin.H{})
	return
}

func Login(ctx *gin.Context) {
	var loginForm forms.LoginForm

	if err := ctx.ShouldBind(&loginForm); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
		return
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"Email": loginForm.Email}))

	mapUser, err := sQB.ApplyQueryRow()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
		return
	}

	user := mapUser["User"].(*models.User)

	if err = bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword),
		[]byte(loginForm.EncryptedPassword)); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
		return
	}

	var expirationTime time.Time
	if loginForm.SaveConnection {
		expirationTime = time.Now().Add(1 * time.Hour)
	} else {
		expirationTime = time.Now().Add(5 * time.Hour)
	}

	claims := jwt.MapClaims{"userId": user.IdUser, "expireAt": expirationTime.Unix()}
	instantiateClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, errToken := instantiateClaims.SignedString([]byte("mySecret"))

	if errToken != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		log.Println(errToken)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}
