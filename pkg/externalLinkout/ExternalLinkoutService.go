/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package externalLinkout

import (
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"time"
)

type ExternalLinkoutService interface {
	Create(request *ExternalLinkoutRequest) (*ExternalLinkoutRequest, error)
	GetAllActiveTools() ([]ExternalLinksMonitoringToolsRequest, error)
	FetchAllActiveLinks(clusterIds []int) ([]ExternalLinkoutRequest, error)
	Update(request *ExternalLinkoutRequest) (*ExternalLinkoutRequest, error)

	Delete(request *ExternalLinkoutRequest) (*ExternalLinkoutRequest, error)

	GetAllActiveLinks(request *ExternalLinkoutRequest) (*ExternalLinkoutRequest, error)
}
type ExternalLinkoutServiceImpl struct {
	logger                          *zap.SugaredLogger
	externalLinkoutToolsRepository  ExternalLinkoutToolsRepository
	externalLinksClustersRepository ExternalLinksClustersRepository
	externalLinksRepository         ExternalLinksRepository
	userAuthService                 user.UserAuthService
}
type ExternalLinksMonitoringToolsRequest struct {
	Id   int    `json:"Id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}
type ExternalLinkoutRequest struct {
	Name             string `json:"name"`
	Url              string `json:"url"`
	IsActive         bool   `json:"is_active"`
	MonitoringToolId int    `json:"monitoringToolId"`
	ClusterId        int    `json:"cluster_id"`
}

type ExternalLinkoutRequest2 struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	Url              string `json:"url"`
	IsActive         bool   `json:"is_active"`
	MonitoringToolId int    `json:"monitoringToolId"`
	ClusterId        []int  `json:"cluster_id"`
}

func NewExternalLinkoutServiceImpl(logger *zap.SugaredLogger, externalLinkoutToolsRepository ExternalLinkoutToolsRepository,
	externalLinksClustersRepository ExternalLinksClustersRepository, externalLinksRepository ExternalLinksRepository, userAuthService user.UserAuthService) *ExternalLinkoutServiceImpl {
	return &ExternalLinkoutServiceImpl{
		logger:                          logger,
		externalLinkoutToolsRepository:  externalLinkoutToolsRepository,
		externalLinksClustersRepository: externalLinksClustersRepository,
		externalLinksRepository:         externalLinksRepository,
		userAuthService:                 userAuthService,
	}
}

func (impl ExternalLinkoutServiceImpl) Create(request *ExternalLinkoutRequest2) (*ExternalLinkoutRequest2, error) {
	impl.logger.Debugw("external linkout create request", "req", request)
	t := &ExternalLinks{
		Name:             request.Name,
		IsActive:         request.IsActive,
		MonitoringToolId: request.MonitoringToolId,
		Url:              request.Url,
		AuditLog:         sql.AuditLog{CreatedOn: time.Now(), UpdatedOn: time.Now()},
	}
	err := impl.externalLinksRepository.Save(t)

	if err != nil {
		impl.logger.Errorw("error in saving team", "data", t, "err", err)
		err = &util.ApiError{
			InternalMessage: "external link failed to create in db",
			UserMessage:     "external link failed to create in db",
		}
		return nil, err
	}

	for _, v := range request.ClusterId {
		x := &ExternalLinksClusters{
			ExternalLinkId: t.Id,
			ClusterId:      v,
			AuditLog:       sql.AuditLog{CreatedOn: time.Now(), UpdatedOn: time.Now()},
		}
		err := impl.externalLinksClustersRepository.Save(x)

		if err != nil {
			impl.logger.Errorw("error in saving cluster id's", "data", t, "err", err)
			err = &util.ApiError{
				InternalMessage: "cluster id failed to create in db",
				UserMessage:     "cluster id failed to create in db",
			}
			return nil, err
		}
	}
	return request, nil
}

func (impl ExternalLinkoutServiceImpl) GetAllActiveTools() ([]ExternalLinksMonitoringToolsRequest, error) {
	impl.logger.Debug("fetch all team from db")
	tools, err := impl.externalLinkoutToolsRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetch all tools", "err", err)
		return nil, err
	}
	var toolRequests []ExternalLinksMonitoringToolsRequest
	for _, tool := range tools {
		providerRes := ExternalLinksMonitoringToolsRequest{
			Id:   tool.Id,
			Name: tool.Name,
			Icon: tool.Icon,
		}
		toolRequests = append(toolRequests, providerRes)
	}
	return toolRequests, err
}

func (impl ExternalLinkoutServiceImpl) FetchAllActiveLinks(clusterId int) ([]ExternalLinkoutRequest, error) {
	impl.logger.Debug("fetch all team from db")
	links, err := impl.externalLinksClustersRepository.FindAllActive(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	var linkRequests []ExternalLinkoutRequest
	for _, link := range links {
		providerRes := ExternalLinkoutRequest{
			Name:      link.ExternalLinks.Name,
			Url:       link.ExternalLinks.Url,
			IsActive:  link.ExternalLinks.IsActive,
			ClusterId: link.ClusterId,
		}
		linkRequests = append(linkRequests, providerRes)
	}
	return linkRequests, err
}

func (impl ExternalLinkoutServiceImpl) Update(request *ExternalLinkoutRequest2) (*ExternalLinkoutRequest2, error) {
	impl.logger.Debugw("team update request", "req", request)
	existingProvider, err0 := impl.externalLinksRepository.FindOne(request.Id)
	if err0 != nil {
		impl.logger.Errorw("No matching entry found for update.", "id", request.Id)
		err0 = &util.ApiError{
			Code:            constants.GitProviderUpdateProviderNotExists,
			InternalMessage: "external links update failed, does not exist",
			UserMessage:     "external links update failed, does not exist",
		}
		return nil, err0
	}
	link := &ExternalLinks{
		Name:             request.Name,
		IsActive:         request.IsActive,
		MonitoringToolId: request.MonitoringToolId,
		Url:              request.Url,
		AuditLog:         sql.AuditLog{CreatedOn: existingProvider.CreatedOn, UpdatedOn: time.Now()},
	}
	err := impl.externalLinksRepository.Update(link)
	if err != nil {
		impl.logger.Errorw("error in updating team", "data", link, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderUpdateFailedInDb,
			InternalMessage: "team failed to update in db",
			UserMessage:     "team failed to update in db",
		}
		return nil, err
	}
	request.Id = link.Id
	return request, nil
}

func (impl ExternalLinkoutServiceImpl) Delete(request *ExternalLinkoutRequest) (*ExternalLinkoutRequest, error) {

	return request, nil
}
