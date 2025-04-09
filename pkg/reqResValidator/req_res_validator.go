package req_res_validator

import (
	"fmt"
	"log/slog"
	"net/http"

	libopenapi "github.com/pb33f/libopenapi"
	libopenapiValidator "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi-validator/errors"
)

type ReqResValidator struct {
	validator libopenapiValidator.Validator
	logger    *slog.Logger
}

/*
oapi - OpenAPI document from file
*/
func New(oapi []byte, logger *slog.Logger) (*ReqResValidator, error) {
	op := "req_res_validator.NewReqResValidator"
	document, docErrs := libopenapi.NewDocument(oapi)

	log := logger.With("op", op)

	if docErrs != nil {
		log.Error(op, "error creating document", slog.String("error", docErrs.Error()))
		return nil, docErrs
	}

	highLevelValidator, validatorErrs := libopenapiValidator.NewValidator(document)

	// Create a new Validator
	if len(validatorErrs) > 0 {
		log.Error("op: %s,error creating validator", slog.Any("error", validatorErrs))
		return nil, fmt.Errorf("error creating validator: %v", validatorErrs)
	}

	return &ReqResValidator{
		validator: highLevelValidator,
		logger:    logger,
	}, nil
}

func MustNew(oapi []byte, logger *slog.Logger) *ReqResValidator {
	validator, err := New(oapi, logger)
	if err != nil {
		panic(err)
	}
	return validator
}

func (r *ReqResValidator) ValidateRequest(req *http.Request) (bool, []*errors.ValidationError) {
	op := "req_res_validator.ValidateRequest"
	log := r.logger.With("op", op)
	requestValid, validationErrors := r.validator.ValidateHttpRequest(req)

	if !requestValid {
		for i := range validationErrors {
			log.Debug(op, "request is failed", slog.String("error", validationErrors[i].Message), "endpoint", req.URL.String())
		}

		return false, validationErrors
	}

	return true, nil
}

func (r *ReqResValidator) ValidateResponse(req *http.Request, res *http.Response) (bool, []*errors.ValidationError) {
	op := "req_res_validator.ValidateResponse"
	log := r.logger.With("op", op)
	responseValid, validationErrors := r.validator.ValidateHttpRequestResponse(req, res)

	if !responseValid {
		for i := range validationErrors {
			log.Debug(op, "response is failed", slog.String("error", validationErrors[i].Message), "endpoint", req.URL.String())
		}

		return false, validationErrors
	}

	return true, nil
}
