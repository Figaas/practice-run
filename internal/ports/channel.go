package ports

import (
	"context"

	"practice-run/internal/message"
)

type MessageChannel interface {
	SendMessage(msg message.Message)
	RespondOK(ctx context.Context)
	RespondError(ctx context.Context, err error)
	Close()
}
