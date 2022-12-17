package main

import (
	"context"
	"errors"
	"fmt"
	razorpay "github.com/mehuled/razorpay-go"
	"io"
	"log"
	"os"
)

const (
	EnvRazorpayAPIKeyId     = "RAZORPAY_API_KEY_ID"
	EnvRazorpayAPIKeySecret = "RAZORPAY_API_KEY_SECRET"
	ContextLogger           = "ctx.logger"
	ContextRazorpayClient   = "ctx.razorpay.client"
)

func initContext(ctx context.Context) context.Context {
	logger := log.New(os.Stdout, "", 0)
	client := razorpay.NewAPIClient(razorpay.NewConfiguration())

	ctx = context.WithValue(ctx, ContextLogger, logger)

	ctx = context.WithValue(ctx, razorpay.ContextBasicAuth, razorpay.BasicAuth{
		UserName: os.Getenv(EnvRazorpayAPIKeyId),
		Password: os.Getenv(EnvRazorpayAPIKeySecret),
	})
	ctx = context.WithValue(ctx, ContextRazorpayClient, client)
	return ctx
}

func main() {
	ctx := context.Background()
	ctx = initContext(ctx)

	logger := ctx.Value(ContextLogger).(*log.Logger)

	order, err := CreateOrder(ctx, 1000, "INR", "receipt #121")

	if err != nil {
		panic(err)
	}

	logger.Println("// Orders")
	logger.Println(fmt.Sprintf("order_id: %s", order.Id))

	payments, err := FetchPayments(ctx)

	logger.Println("// Payments")
	for cnt, payment := range payments {
		logger.Println(fmt.Sprintf("%d. payment_id : %s \t amount : %d", cnt+1, payment.Id, payment.Amount))
	}

}

func CreateOrder(ctx context.Context, amount int64, currency string, receipt string) (*razorpay.Order, error) {
	logger := ctx.Value(ContextLogger).(*log.Logger)
	client := ctx.Value(ContextRazorpayClient).(*razorpay.APIClient)

	order, response, err := client.OrdersApi.OrdersPost(ctx, razorpay.OrdersBody{
		Amount:   amount,
		Currency: currency,
		Receipt:  receipt,
	})

	if err != nil {
		logger.Fatal(err)
		return nil, err
	}

	if response.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(response.Body)
		return nil, errors.New(string(responseBody))
	}

	return &order, nil
}

func FetchPayments(ctx context.Context) ([]razorpay.Payment, error) {
	logger := ctx.Value(ContextLogger).(*log.Logger)
	client := ctx.Value(ContextRazorpayClient).(*razorpay.APIClient)

	payments, response, err := client.PaymentsApi.PaymentsGet(ctx)

	if err != nil {
		logger.Fatal(err)
		return nil, err
	}

	if response.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(response.Body)
		return nil, errors.New(string(responseBody))
	}

	return payments.Items, nil

}
