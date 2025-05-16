package handlers

import (
	"context"
	// "encoding/json"
	"fmt"

	// "fmt"
	"log"
	"notification-service/internal/common/api/dtos"
	"notification-service/internal/common/services"
	"notification-service/pkg/logger"

	"notification-service/proto"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type GRPCServer struct {
	proto.UnimplementedNotificationServiceServer
	svc services.NotificationServiceInterface
}

type NotificationServiceInterface interface {
	SaveConfig(ctx context.Context, configs []dtos.ChannelConfig) error
	GetConfigure() (*proto.GetConfigureResponse, error)
}

var validate = validator.New()

func init() {
	dtos.RegisterValidations(validate) // Register custom validation rules (e.g., phone validation)
}

func NewGRPCServer(svc services.NotificationServiceInterface) *GRPCServer {
	if svc == nil {
		log.Fatal("NotificationService cannot be nil")
	}
	return &GRPCServer{svc: svc}
}

func (g *GRPCServer) Configure(ctx context.Context, req *proto.ConfigureRequest) (*proto.ConfigureResponse, error) {
	tracer := otel.Tracer("notification-service")
	ctx, span := tracer.Start(ctx, "Configure")
	defer span.End()

	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	logger.Log.WithFields(logrus.Fields{
		"trace_id": traceID,
		"span_id":  spanID,
	}).Info("gRPC Configure post API called")

	// 🌟 Record incoming payload
	span.SetAttributes(attribute.String("request.configs", fmt.Sprintf("%v", req.Configs)))

	var configs []dtos.ChannelConfig
	var err error

	// 🛠️ Sub-span: Convert Proto to DTO
	ctxConvert, spanConvert := tracer.Start(ctx, "ConvertProtoToDTOs")
	for _, cfg := range req.Configs {
		configs = append(configs, dtos.ChannelConfig{
			Service:  cfg.Service,
			Primary:  cfg.Primary,
			Fallback: cfg.Fallback,
		})
	}
	spanConvert.SetAttributes(attribute.Int("converted.configs.count", len(configs)))
	spanConvert.End()
	ctx = ctxConvert

	// 🛠️ Sub-span: Validate Configs
	ctxValidate, spanValidate := tracer.Start(ctx, "ValidateConfigs")
	if err = validate.Var(configs, "required,dive"); err != nil {
		validationErr := err.(validator.ValidationErrors)[0]

		spanValidate.RecordError(err)
		spanValidate.SetStatus(codes.Error, "Validation Failed")
		spanValidate.SetAttributes(attribute.String("validation.error.field", validationErr.Field()))
		spanValidate.End()

		logger.Log.WithFields(logrus.Fields{
			"trace_id": traceID,
			"span_id":  spanID,
			"error":    err.Error(),
		}).Error("Validation error in Configure")

		fieldErr := &proto.FieldError{
			Field:     validationErr.Field(),
			ErrorMsg:  validationErr.Error(),
			ErrorCode: "VLD014",
		}

		return &proto.ConfigureResponse{
			Status:     "error",
			StatusCode: 400,
			Message:    "Validation failed",
			FieldError: fieldErr,
			Data:       "",
		}, nil
	}
	spanValidate.End()
	ctx = ctxValidate

	// 🛠️ Sub-span: Save Config
	ctxSave, spanSave := tracer.Start(ctx, "SaveConfig")
	if err := g.svc.SaveConfig(ctxSave, configs); err != nil {
		spanSave.RecordError(err)
		spanSave.SetStatus(codes.Error, "SaveConfig Failed")
		spanSave.End()

		logger.Log.WithFields(logrus.Fields{
			"trace_id": traceID,
			"span_id":  spanID,
			"error":    err.Error(),
		}).Error("Error while saving config")

		return &proto.ConfigureResponse{
			Status:     "error",
			StatusCode: 500,
			Message:    "Failed to save configuration",
			FieldError: &proto.FieldError{
				Field:     "configs",
				ErrorMsg:  "Internal error while saving configuration",
				ErrorCode: "SRV001",
			},
			Data: "",
		}, nil
	}
	spanSave.End()
	ctx = ctxSave

	logger.Log.WithFields(logrus.Fields{
		"trace_id": traceID,
		"span_id":  spanID,
	}).Info("Configuration saved successfully")

	// 🌟 Record outgoing response
	span.SetAttributes(attribute.String("response.message", "Channel configuration saved successfully"))
	span.SetStatus(codes.Ok, "Configuration updated successfully")

	return &proto.ConfigureResponse{
		Status:     "success",
		StatusCode: 200,
		Message:    "Configuration updated",
		FieldError: nil,
		Data:       "Channel configuration saved successfully",
	}, nil
}

func (g *GRPCServer) GetConfigure(ctx context.Context, req *proto.ConfigureRequest) (*proto.GetConfigureResponse, error) {
	tracer := otel.Tracer("notification-service")
	ctx, span := tracer.Start(ctx, "GetConfigure")
	defer span.End()

	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	logger.Log.WithFields(logrus.Fields{
		"trace_id": traceID,
		"span_id":  spanID,
	}).Info("Started GetConfigure tracing")

	// 🌟 Record incoming payload
	span.SetAttributes(attribute.String("request", fmt.Sprintf("%v", req)))

	// 🛠️ Sub-span: Call service layer
	ctxService, spanService := tracer.Start(ctx, "CallServiceLayerGetConfigure")
	resp, err := g.svc.GetConfigure()
	if err != nil {
		spanService.RecordError(err)
		spanService.SetStatus(codes.Error, err.Error())
		spanService.End()

		logger.Log.WithFields(logrus.Fields{
			"trace_id": traceID,
			"span_id":  spanID,
			"error":    err.Error(),
		}).Error("Error in GetConfigure service call")

		return nil, err
	}
	spanService.End()
	ctx = ctxService

	logger.Log.WithFields(logrus.Fields{
		"trace_id": traceID,
		"span_id":  spanID,
		"response": resp,
	}).Info("GetConfigure response fetched successfully")

	// 🌟 Record outgoing response
	span.SetAttributes(attribute.String("response", fmt.Sprintf("%v", resp)))
	span.SetStatus(codes.Ok, "GetConfigure fetched successfully")

	return resp, nil
}
