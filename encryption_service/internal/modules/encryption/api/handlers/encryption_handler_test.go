package handlers_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

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
			tokenUserData, tokenErr := tc.getUserDataFunc(tc.token)
			mockUserSvc.EXPECT().GetUserData(tc.token).Return(tokenUserData, tokenErr)
			if tokenErr == nil {
				mockUsecase.EXPECT().
					GenerateEDEK(gomock.Any(), tokenUserData.ID).
					Return(tc.genEDEKResp, tc.genEDEKErr)
				if tc.genEDEKErr == nil {
					mockUserSvc.EXPECT().
						UpdateUser(tokenUserData.ID, tc.genEDEKResp.EDEKPrivate, tc.genEDEKResp.EDEKPublic).
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
		name             string
		token            string
		getUserDataFunc  func(string) (*user.User, *errors.CustomError)
		getUserDataFunc2 func(string) (*user.User, *errors.CustomError)
		encResp          *dtos.EncryptResponse
		encErr           *errors.CustomError
		expectedData     []map[string]string
		expectedCode     codes.Code
		expectedMsg      string
		userId           *string
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
		{
			name:  "userid success",
			token: "tkn",
			getUserDataFunc: func(string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1", EDEKPrivate: "x", EDEKPublic: "y"}, nil
			},
			getUserDataFunc2: func(string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u2", EDEKPrivate: "x2", EDEKPublic: "y2"}, nil
			},
			encResp:      &dtos.EncryptResponse{Data: []map[string]string{{"k": "v"}}},
			expectedData: []map[string]string{{"k": "v"}},
			expectedCode: codes.OK,
			userId:       &[]string{"invalid"}[0],
		},
		{
			name:  "userId Error",
			token: "tkn",
			getUserDataFunc: func(string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1", EDEKPrivate: "x", EDEKPublic: "y"}, nil
			},
			getUserDataFunc2: func(string) (*user.User, *errors.CustomError) {
				return nil, errors.NewCustomError(errors.USRErrFetchUserData, fmt.Errorf("user not found"))
			},
			encErr:       errors.NewCustomError(errors.ESErrEncrypt, fmt.Errorf("enc err")),
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "enc err",
			userId:       &[]string{"valid"}[0],
		},
		{
			name:  "panic",
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
			tokenUserData, tokenErr := tc.getUserDataFunc(tc.token)

			if tc.name == "panic" {
				mockUserSvc.EXPECT().GetUserData(tc.token).Do(func(token string) {
					panic("panic")
				}).Return(nil, nil)
			} else {
				mockUserSvc.EXPECT().GetUserData(tc.token).Return(tokenUserData, tokenErr)
			}
			if tc.userId != nil {
				ud2, ue2 := tc.getUserDataFunc2(*tc.userId)
				mockUserSvc.EXPECT().GetUserData(*tc.userId).Return(ud2, ue2)
				if ue2 == nil {
					mockUsecase.EXPECT().
						Encrypt(gomock.Any(), ud2.ID, ud2.EDEKPrivate, ud2.EDEKPublic, gomock.Any()).
						Return(tc.encResp, tc.encErr)
				}
			} else {
				if tokenErr == nil && tc.name != "panic" {
					mockUsecase.EXPECT().
						Encrypt(gomock.Any(), tokenUserData.ID, tokenUserData.EDEKPrivate, tokenUserData.EDEKPublic, gomock.Any()).
						Return(tc.encResp, tc.encErr)
				}
			}

			handler := handlers.NewEncryptionHandler(mockUsecase, mockUserSvc)
			dataPB := mapper.ConvertToStructPB([]map[string]string{{"k": "v"}})
			resp, err := handler.Encrypt(context.Background(), &pb.EncryptDataRequest{Token: tc.token, Data: dataPB, UserId: tc.userId})

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
		name             string
		token            string
		getUserDataFunc  func(string) (*user.User, *errors.CustomError)
		decResp          *dtos.DecryptResponse
		decErr           *errors.CustomError
		expectedData     []map[string]string
		expectedCode     codes.Code
		expectedMsg      string
		getUserDataFunc2 func(string) (*user.User, *errors.CustomError)
		userId           *string
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
		{
			name:   "userid",
			token:  "tkn",
			userId: &[]string{"valid"}[0],
			getUserDataFunc: func(string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1", EDEKPrivate: "x", EDEKPublic: "y"}, nil
			},
			getUserDataFunc2: func(string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u2", EDEKPrivate: "x2", EDEKPublic: "y2"}, nil
			},
			decResp:      &dtos.DecryptResponse{Data: []map[string]string{{"k": "v"}}},
			expectedData: []map[string]string{{"k": "v"}},
			expectedCode: codes.OK,
		},
		{
			name:  "userid error",
			token: "tkn",
			getUserDataFunc: func(string) (*user.User, *errors.CustomError) {
				return &user.User{ID: "u1", EDEKPrivate: "x", EDEKPublic: "y"}, nil
			},
			getUserDataFunc2: func(string) (*user.User, *errors.CustomError) {
				return nil, errors.NewCustomError(errors.USRErrFetchUserData, fmt.Errorf("user not found"))
			},
			userId:       &[]string{"invalid"}[0],
			decErr:       errors.NewCustomError(errors.ESErrDecrypt, fmt.Errorf("dec err")),
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "dec err",
		},
		{
			name:  "panic",
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
			tokenUserData, tokenErr := tc.getUserDataFunc(tc.token)
			if tc.name == "panic" {
				mockUserSvc.EXPECT().GetUserData(tc.token).Do(func(token string) {
					panic("panic")
				})
			} else {
				mockUserSvc.EXPECT().GetUserData(tc.token).Return(tokenUserData, tokenErr)
			}

			if tc.userId != nil {
				ud2, ue2 := tc.getUserDataFunc2(*tc.userId)
				mockUserSvc.EXPECT().GetUserData(*tc.userId).Return(ud2, ue2)
				if ue2 == nil {
					mockUsecase.EXPECT().
						Decrypt(gomock.Any(), ud2.ID, ud2.EDEKPrivate, ud2.EDEKPublic, gomock.Any()).
						Return(tc.decResp, tc.decErr)
				}
			} else {
				if tokenErr == nil && tc.name != "panic" {
					mockUsecase.EXPECT().
						Decrypt(gomock.Any(), tokenUserData.ID, tokenUserData.EDEKPrivate, tokenUserData.EDEKPublic, gomock.Any()).
						Return(tc.decResp, tc.decErr)
				}
			}
			handler := handlers.NewEncryptionHandler(mockUsecase, mockUserSvc)
			dataPB := mapper.ConvertToStructPB([]map[string]string{{"k": "v"}})
			resp, err := handler.Decrypt(context.Background(), &pb.DecryptRequest{Token: tc.token, Data: dataPB, UserId: tc.userId})

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

func TestEncryptionHandlerImpl_Health(t *testing.T) {
	tests := []struct {
		name         string
		decResp      *dtos.DecryptResponse
		expectedData []map[string]string
		expectedCode codes.Code
	}{
		{
			name:         "success",
			decResp:      &dtos.DecryptResponse{Data: []map[string]string{{"k": "v"}}},
			expectedData: []map[string]string{{"k": "v"}},
			expectedCode: codes.OK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockEncryptionUseCase(ctrl)
			mockUserSvc := mocks.NewMockUserService(ctrl)

			handler := handlers.NewEncryptionHandler(mockUsecase, mockUserSvc)
			_, err := handler.HealthCheck(context.Background(), &pb.HealthCheckRequest{})

			if tc.expectedData != nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
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
