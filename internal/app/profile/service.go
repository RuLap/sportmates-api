package profile

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/RuLap/sportmates-api/internal/app/refdata"
	"github.com/RuLap/sportmates-api/internal/pkg/errors"
	"github.com/RuLap/sportmates-api/internal/pkg/storage/minio"
	"github.com/google/uuid"
)

type Service interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*GetProfileResponse, error)
	SaveProfile(ctx context.Context, req *SaveProfileRequest, id *uuid.UUID) (*GetProfileResponse, error)
	GetAvatarUploadURL(ctx context.Context, userID uuid.UUID) (*GetUploadURLResponse, error)
	ConfirmAvatarUpload(ctx context.Context, userID uuid.UUID) (*ConfirmUploadAvatarResponse, error)
}

type service struct {
	log            *slog.Logger
	minio          *minio.Service
	bucketName     string
	repo           Repository
	refdataService refdata.Service
}

func NewService(log *slog.Logger, minio *minio.Service, repo Repository, refdataService refdata.Service) Service {
	return &service{
		log:            log,
		minio:          minio,
		bucketName:     "trackmus_avatars",
		repo:           repo,
		refdataService: refdataService,
	}
}

func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*GetProfileResponse, error) {
	profile, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	sports, err := s.getUserSports(ctx, id)
	if err != nil {
		return nil, err
	}

	city, err := s.getCity(ctx, int(profile.CityID))
	if err != nil {
		return nil, err
	}

	profileDTO := ProfileToGetResponse(profile, city, sports)

	return profileDTO, nil
}

func (s *service) SaveProfile(ctx context.Context, req *SaveProfileRequest, id *uuid.UUID) (*GetProfileResponse, error) {
	profile, err := SaveRequestToProfile(req)
	if err != nil {
		return nil, err
	}

	var result *Profile
	if id == nil {
		newID, err := uuid.Parse(req.ID)
		if err != nil {
			return nil, err
		}
		profile.ID = newID

		result, err = s.repo.Create(ctx, profile, req.Sports)
		if err != nil {
			return nil, err
		}
	} else {
		profile.ID = *id

		result, err = s.repo.Update(ctx, profile, req.Sports)
		if err != nil {
			return nil, err
		}
	}

	sports, err := s.getUserSports(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	city, err := s.getCity(ctx, int(profile.CityID))
	if err != nil {
		return nil, err
	}

	response := ProfileToGetResponse(result, city, sports)

	return response, nil
}

func (s *service) GetAvatarUploadURL(ctx context.Context, userID uuid.UUID) (*GetUploadURLResponse, error) {
	s3key := userID.String()

	avatarURL, err := s.minio.GenerateUploadURL(ctx, s.bucketName, s3key)
	if err != nil {
		s.log.Error("failed to generate upload url", "objName", s3key)
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	return &GetUploadURLResponse{
		URL: avatarURL,
	}, nil
}

func (s *service) ConfirmAvatarUpload(ctx context.Context, userID uuid.UUID) (*ConfirmUploadAvatarResponse, error) {
	s3key := userID.String()

	avatarURL, err := s.minio.GenerateUploadURL(ctx, s.bucketName, s3key)
	if err != nil {
		s.log.Error("failed to generate upload url", "objName", s3key)
		return nil, fmt.Errorf(errors.ErrFailedToSaveData)
	}

	return &ConfirmUploadAvatarResponse{
		URL: avatarURL,
	}, nil
}

func (s *service) getDownloadAvatarURL(ctx context.Context, userID uuid.UUID) (string, error) {
	s3key := userID.String()
	downloadUrl, err := s.minio.GenerateDownloadURL(ctx, s.bucketName, s3key, "avatar")
	if err != nil {
		s.log.Error("failed to generate download URL",
			"objName", s3key,
			"error", err,
		)
		return "", fmt.Errorf(errors.ErrFailedToLoadData)
	}

	return downloadUrl, nil
}

func (s *service) getUserSports(ctx context.Context, userID uuid.UUID) ([]refdata.GetSportResponse, error) {
	userSports, err := s.repo.GetUserSports(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(userSports) == 0 {
		return []refdata.GetSportResponse{}, nil
	}

	var sportIDs []string
	for _, us := range userSports {
		sportIDs = append(sportIDs, us.SportID.String())
	}

	sports, err := s.refdataService.GetSportsByIDs(ctx, sportIDs)
	if err != nil {
		return nil, err
	}

	sportDTOs := []refdata.GetSportResponse{}
	for _, s := range sports {
		sportDTOs = append(sportDTOs, *s)
	}

	return sportDTOs, nil
}

func (s *service) getCity(ctx context.Context, cityID int) (*refdata.GetCityResponse, error) {
	city, err := s.refdataService.GetCityByID(ctx, cityID)
	if err != nil {
		return nil, err
	}

	return city, nil
}
