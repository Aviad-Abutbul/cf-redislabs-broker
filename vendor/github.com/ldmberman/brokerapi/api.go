package brokerapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ldmberman/brokerapi/auth"
	"github.com/pivotal-golang/lager"
)

const provisionLogKey = "provision"
const updateLogKey = "update"
const deprovisionLogKey = "deprovision"
const bindLogKey = "bind"
const unbindLogKey = "unbind"
const lastOperationLogKey = "lastOperation"

const instanceIDLogKey = "instance-id"
const instanceDetailsLogKey = "instance-details"
const instanceUpdateDetailsLogKey = "instance-update-details"
const bindingIDLogKey = "binding-id"

const invalidServiceDetailsErrorKey = "invalid-service-details"
const invalidUpdateDetailsErrorKey = "invalid-update-details"
const invalidBindDetailsErrorKey = "invalid-bind-details"
const instanceLimitReachedErrorKey = "instance-limit-reached"
const instanceAlreadyExistsErrorKey = "instance-already-exists"
const bindingAlreadyExistsErrorKey = "binding-already-exists"
const instanceMissingErrorKey = "instance-missing"
const bindingMissingErrorKey = "binding-missing"
const asyncRequiredErrorKey = "async-required"
const unknownErrorKey = "unknown-error"

const statusUnprocessableEntity = 422

type BrokerCredentials struct {
	Username string
	Password string
}

func New(serviceBroker ServiceBroker, logger lager.Logger, brokerCredentials BrokerCredentials) http.Handler {
	router := newHttpRouter()

	router.Get("/v2/catalog", catalog(serviceBroker, router, logger))

	router.Put("/v2/service_instances/{instance_id}", provision(serviceBroker, router, logger))
	router.Patch("/v2/service_instances/{instance_id}", update(serviceBroker, router, logger))
	router.Delete("/v2/service_instances/{instance_id}", deprovision(serviceBroker, router, logger))
	router.Get("/v2/service_instances/{instance_id}/last_operation", lastOperation(serviceBroker, router, logger))

	router.Put("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", bind(serviceBroker, router, logger))
	router.Delete("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", unbind(serviceBroker, router, logger))

	return wrapAuth(router, brokerCredentials)
}

func wrapAuth(router httpRouter, credentials BrokerCredentials) http.Handler {
	return auth.NewWrapper(credentials.Username, credentials.Password).Wrap(router)
}

func catalog(serviceBroker ServiceBroker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		catalog := CatalogResponse{
			Services: serviceBroker.Services(),
		}

		respond(w, http.StatusOK, catalog)
	}
}

func provision(serviceBroker ServiceBroker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		instanceID := vars["instance_id"]

		logger := logger.Session(provisionLogKey, lager.Data{
			instanceIDLogKey: instanceID,
		})

		var details ProvisionDetails
		if err := json.NewDecoder(req.Body).Decode(&details); err != nil {
			logger.Error(invalidServiceDetailsErrorKey, err)
			respond(w, statusUnprocessableEntity, ErrorResponse{
				Description: err.Error(),
			})
			return
		}

		acceptsIncompleteFlag, _ := strconv.ParseBool(req.URL.Query().Get("accepts_incomplete"))

		logger = logger.WithData(lager.Data{
			instanceDetailsLogKey: details,
		})

		async, err := serviceBroker.Provision(instanceID, details, acceptsIncompleteFlag)

		if err != nil {
			switch err {
			case ErrInstanceAlreadyExists:
				logger.Error(instanceAlreadyExistsErrorKey, err)
				respond(w, http.StatusConflict, EmptyResponse{})
			case ErrInstanceLimitMet:
				logger.Error(instanceLimitReachedErrorKey, err)
				respond(w, http.StatusInternalServerError, ErrorResponse{
					Description: err.Error(),
				})
			case ErrAsyncRequired:
				logger.Error(instanceLimitReachedErrorKey, err)
				respond(w, 422, ErrorResponse{
					Error:       "AsyncRequired",
					Description: err.Error(),
				})
			default:
				logger.Error(unknownErrorKey, err)
				respond(w, http.StatusInternalServerError, ErrorResponse{
					Description: err.Error(),
				})
			}
			return
		}

		if async {
			respond(w, http.StatusAccepted, ProvisioningResponse{})
		} else {
			respond(w, http.StatusCreated, ProvisioningResponse{})
		}
	}
}

func update(serviceBroker ServiceBroker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		instanceID := vars["instance_id"]

		logger := logger.Session(updateLogKey, lager.Data{
			instanceIDLogKey: instanceID,
		})

		var details UpdateDetails
		if err := json.NewDecoder(req.Body).Decode(&details); err != nil {
			logger.Error(invalidUpdateDetailsErrorKey, err)
			respond(w, statusUnprocessableEntity, ErrorResponse{
				Description: err.Error(),
			})
			return
		}

		acceptsIncompleteFlag, _ := strconv.ParseBool(req.URL.Query().Get("accepts_incomplete"))

		logger = logger.WithData(lager.Data{
			instanceUpdateDetailsLogKey: details,
		})

		async, err := serviceBroker.Update(instanceID, details, acceptsIncompleteFlag)

		if err != nil {
			switch err {
			case ErrInstanceDoesNotExist:
				logger.Error(instanceMissingErrorKey, err)
				respond(w, http.StatusNotFound, EmptyResponse{})
			case ErrAsyncRequired:
				logger.Error(asyncRequiredErrorKey, err)
				respond(w, 422, ErrorResponse{
					Error:       "AsyncRequired",
					Description: err.Error(),
				})
			default:
				logger.Error(unknownErrorKey, err)
				respond(w, http.StatusInternalServerError, ErrorResponse{
					Description: err.Error(),
				})
			}
			return
		}

		if async {
			respond(w, http.StatusAccepted, UpdateResponse{})
		} else {
			respond(w, http.StatusOK, UpdateResponse{})
		}
	}
}

func deprovision(serviceBroker ServiceBroker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		instanceID := vars["instance_id"]
		logger := logger.Session(deprovisionLogKey, lager.Data{
			instanceIDLogKey: instanceID,
		})

		if err := serviceBroker.Deprovision(instanceID); err != nil {
			switch err {
			case ErrInstanceDoesNotExist:
				logger.Error(instanceMissingErrorKey, err)
				respond(w, http.StatusGone, EmptyResponse{})
			default:
				logger.Error(unknownErrorKey, err)
				respond(w, http.StatusInternalServerError, ErrorResponse{
					Description: err.Error(),
				})
			}
			return
		}

		respond(w, http.StatusOK, EmptyResponse{})
	}
}

func bind(serviceBroker ServiceBroker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		instanceID := vars["instance_id"]
		bindingID := vars["binding_id"]

		logger := logger.Session(bindLogKey, lager.Data{
			instanceIDLogKey: instanceID,
			bindingIDLogKey:  bindingID,
		})

		var details BindDetails
		if err := json.NewDecoder(req.Body).Decode(&details); err != nil {
			logger.Error(invalidBindDetailsErrorKey, err)
			respond(w, statusUnprocessableEntity, ErrorResponse{
				Description: err.Error(),
			})
			return
		}

		credentials, err := serviceBroker.Bind(instanceID, bindingID, details)
		if err != nil {
			switch err {
			case ErrInstanceDoesNotExist:
				logger.Error(instanceMissingErrorKey, err)
				respond(w, http.StatusNotFound, ErrorResponse{
					Description: err.Error(),
				})
			case ErrBindingAlreadyExists:
				logger.Error(bindingAlreadyExistsErrorKey, err)
				respond(w, http.StatusConflict, ErrorResponse{
					Description: err.Error(),
				})
			default:
				logger.Error(unknownErrorKey, err)
				respond(w, http.StatusInternalServerError, ErrorResponse{
					Description: err.Error(),
				})
			}
			return
		}

		bindingResponse := BindingResponse{
			Credentials: credentials,
		}

		respond(w, http.StatusCreated, bindingResponse)
	}
}

func unbind(serviceBroker ServiceBroker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		instanceID := vars["instance_id"]
		bindingID := vars["binding_id"]

		logger := logger.Session(unbindLogKey, lager.Data{
			instanceIDLogKey: instanceID,
			bindingIDLogKey:  bindingID,
		})

		if err := serviceBroker.Unbind(instanceID, bindingID); err != nil {
			switch err {
			case ErrInstanceDoesNotExist:
				logger.Error(instanceMissingErrorKey, err)
				respond(w, http.StatusNotFound, EmptyResponse{})
			case ErrBindingDoesNotExist:
				logger.Error(bindingMissingErrorKey, err)
				respond(w, http.StatusGone, EmptyResponse{})
			default:
				logger.Error(unknownErrorKey, err)
				respond(w, http.StatusInternalServerError, ErrorResponse{
					Description: err.Error(),
				})
			}
			return
		}

		respond(w, http.StatusOK, EmptyResponse{})
	}
}

func respond(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

func lastOperation(serviceBroker ServiceBroker, router httpRouter, logger lager.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := router.Vars(req)
		instanceID := vars["instance_id"]

		logger := logger.Session(lastOperationLogKey, lager.Data{
			instanceIDLogKey: instanceID,
		})

		logger.Info("starting-check-for-operation")

		lastOperation, err := serviceBroker.LastOperation(instanceID)

		if err != nil {
			switch err {
			case ErrInstanceDoesNotExist:
				logger.Error(instanceMissingErrorKey, err)
				respond(w, http.StatusNotFound, ErrorResponse{
					Description: err.Error(),
				})
			default:
				logger.Error(unknownErrorKey, err)
				respond(w, http.StatusInternalServerError, ErrorResponse{
					Description: err.Error(),
				})
			}

			return
		}

		logger.WithData(lager.Data{"state": lastOperation.State}).Info("done-check-for-operation")

		lastOperationResponse := LastOperationResponse{
			State:       string(lastOperation.State),
			Description: lastOperation.Description,
		}

		respond(w, http.StatusOK, lastOperationResponse)
	}
}
