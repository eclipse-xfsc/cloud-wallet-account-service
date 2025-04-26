package model

import (
	"errors"
	"gorm.io/gorm"
	"regexp"
	"time"
)

const SqlstateUniqueViolationState = "23505"

var PresentationAlreadyExistsError = errors.New("presentation request already exists")

type PresentationRequest struct {
	*gorm.Model
	UserId         string `gorm:"user_id"`
	RequestId      string `gorm:"request_id"`
	ProofRequestId string `gorm:"proof_request_id"`
	TTL            int    `gorm:"ttl"`
}

func GetPresentationRequestById(db *gorm.DB, userId string, requestId string) (*PresentationRequest, error) {
	var request PresentationRequest
	err := setSchema(db).
		Where("user_id=? AND request_id=?", userId, requestId).
		First(&request).Error
	return &request, err
}

func CreatePresentationRequestDBEntry(db *gorm.DB, userId string, requestId string, ttl int, proofRequestId string) error {
	request := PresentationRequest{UserId: userId, RequestId: requestId, TTL: ttl, ProofRequestId: proofRequestId}
	sq := setSchema(db).Create(&request)
	if sq.Error == nil {
		return nil
	}
	re := regexp.MustCompile(`SQLSTATE (\d+)`)
	match := re.FindStringSubmatch(sq.Error.Error())
	if len(match) > 1 && match[1] == SqlstateUniqueViolationState {
		return PresentationAlreadyExistsError
	}
	return sq.Error
}

func GetAllPresentationRequests(db *gorm.DB, userId string) ([]PresentationRequest, error) {
	var requests []PresentationRequest
	err := setSchema(db).
		Where("user_id=?", userId).
		Find(&requests).Error
	if err == nil {
		res := make([]PresentationRequest, 0)
		expIds := make([]string, 0)
		for _, request := range requests {
			isExpired := request.TTL != 0 && int(time.Now().Sub(request.CreatedAt).Seconds()) > request.TTL
			if isExpired {
				expIds = append(expIds, request.RequestId)
			} else {
				res = append(res, request)
			}
		}
		er := DeletePresentationRequests(db, userId, expIds)
		if er != nil {
			logger.Error(er, "could not delete expired presentation requests",
				"count", len(expIds), "ids", expIds)
		}
		return res, nil
	}
	return requests, err
}

func DeletePresentationRequests(db *gorm.DB, userId string, requestIds []string) error {
	return setSchema(db).
		Where("user_id=? AND request_id IN ?", userId, requestIds).
		Delete(&PresentationRequest{}).Error
}
