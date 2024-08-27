package sdk

import (
	"fmt"
	"log"

	"github.com/praetorian-inc/chariot-client/pkg/sdk/model"
)

type Service[T any] struct {
	Client *Chariot
}

func (s *Service[T]) KeyName() string {
	var t T
	switch any(t).(type) {
	case model.Asset:
		return "asset"
	case model.Attribute:
		return "attribute"
	case model.Risk:
		return "risk"
	case model.File:
		return "file"
	case model.Job:
		return "job"
	case model.Account:
		return "account"
	default:
		log.Fatal(fmt.Errorf("unsupported type %T", t))
		return ""
	}
}

type RiskService struct {
	Service[model.Risk]
}

func NewRiskService(client *Chariot) *RiskService {
	return &RiskService{
		Service: Service[model.Risk]{
			Client: client,
		},
	}
}

type AttributeService struct {
	Service[model.Attribute]
}

func NewAttributeService(client *Chariot) *AttributeService {
	return &AttributeService{
		Service: Service[model.Attribute]{
			Client: client,
		},
	}
}

type AssetService struct {
	Service[model.Asset]
}

func NewAssetService(client *Chariot) *AssetService {
	return &AssetService{
		Service: Service[model.Asset]{
			Client: client,
		},
	}
}

type FileService struct {
	Service[model.File]
}

func NewFileService(client *Chariot) *FileService {
	return &FileService{
		Service: Service[model.File]{
			Client: client,
		},
	}
}

type JobService struct {
	Service[model.Job]
}

func NewJobService(client *Chariot) *JobService {
	return &JobService{
		Service: Service[model.Job]{
			Client: client,
		},
	}
}

type AccountService struct {
	Service[model.Account]
}

func NewAccountService(client *Chariot) *AccountService {
	return &AccountService{
		Service: Service[model.Account]{
			Client: client,
		},
	}
}
