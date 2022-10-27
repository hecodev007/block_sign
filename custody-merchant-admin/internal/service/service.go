package service

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/base"
)

type Service struct {
	UserJWT *domain.JwtCustomClaims
	Base    *base.CasbinService
}

func (s *Service) SetUserJWTInfo(user *domain.JwtCustomClaims) {
	s.UserJWT = user
}

func (s *Service) GetUserJWTInfo(user *domain.JwtCustomClaims) *Service {
	s.UserJWT = user
	return s
}

func (s *Service) CasbinCreate(param *domain.CasbinCreateRequest) error {
	for _, v := range param.CasbinInfos {
		err := s.Base.CasbinCreate(param.UserId, v.Path, v.Method)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) CasbinList(param *domain.CasbinListRequest) [][]string {
	return s.Base.CasbinList(param.RoleID)
}

func (s *Service) CasbinRemove(param *domain.CasbinCreateRequest) error {
	for _, v := range param.CasbinInfos {
		err := s.Base.CasbinRemove(param.UserId, v.Path, v.Method)
		if err != nil {
			return err
		}
	}
	return nil
}
