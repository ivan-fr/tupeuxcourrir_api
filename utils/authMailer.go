package utils

import "net/smtp"

func GetAuthMailer() smtp.Auth {
	return smtp.PlainAuth("", "tupeuxcourrir@gmail.com", "ujmjgievlrwcaqxo", "smtp.gmail.com")
}
