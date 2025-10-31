package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DBHost                    string
	DBPort                    string
	DBUser                    string
	DBPassword                string
	DBName                    string
	DBSSLMode                 string
	ServerPort                string
	JWTSecret                 string
	NotificationServiceURL    string
	NotificationCheckInterval string
}

func Load() *Config {
	loadEnvFile(".env")

	return &Config{
		DBHost:                    getEnv("DB_HOST", "localhost"),
		DBPort:                    getEnv("DB_PORT", "5432"),
		DBUser:                    getEnv("DB_USER", "postgres"),
		DBPassword:                getEnv("DB_PASSWORD", "password"),
		DBName:                    getEnv("DB_NAME", "taskManager"),
		DBSSLMode:                 getEnv("DB_SSL_MODE", "disable"),
		ServerPort:                getEnv("SERVER_PORT", "8080"),
		JWTSecret:                 getEnv("JWT_SECRET", "your-default-secret-key"),
		NotificationServiceURL:    getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service-app:8081"),
		NotificationCheckInterval: getEnv("NOTIFICATION_CHECK_INTERVAL", "1m"),
	}
}

func (c *Config) GetConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func (c *Config) GetServerPortString() string {
	return fmt.Sprintf(":%s",
		c.ServerPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

/*func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}*/

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		//fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Разделяем ключ и значение
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Убираем кавычки если есть
		value = strings.Trim(value, `"'`)

		// Устанавливаем переменную окружения (только если еще не установлена)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
