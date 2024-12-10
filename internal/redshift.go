package internal

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"viola/voc/internal/model"
	"viola/voc/utils"

	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
)

var RedShiftClient *redshiftdata.Client

type RedshiftClient interface {
	ExecuteStatement(ctx context.Context, params *redshiftdata.ExecuteStatementInput, optFns ...func(*redshiftdata.Options)) (*redshiftdata.ExecuteStatementOutput, error)
}

// InsertReviewResultAsync 异步插入数据到 Redshift
func InsertReviewResultAsync(
	ctx context.Context,
	workgroupName string,
	databaseName string,
	data []model.Review,
) chan Result {
	resultChan := make(chan Result, 1)

	go func() {
		defer close(resultChan)
		result := processReviewResults(ctx, workgroupName, databaseName, data)
		resultChan <- result
	}()

	return resultChan
}

// Result 存储处理结果
type Result struct {
	Count int
	Error error
}

func BatchInsertWithValues(ctx context.Context, workgroupName, databaseName string, data []model.Review) error {
	const (
		batchSize     = 50 // 每组100条数据
		maxGoroutines = 10 // 最大并发数
	)

	// 创建错误通道和等待组
	errChan := make(chan error, len(data)/batchSize+1)
	var wg sync.WaitGroup

	// 创建信号量来控制并发数
	semaphore := make(chan struct{}, maxGoroutines)

	// 按批次处理数据
	for i := 0; i < len(data); i += batchSize {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}

		batch := data[i:end]

		go func(batch []model.Review) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			// 构建批量插入的值
			values := make([]string, len(batch))
			for j, review := range batch {
				// 处理特殊字符
				productType := strings.ReplaceAll(review.ProductType, "'", "''")
				reviewContent := strings.ReplaceAll(review.ReviewContent, "'", "''")

				values[j] = fmt.Sprintf("('%s', '%s', '%s', '%s')",
					review.Id,
					productType,
					time.Unix(review.ReviewDate, 0).Format(time.RFC3339),
					reviewContent,
				)
			}

			// 构建SQL语句
			sqlStatement := fmt.Sprintf(`
                INSERT INTO reviews 
                (id, product_name, review_date, review_content)
                VALUES %s
            `, strings.Join(values, ","))

			// 执行插入
			input := &redshiftdata.ExecuteStatementInput{
				WorkgroupName: aws.String(workgroupName),
				Database:      aws.String(databaseName),
				Sql:           aws.String(sqlStatement),
			}

			selectInput, err := RedShiftClient.ExecuteStatement(ctx, input)
			fmt.Println("DBUser:", *selectInput.DbUser)
			if err != nil {
				errChan <- fmt.Errorf("batch insert error: %v", err)
				return
			}

			fmt.Printf("Successfully inserted batch of %d records,reqID:%s\n", len(batch), selectInput.Id)
		}(batch)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集所有错误
	var errors []error
	for err := range errChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	// 如果有错误，返回组合的错误信息
	if len(errors) > 0 {
		return fmt.Errorf("multiple errors occurred: %v", errors)
	}

	return nil
}

func processReviewResults(
	ctx context.Context,
	workgroupName string,
	databaseName string,
	data []model.Review,
) Result {
	var (
		count int
		wg    sync.WaitGroup
		mu    sync.Mutex
		errs  []error
	)

	semaphore := make(chan struct{}, 10)

	for _, review := range data {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(review model.Review) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			// 使用 types.SqlParameter 而不是 redshiftdata.SqlParameter
			params := []types.SqlParameter{
				{
					Name:  aws.String("id"),
					Value: aws.String(review.Id),
				},
				{
					Name:  aws.String("product_name"),
					Value: aws.String(review.ProductType),
				},
				{
					Name:  aws.String("review_date"),
					Value: aws.String(time.Unix(review.ReviewDate, 0).Format(time.RFC3339)), // 转换为 ISO 8601
				},
				{
					Name:  aws.String("review_content"),
					Value: aws.String(review.ReviewContent),
				},
			}

			sqlStatement := `
                INSERT INTO review_results 
                (id, product_name, review_date, review_content)
                VALUES 
                (:id, :product_name, :review_date, :review_content)
            `

			input := &redshiftdata.ExecuteStatementInput{
				WorkgroupName: aws.String(workgroupName),
				Database:      aws.String(databaseName),
				Sql:           aws.String(sqlStatement),
				Parameters:    params,
			}

			_, err := RedShiftClient.ExecuteStatement(ctx, input)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("error inserting data for ID %s: %v", review.Id, err))
				fmt.Printf("err【%+v】", err)
				mu.Unlock()
				return
			}

			mu.Lock()
			count++
			fmt.Printf("Inserted review result for ID: %s\n", review.Id)
			mu.Unlock()
		}(review)
	}

	wg.Wait()
	close(semaphore)

	if len(errs) > 0 {
		return Result{Count: count, Error: fmt.Errorf("multiple errors occurred: %v", errs)}
	}

	return Result{Count: count}
}

// InitRedShift 初始化 Redshift 客户端
func InitRedShift() {
	ctx := context.Background()

	// 初始化 AWS 配置和客户端
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     utils.ACCESS_KEY,
				SecretAccessKey: utils.SECRET_KEY,
			},
		}))
	if err != nil {
		fmt.Printf("Unable to load SDK config: %v\n", err)
		return
	}

	RedShiftClient = redshiftdata.NewFromConfig(cfg)
}

// InsertDataToRedshift 将数据插入到 Redshift
func InsertDataToRedshift(ctx context.Context, data []model.Review) {
	//resultChan := InsertReviewResultAsync(ctx,
	//	"voc",
	//	"dev",
	//	data)
	//
	//// 等待结果
	//result := <-resultChan
	//if result.Error != nil {
	//	fmt.Printf("Error: %v\n", result.Error)
	//} else {
	//	fmt.Printf("%d records were added to the review_results table.\n", result.Count)
	//}

	err := BatchInsertWithValues(ctx, "voc", "dev", data)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("%d records were added to the review table.\n", err)
	}

	fmt.Println("插入redshift数据完成")
}

// 辅助函数：等待查询完成
func waitForQueryCompletion(ctx context.Context, queryID *string) error {
	for {
		status, err := RedShiftClient.DescribeStatement(ctx, &redshiftdata.DescribeStatementInput{
			Id: queryID,
		})
		if err != nil {
			return err
		}
		fmt.Printf("status:【%+v】\n", status.Status)
		switch status.Status {
		case types.StatusStringFinished:
			// 查询完成，继续处理结果
			fmt.Println("Query completed successfully")
		case types.StatusStringFailed:
			fmt.Printf("failed:【%s】\n", *status.Error)
			return fmt.Errorf("query failed: %v", status.Error)
		case types.StatusStringAborted:
			fmt.Printf("StatusStringAborted:【%s】\n", *status.Error)
			return fmt.Errorf("query was aborted")
		default:
			// 查询仍在进行中，等待后继续检查
			time.Sleep(time.Second)
		}

		time.Sleep(time.Second)
	}
}
