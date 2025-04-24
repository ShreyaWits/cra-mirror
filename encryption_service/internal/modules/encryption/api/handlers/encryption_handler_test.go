package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mocks "encryption_microservice/internal/common/mocks"
	pb "encryption_microservice/internal/common/proto_gen"
	dtos "encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/internal/modules/encryption/api/handlers"
	mapper "encryption_microservice/internal/modules/encryption/api/mapper"
	user "encryption_microservice/internal/modules/encryption/services/user"
	errors "encryption_microservice/pkg/errors"
)

func TestEncryptionHandlerImpl_GenerateEDEK(t *testing.T) {
	tests := []struct {
		name            string
		token           string
		getUserDataFunc func(token string) (*user.User, *errors.CustomError)
		genEDEKResp     *dtos.GenerateEDEKResponse
		genEDEKErr      *errors.CustomError
		updateUserErr   *errors.CustomError
		expectedResp    *pb.GenerateEDEKResponse
		expectedCode    codes.Code
		expectedMsg     string
	}{
		{
			name:  "success",
			token: "tkn",
			getUserDataFunc: func(token string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1"}, nil
			},
			genEDEKResp:   &dtos.GenerateEDEKResponse{EDEKPrivate: "p", EDEKPublic: "q"},
			genEDEKErr:    nil,
			updateUserErr: nil,
			expectedResp:  &pb.GenerateEDEKResponse{EDEKPrivate: "p", EDEKPublic: "q"},
			expectedCode:  codes.OK,
		},
		{
			name:  "unauthenticated",
			token: "",
			getUserDataFunc: func(token string) (*user.User, *errors.CustomError) {
				return nil, errors.NewCustomError(errors.USRErrTokenRequired, fmt.Errorf("token missing"))
			},
			expectedResp: nil,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "token missing",
		},
		{
			name:  "usecase error",
			token: "tkn",
			getUserDataFunc: func(token string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1"}, nil
			},
			genEDEKResp:  nil,
			genEDEKErr:   errors.NewCustomError(errors.ESErrGenerateEDEK, fmt.Errorf("uc err")),
			expectedResp: nil,
			expectedCode: codes.Internal,
			expectedMsg:  "uc err",
		},
		{
			name:  "update error",
			token: "tkn",
			getUserDataFunc: func(token string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1"}, nil
			},
			genEDEKResp:   &dtos.GenerateEDEKResponse{EDEKPrivate: "p", EDEKPublic: "q"},
			genEDEKErr:    nil,
			updateUserErr: errors.NewCustomError(errors.USRErrUpdateUser, fmt.Errorf("upd err")),
			expectedResp:  nil,
			expectedCode:  codes.Internal,
			expectedMsg:   "upd err",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockEncryptionUseCase(ctrl)
			mockUserSvc := mocks.NewMockUserService(ctrl)
			ud, ue := tc.getUserDataFunc(tc.token)
			mockUserSvc.EXPECT().GetUserData(tc.token).Return(ud, ue)
			if ue == nil {
				mockUsecase.EXPECT().
					GenerateEDEK(gomock.Any(), ud.ID).
					Return(tc.genEDEKResp, tc.genEDEKErr)
				if tc.genEDEKErr == nil {
					mockUserSvc.EXPECT().
						UpdateUser(ud.ID, tc.genEDEKResp.EDEKPrivate, tc.genEDEKResp.EDEKPublic).
						Return(tc.updateUserErr)
				}
			}
			handler := handlers.NewEncryptionHandler(mockUsecase, mockUserSvc)
			resp, err := handler.GenerateEDEK(context.Background(), &pb.GenerateEDEKRequest{Token: tc.token})

			if tc.expectedResp != nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp.GetEDEKPrivate() != tc.expectedResp.EDEKPrivate || resp.GetEDEKPublic() != tc.expectedResp.EDEKPublic {
					t.Errorf(
						"expected EDEKPrivate=%q, EDEKPublic=%q; got EDEKPrivate=%q, EDEKPublic=%q",
						tc.expectedResp.EDEKPrivate, tc.expectedResp.EDEKPublic,
						resp.GetEDEKPrivate(), resp.GetEDEKPublic(),
					)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				st, _ := status.FromError(err)
				if st.Code() != tc.expectedCode {
					t.Errorf("expected code %v, got %v", tc.expectedCode, st.Code())
				}
			}
		})
	}
}

func TestEncryptionHandlerImpl_Encrypt(t *testing.T) {
	tests := []struct {
		name            string
		token           string
		getUserDataFunc func(string) (*user.User, *errors.CustomError)
		encResp         *dtos.EncryptResponse
		encErr          *errors.CustomError
		expectedData    []map[string]string
		expectedCode    codes.Code
		expectedMsg     string
	}{
		{
			name:  "success",
			token: "tkn",
			getUserDataFunc: func(string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1", EDEKPrivate: "x", EDEKPublic: "y"}, nil
			},
			encResp:      &dtos.EncryptResponse{Data: []map[string]string{{"k": "v"}}},
			expectedData: []map[string]string{{"k": "v"}},
			expectedCode: codes.OK,
		},
		{
			name:  "unauthenticated",
			token: "",
			getUserDataFunc: func(string) (*user.User, *errors.CustomError) {
				return nil, errors.NewCustomError(errors.USRErrTokenRequired, fmt.Errorf("token missing"))
			},
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "token missing",
		},
		{
			name:  "usecase error",
			token: "tkn",
			getUserDataFunc: func(string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1", EDEKPrivate: "x", EDEKPublic: "y"}, nil
			},
			encErr:       errors.NewCustomError(errors.ESErrEncrypt, fmt.Errorf("enc err")),
			expectedCode: codes.Internal,
			expectedMsg:  "enc err",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockEncryptionUseCase(ctrl)
			mockUserSvc := mocks.NewMockUserService(ctrl)
			ud, ue := tc.getUserDataFunc(tc.token)
			mockUserSvc.EXPECT().GetUserData(tc.token).Return(ud, ue)
			if ue == nil {
				mockUsecase.EXPECT().
					Encrypt(gomock.Any(), ud.ID, ud.EDEKPrivate, ud.EDEKPublic, gomock.Any()).
					Return(tc.encResp, tc.encErr)
			}
			handler := handlers.NewEncryptionHandler(mockUsecase, mockUserSvc)
			dataPB := mapper.ConvertToStructPB([]map[string]string{{"k": "v"}})
			resp, err := handler.Encrypt(context.Background(), &pb.EncryptDataRequest{Token: tc.token, Data: dataPB})

			if tc.expectedData != nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				got := mapper.ConvertFromStructPB(resp.GetData())
				if !reflect.DeepEqual(tc.expectedData, got) {
					t.Errorf("expected %v, got %v", tc.expectedData, got)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				st, _ := status.FromError(err)
				if st.Code() != tc.expectedCode {
					t.Errorf("expected code %v, got %v", tc.expectedCode, st.Code())
				}
			}
		})
	}
}

func TestEncryptionHandlerImpl_Decrypt(t *testing.T) {
	tests := []struct {
		name            string
		token           string
		getUserDataFunc func(string) (*user.User, *errors.CustomError)
		decResp         *dtos.DecryptResponse
		decErr          *errors.CustomError
		expectedData    []map[string]string
		expectedCode    codes.Code
		expectedMsg     string
	}{
		{
			name:  "success",
			token: "tkn",
			getUserDataFunc: func(string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1", EDEKPrivate: "x", EDEKPublic: "y"}, nil
			},
			decResp:      &dtos.DecryptResponse{Data: []map[string]string{{"k": "v"}}},
			expectedData: []map[string]string{{"k": "v"}},
			expectedCode: codes.OK,
		},
		{
			name:  "unauthenticated",
			token: "",
			getUserDataFunc: func(string) (*user.User, *errors.CustomError) {
				return nil, errors.NewCustomError(errors.USRErrTokenRequired, fmt.Errorf("token missing"))
			},
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "token missing",
		},
		{
			name:  "usecase error",
			token: "tkn",
			getUserDataFunc: func(string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1", EDEKPrivate: "x", EDEKPublic: "y"}, nil
			},
			decErr:       errors.NewCustomError(errors.ESErrDecrypt, fmt.Errorf("dec err")),
			expectedCode: codes.Internal,
			expectedMsg:  "dec err",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockEncryptionUseCase(ctrl)
			mockUserSvc := mocks.NewMockUserService(ctrl)
			ud, ue := tc.getUserDataFunc(tc.token)
			mockUserSvc.EXPECT().GetUserData(tc.token).Return(ud, ue)
			if ue == nil {
				mockUsecase.EXPECT().
					Decrypt(gomock.Any(), ud.ID, ud.EDEKPrivate, ud.EDEKPublic, gomock.Any()).
					Return(tc.decResp, tc.decErr)
			}
			handler := handlers.NewEncryptionHandler(mockUsecase, mockUserSvc)
			dataPB := mapper.ConvertToStructPB([]map[string]string{{"k": "v"}})
			resp, err := handler.Decrypt(context.Background(), &pb.DecryptRequest{Token: tc.token, Data: dataPB})

			if tc.expectedData != nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				got := mapper.ConvertFromStructPB(resp.GetData())
				if !reflect.DeepEqual(tc.expectedData, got) {
					t.Errorf("expected %v, got %v", tc.expectedData, got)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				st, _ := status.FromError(err)
				if st.Code() != tc.expectedCode {
					t.Errorf("expected code %v, got %v", tc.expectedCode, st.Code())
				}
			}
		})
	}
}

func TestEncryptionHandlerImpl_HandleHealth(t *testing.T) {
	app := fiber.New()
	handler := handlers.NewEncryptionHandler(nil, nil)
	app.Get("/health", handler.HandleHealth)
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %v", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var m map[string]string
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if m["status"] != "healthy" || m["version"] != "1.0.0" {
		t.Errorf("unexpected body: %v", m)
	}
}
