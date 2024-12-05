package adapter

import (
	"strings"
	"viola/voc/internal/model"

	"github.com/google/uuid"
)

// ConvertReviewResultToReviewResultDBModel 将ReviewResult 转变为ReviewResultDBModel数组
func ConvertReviewResultToReviewResultDBModel(reviewResults model.ReviewResult) []model.ReviewResultDBModel {
	var reviewResultDBModels []model.ReviewResultDBModel
	for _, firstReviewResult := range reviewResults.Review {
		for _, secondReview := range firstReviewResult.SecondLevel {
			reviewResultDBModels = append(reviewResultDBModels, model.ReviewResultDBModel{
				ID:                   uuid.New().String(),
				ProductName:          reviewResults.ProductName,
				FirstLevel:           firstReviewResult.FirstLevel,
				FirstLevelPercentage: firstReviewResult.Percentage,
				SecondLevel:          secondReview.Categorization,
				SecondPercentage:     secondReview.Percentage,
				SecondDetails:        strings.Join(secondReview.Details, ","),
			})
		}

	}
	return reviewResultDBModels
}
