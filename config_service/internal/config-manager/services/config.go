package services

import (
	"nps-config-service/internal/config-manager/repositories"
)

func StoreConfigService(env string, service string, req map[string]interface{}) (interface{} ,error) {

	response, err := repositories.StoreConfig(env, service, req)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func GetConfigService(service string, env string) (interface{}, error){

	response, err := repositories.GetConfig(service, env)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func GetConfigValueService(serviceName string, env string, key string) (interface{}, error){
	response, err := repositories.GetConfigValue(serviceName, env, key)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func GetConfigMetadataService(serviceName string, env string) (interface{}, error){
		response, err := repositories.GetConfigMetadata(serviceName, env)

	if err != nil {
		return nil, err
	}

	return response, nil
}