package controllers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
)

func jsonErrorFormat(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func SignUp(context *gin.Context) {
	var form models.User

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

	iQB := orm.GetInsertQueryBuilder(models.NewUser(), &form)

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
		log.Fatal(errToken)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}
