package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"TaskManager/internal/models"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
	topic  string
}

// NewKafkaProducer создает нового Kafka producer
func NewKafkaProducer(brokers string, topic string) (*KafkaProducer, error) {
	brokerList := strings.Split(brokers, ",")

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokerList...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	// Проверяем соединение, написав тестовое сообщение
	log.Printf("Kafka producer инициализирован для топика: %s, brokers: %v", topic, brokerList)

	return &KafkaProducer{
		writer: writer,
		topic:  topic,
	}, nil
}

// SendNotification отправляет уведомление в Kafka
func (kp *KafkaProducer) SendNotification(notification *models.Notification) error {
	// Сериализуем notification в JSON
	jsonData, err := json.Marshal(*notification)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга notification: %v", err)
	}

	// Создаем сообщение Kafka с task_id в качестве ключа
	message := kafka.Message{
		Key:   []byte(notification.Task_id),
		Value: jsonData,
	}

	// Отправляем сообщение в Kafka
	err = kp.writer.WriteMessages(context.Background(), message)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения в Kafka: %v", err)
	}

	log.Printf("Сообщение отправлено в Kafka для задачи: %s", notification.Task_id)
	return nil
}

// Close закрывает connection к Kafka
func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}
