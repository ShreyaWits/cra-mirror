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

	"github.com/gocql/gocql"
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

// For testing purposes only
type GenerateUrlRepositoryInterface interface {
	SaveGeneratedLink(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error)
	GetTokenData(link *string) (*apiDtos.SecurePayload, error)
	DeleteShortCode(link string) (*commonDtos.ApiResponseDto, error)
}

// GenerateLinkServiceForTest is a test-friendly version of GenerateLinkService
type GenerateLinkServiceForTest struct {
	Repository GenerateUrlRepositoryInterface
	Redis      *database.RedisConfig
	Cassandra  repository.ICassandraRepository
	OTPRepo    authRepository.IOTPRepository
}

// SaveGeneratedLink handles the generation and saving of protected links (for tests)
func (s *GenerateLinkServiceForTest) SaveGeneratedLink(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error) {
	const operation = "SaveGeneratedLink"
	log.Printf("%s %s: processing request for user %s", operationPrefix, operation, dto.UserID)
	return s.Repository.SaveGeneratedLink(dto)
}

// DeleteGeneratedLink handles the deletion of protected links (for tests)
func (s *GenerateLinkServiceForTest) DeleteGeneratedLink(link string) (*commonDtos.ApiResponseDto, error) {
	const operation = "DeleteGeneratedLink"
	log.Printf("%s %s: processing request for link %s", operationPrefix, operation, link)

	result, err := s.Repository.GetTokenData(&link)
	if err != nil {
		log.Printf("%s %s: failed to retrieve token data: %v", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to retrieve token data: %w", err)
	}

	if result.ModelType == "hybrid" {
		// Safe type assertion for hybrid model
		dbId, ok := result.Data.(string)
		if !ok || dbId == "" {
			log.Printf("%s %s: invalid or missing data for hybrid model", operationPrefix, operation)
			return nil, fmt.Errorf("invalid or missing data for hybrid model")
		}

		if s.Cassandra == nil {
			log.Printf("%s %s: Cassandra dependency is nil", operationPrefix, operation)
			return nil, fmt.Errorf("Cassandra dependency is not initialized")
		}

		if err := s.Cassandra.DeleteById(dbId); err != nil {
			log.Printf("%s %s: failed to delete from Cassandra: %v", operationPrefix, operation, err)
			return nil, fmt.Errorf("failed to delete from Cassandra: %w", err)
		}
	}

	return s.Repository.DeleteShortCode(link)
}

// GetExtractData handles the retrieval of data from protected links (for tests)
func (s *GenerateLinkServiceForTest) GetExtractData(link *string) (*commonDtos.ApiResponseDto, error) {
	const operation = "GetExtractData"
	log.Printf("%s %s: processing request for link %s", operationPrefix, operation, *link)

	// Validate dependencies
	if s.Repository == nil {
		log.Printf("%s %s: Repository dependency is nil", operationPrefix, operation)
		return nil, fmt.Errorf("repository dependency is not initialized")
	}

	result, err := s.Repository.GetTokenData(link)
	if err != nil {
		log.Printf("%s %s: failed to retrieve token data: %v", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to retrieve token data: %w", err)
	}

	if result.ModelType == "hybrid" {
		// Safe type assertion for hybrid model
		dbId, ok := result.Data.(string)
		if !ok || dbId == "" {
			log.Printf("%s %s: invalid or missing data for hybrid model", operationPrefix, operation)
			return nil, fmt.Errorf("invalid or missing data for hybrid model")
		}

		// Check if Cassandra is initialized
		if s.Cassandra == nil {
			log.Printf("%s %s: Cassandra dependency is nil", operationPrefix, operation)
			return nil, fmt.Errorf("Cassandra dependency is not initialized")
		}

		res, err := s.Cassandra.GetDataByID(dbId)
		if err != nil {
			log.Printf("%s %s: failed to retrieve data from Cassandra: %v", operationPrefix, operation, err)
			return nil, fmt.Errorf("failed to retrieve data from Cassandra: %w", err)
		}

		if res == nil {
			log.Printf("%s %s: no data found in Cassandra for ID %s", operationPrefix, operation, dbId)
			return nil, fmt.Errorf("no data found for ID %s", dbId)
		}

		if res.OtpRequired {
			// Check if Redis is initialized
			if s.Redis == nil {
				log.Printf("%s %s: Redis dependency is nil", operationPrefix, operation)
				return nil, fmt.Errorf("Redis dependency is not initialized")
			}

			// Check if Cassandra is initialized (again for clarity)
			if s.Cassandra == nil {
				log.Printf("%s %s: Cassandra dependency is nil", operationPrefix, operation)
				return nil, fmt.Errorf("Cassandra dependency is not initialized")
			}

			repo := authRepository.NewOTPRepository(s.Redis, s.Cassandra)
			if repo == nil {
				log.Printf("%s %s: failed to create OTP repository", operationPrefix, operation)
				return nil, fmt.Errorf("failed to create OTP repository")
			}

			services := srv.NewAuthenticationService(repo, s.Cassandra)
			if services == nil {
				log.Printf("%s %s: failed to create authentication service", operationPrefix, operation)
				return nil, fmt.Errorf("failed to create authentication service")
			}

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
		// Check if Redis is initialized
		if s.Redis == nil {
			log.Printf("%s %s: Redis dependency is nil", operationPrefix, operation)
			return nil, fmt.Errorf("Redis dependency is not initialized")
		}

		// Check if Cassandra is initialized
		if s.Cassandra == nil {
			log.Printf("%s %s: Cassandra dependency is nil", operationPrefix, operation)
			return nil, fmt.Errorf("Cassandra dependency is not initialized")
		}

		repo := s.OTPRepo
		if repo == nil {
			log.Printf("%s %s: failed to create OTP repository", operationPrefix, operation)
			return nil, fmt.Errorf("failed to create OTP repository")
		}

		services := srv.NewAuthenticationService(repo, s.Cassandra)
		if services == nil {
			log.Printf("%s %s: failed to create authentication service", operationPrefix, operation)
			return nil, fmt.Errorf("failed to create authentication service")
		}

		return services.SendOtp(dto, "")
	}

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: messageUtility.GetMessage(string(constants.DataFetchedSuccessfully)),
		Data:    result.Data,
	}, nil
}

// SaveGeneratedLink handles the generation and saving of protected links
func (s *GenerateLinkService) SaveGeneratedLink(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error) {
	const operation = "SaveGeneratedLink"
	log.Printf("%s %s: processing request for user %s", operationPrefix, operation, dto.UserID)

	// Validate repository dependency
	if s.repo == nil {
		log.Printf("%s %s: Repository dependency is nil", operationPrefix, operation)
		return nil, fmt.Errorf("repository dependency is not initialized")
	}

	return s.repo.SaveGeneratedLink(dto)
}

// DeleteGeneratedLink handles the deletion of protected links
func (s *GenerateLinkService) DeleteGeneratedLink(link string) (*commonDtos.ApiResponseDto, error) {
	const operation = "DeleteGeneratedLink"
	log.Printf("%s %s: processing request for link %s", operationPrefix, operation, link)

	// Validate repository dependency
	if s.repo == nil {
		log.Printf("%s %s: Repository dependency is nil", operationPrefix, operation)
		return nil, fmt.Errorf("repository dependency is not initialized")
	}

	result, err := s.repo.GetTokenData(&link)
	if err != nil {
		log.Printf("%s %s: failed to retrieve token data: %v", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to retrieve token data: %w", err)
	}

	if result.ModelType == "hybrid" {
		// Safe type assertion for hybrid model
		dbId, ok := result.Data.(string)
		if !ok || dbId == "" {
			log.Printf("%s %s: invalid or missing data for hybrid model", operationPrefix, operation)
			return nil, fmt.Errorf("invalid or missing data for hybrid model")
		}

		// Check if Cassandra is initialized
		if s.cassandra == nil {
			log.Printf("%s %s: Cassandra dependency is nil", operationPrefix, operation)
			return nil, fmt.Errorf("Cassandra dependency is not initialized")
		}

		if err := s.cassandra.DeleteById(dbId); err != nil {
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

	// Validate repository dependency
	if s.repo == nil {
		log.Printf("%s %s: Repository dependency is nil", operationPrefix, operation)
		return nil, fmt.Errorf("repository dependency is not initialized")
	}

	result, err := s.repo.GetTokenData(link)
	if err != nil {
		log.Printf("%s %s: failed to retrieve token data: %v", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to retrieve token data: %w", err)
	}

	if result.ModelType == "hybrid" {
		// Safe type assertion for hybrid model
		dbId, ok := result.Data.(string)
		if !ok || dbId == "" {
			log.Printf("%s %s: invalid or missing data for hybrid model", operationPrefix, operation)
			return nil, fmt.Errorf("invalid or missing data for hybrid model")
		}

		// Check if Cassandra is initialized
		if s.cassandra == nil {
			log.Printf("%s %s: Cassandra dependency is nil", operationPrefix, operation)
			return nil, fmt.Errorf("Cassandra dependency is not initialized")
		}

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
			// Check if Redis and Cassandra are initialized
			if s.redis == nil {
				log.Printf("%s %s: Redis dependency is nil", operationPrefix, operation)
				return nil, fmt.Errorf("Redis dependency is not initialized")
			}

			repo := authRepository.NewOTPRepository(s.redis, s.cassandra)
			if repo == nil {
				log.Printf("%s %s: failed to create OTP repository", operationPrefix, operation)
				return nil, fmt.Errorf("failed to create OTP repository")
			}

			services := srv.NewAuthenticationService(repo, s.cassandra)
			if services == nil {
				log.Printf("%s %s: failed to create authentication service", operationPrefix, operation)
				return nil, fmt.Errorf("failed to create authentication service")
			}

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
		// Check if Redis and Cassandra are initialized
		if s.redis == nil {
			log.Printf("%s %s: Redis dependency is nil", operationPrefix, operation)
			return nil, fmt.Errorf("Redis dependency is not initialized")
		}

		if s.cassandra == nil {
			log.Printf("%s %s: Cassandra dependency is nil", operationPrefix, operation)
			return nil, fmt.Errorf("Cassandra dependency is not initialized")
		}

		repo := authRepository.NewOTPRepository(s.redis, s.cassandra)
		if repo == nil {
			log.Printf("%s %s: failed to create OTP repository", operationPrefix, operation)
			return nil, fmt.Errorf("failed to create OTP repository")
		}

		services := srv.NewAuthenticationService(repo, s.cassandra)
		if services == nil {
			log.Printf("%s %s: failed to create authentication service", operationPrefix, operation)
			return nil, fmt.Errorf("failed to create authentication service")
		}

		return services.SendOtp(dto, "")
	}

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: messageUtility.GetMessage(string(constants.DataFetchedSuccessfully)),
		Data:    result.Data,
	}, nil
}

// For testing purposes only
func (s *GenerateLinkService) GetDependencies() (repo *repositories.GeneratedRepository, redis *database.RedisConfig, cassandra repository.ICassandraRepository) {
	return s.repo, s.redis, s.cassandra
}

// For testing purposes only - to improve coverage
func NewGenerateLinkServiceWithInterfaces(repo GenerateUrlRepositoryInterface, redis *database.RedisConfig, cassandra repository.ICassandraRepository) *GenerateLinkServiceForTest {
	return &GenerateLinkServiceForTest{
		Repository: repo,
		Redis:      redis,
		Cassandra:  cassandra,
	}
}

// For testing purposes only - direct access to real methods
func GetRealServiceMethods() (
	func(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error),
	func(link string) (*commonDtos.ApiResponseDto, error),
	func(link *string) (*commonDtos.ApiResponseDto, error)) {

	// Create a service with non-nil dependencies
	repo := &repositories.GeneratedRepository{}
	redis := &database.RedisConfig{}
	cassandra := &mockCassandraRepo{} // Using a simple mock implementation
	service := NewGenerateLinkService(repo, redis, cassandra)

	return service.SaveGeneratedLink, service.DeleteGeneratedLink, service.GetExtractData
}

// Simple mock implementation for tests
type mockCassandraRepo struct{}

func (m *mockCassandraRepo) DeleteById(id string) error {
	return nil
}

func (m *mockCassandraRepo) GetDataByID(id string) (*apiDtos.GenerateUrlRequest, error) {
	return &apiDtos.GenerateUrlRequest{}, nil
}

func (m *mockCassandraRepo) SaveData(data *apiDtos.GenerateUrlRequest) (gocql.UUID, error) {
	return gocql.UUID{}, nil
}
