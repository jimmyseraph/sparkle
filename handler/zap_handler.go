package handler

import (
	"github.com/jimmyseraph/sparkle/engine"
	"github.com/jimmyseraph/sparkle/utils/logger"

	"go.uber.org/zap"
)

var _ engine.MessageHandler = (*zapHandler)(nil)

type zapHandler struct {
	handler *zap.SugaredLogger
}

func NewZapHandler() *zapHandler {
	handler := logger.NewLogger("", "info")
	return &zapHandler{handler: handler}
}

func (h *zapHandler) Send(assertion *engine.Assertion) {
	for _, detail := range assertion.GetDetails() {
		h.handler.Infof("%s: %s, %v", detail.Name, detail.Message, detail.RecordTime)
	}

}
