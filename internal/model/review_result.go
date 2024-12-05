package model

// {\"product_name\":\"阴影盘\",\n\"comments\":[{\n\"first-level\":\"物流问题\",\n\"percentage\":\"10%\",\n\"second-level\":[{\"categorization\":\"不送货上门\",\"percentage\":\"1%\"}]}]}"
type ReviewResult struct {
	//ID          string `dynamodbav:"id" json:"id,omitempty"`
	ProductName string `dynamodbav:"product_name" json:"product_name,omitempty"`
	Review      []struct {
		FirstLevel  string `dynamodbav:"first_level" json:"first_level,omitempty" `
		Percentage  string `dynamodbav:"percentage" json:"percentage,omitempty" `
		SecondLevel []struct {
			Categorization string   `dynamodbav:"categorization" json:"categorization,omitempty"`
			Percentage     string   `dynamodbav:"percentage" json:"percentage,omitempty" `
			Details        []string `dynamodbav:"details" json:"details,omitempty" `
		} `dynamodbav:"second_level" json:"second_level,omitempty" `
	} `dynamodbav:"comments,omitempty" json:"comments,omitempty"`
}

type ReviewResultDBModel struct {
	ID                   string `dynamodbav:"id" json:"id"`
	ProductName          string `dynamodbav:"product_name" json:"product_name,omitempty"`
	FirstLevel           string `dynamodbav:"first_level" json:"first_level,omitempty" `
	FirstLevelPercentage string `dynamodbav:"first_level_percentage" json:"first_level_percentage,omitempty" `
	SecondLevel          string `dynamodbav:"second_level" json:"second_level,omitempty"`
	SecondPercentage     string `dynamodbav:"second_percentage" json:"second_percentage,omitempty"`
	SecondDetails        string `dynamodbav:"second_details" json:"second_details,omitempty"`
}
