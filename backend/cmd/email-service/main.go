package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/mailer"
	"github.com/HappyNaCl/new-debyut/internal/common/rabbitmq"
	topics "github.com/HappyNaCl/new-debyut/internal/constants"
	"github.com/HappyNaCl/new-debyut/internal/email/consumers"
)

func main() {
	if err := l.Init(); err != nil {
		panic("Failed to initialize logger.")
	}
	defer l.Sync()

	emailHelper, err := mailer.NewEmailHelper()
	if err != nil {
		panic("Failed to initialize email helper: " + err.Error())
	}

	conn, ch, err := rabbitmq.Connect()
	if err != nil {
		panic("Failed to connect to RabbitMQ: " + err.Error())
	}
	defer conn.Close()
	defer ch.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := consumers.NewWelcomeUserEmailConsumer(emailHelper)
	go func() {
		if err := rabbitmq.Consume(ctx, ch, "email-service.user-registered", topics.USER_REGISTERED_TOPIC, consumer.Handle); err != nil && ctx.Err() == nil {
			l.Logger.Warn("Consumer stopped: " + err.Error())
		}
	}()

	otpCh, err := conn.Channel()
	if err != nil {
		panic("Failed to open OTP consumer channel: " + err.Error())
	}
	defer otpCh.Close()

	otpConsumer := consumers.NewOtpEmailConsumer(emailHelper)
	go func() {
		if err := rabbitmq.Consume(ctx, otpCh, "email-service.otp-requested", topics.OTP_REQUESTED_TOPIC, otpConsumer.Handle); err != nil && ctx.Err() == nil {
			l.Logger.Warn("Consumer stopped: " + err.Error())
		}
	}()

	invitationCh, err := conn.Channel()
	if err != nil {
		panic("Failed to open invitation consumer channel: " + err.Error())
	}
	defer invitationCh.Close()

	invitationConsumer := consumers.NewInvitationEmailConsumer(emailHelper)
	go func() {
		if err := rabbitmq.Consume(ctx, invitationCh, "email-service.organization-invitation-created", topics.ORGANIZATION_INVITATION_CREATED_TOPIC, invitationConsumer.Handle); err != nil && ctx.Err() == nil {
			l.Logger.Warn("Consumer stopped: " + err.Error())
		}
	}()

	l.Logger.Info("Email service is ready and running!")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
}
