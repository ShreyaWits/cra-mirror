package services

import (
	"fmt"
	"log"

	commonDtos "protected_link/internal/common/api/dtos"

	"protected_link/internal/common/constants"
	messageUtility "protected_link/internal/common/utils"
	authRepository "protected_link/internal/modules/authentication/repositories"
	srv "protected_link/internal/modules/authentication/services"
	repository "protected_link/internal/modules/cassandra/repository"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/internal/modules/link_generation/repositories"
	link_utils "protected_link/internal/modules/link_generation/utils"

	database "protected_link/pkg/redis"
)

const (
	operationPrefix = "🔒"
)

type GenerateLinkService struct {
	repo      *repositories.GeneratedRepository
	redis     *database.RedisConfig
	cassandra repository.ICassandraRepository
}

// NewGenerateLinkService creates a new instance of GenerateLinkService
func NewGenerateLinkService(repo *repositories.GeneratedRepository, redis *database.RedisConfig, cassandra repository.ICassandraRepository) *GenerateLinkService {
	return &GenerateLinkService{
		repo:      repo,
		redis:     redis,
		cassandra: cassandra,
	}
}

// SaveGeneratedLink handles the generation and saving of protected links
func (s *GenerateLinkService) SaveGeneratedLink(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error) {
	const operation = "SaveGeneratedLink"
	log.Printf("%s %s: processing request for user %s", operationPrefix, operation, dto.UserID)
	return s.repo.SaveGeneratedLink(dto)
}

// DeleteGeneratedLink handles the deletion of protected links
func (s *GenerateLinkService) DeleteGeneratedLink(link string) (*commonDtos.ApiResponseDto, error) {
	const operation = "DeleteGeneratedLink"
	log.Printf("%s %s: processing request for link %s", operationPrefix, operation, link)

	result, err := s.repo.GetTokenData(&link)
	if err != nil {
		log.Printf("%s %s: failed to retrieve token data: %v", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to retrieve token data: %w", err)
	}

	if result.ModelType == "hybrid" {
		if err := s.cassandra.DeleteById(result.Data.(string)); err != nil {
			log.Printf("%s %s: failed to delete from Cassandra: %v", operationPrefix, operation, err)
			return nil, fmt.Errorf("failed to delete from Cassandra: %w", err)
		}
	}

	return s.repo.DeleteShortCode(link)
}

// GetExtractData handles the retrieval of data from protected links
func (s *GenerateLinkService) GetExtractData(link *string) (*commonDtos.ApiResponseDto, error) {
	const operation = "GetExtractData"
	log.Printf("%s %s: processing request for link %s", operationPrefix, operation, *link)

	result, err := s.repo.GetTokenData(link)
	if err != nil {
		log.Printf("%s %s: failed to retrieve token data: %v", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to retrieve token data: %w", err)
	}

	if result.ModelType == "hybrid" {
		dbId := result.Data.(string)
		res, err := s.cassandra.GetDataByID(dbId)
		if err != nil {
			log.Printf("%s %s: failed to retrieve data from Cassandra: %v", operationPrefix, operation, err)
			return nil, fmt.Errorf("failed to retrieve data from Cassandra: %w", err)
		}

		if res == nil {
			log.Printf("%s %s: no data found in Cassandra for ID %s", operationPrefix, operation, dbId)
			return nil, fmt.Errorf("no data found for ID %s", dbId)
		}

		if res.OtpRequired {
			repo := authRepository.NewOTPRepository(s.redis, s.cassandra)
			services := srv.NewAuthenticationService(repo, s.cassandra)
			return services.SendOtp(res, dbId)
		}

		return &commonDtos.ApiResponseDto{
			Success: true,
			Message: messageUtility.GetMessage(string(constants.DataFetchedSuccessfully)),
			Data:    res,
		}, nil
	}

	dto, err := link_utils.ConvertToGenerateUrlRequest(result.Data)
	if err != nil {
		log.Printf("%s %s: failed to convert data: %v", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to convert data: %w", err)
	}

	if dto.OtpRequired {
		repo := authRepository.NewOTPRepository(s.redis, s.cassandra)
		services := srv.NewAuthenticationService(repo, s.cassandra)
		return services.SendOtp(dto, "")
	}

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: messageUtility.GetMessage(string(constants.DataFetchedSuccessfully)),
		Data:    result.Data,
	}, nil
}
