package controllers

import "TaskManager/internal/models"

func GenerateToken(user *models.User) (string, error) {
	return "", nil
}

func ValidateToken(tokenString string) (*models.User, error) {
	return nil, nil
}
