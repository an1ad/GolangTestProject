package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/resty.v1"
)

func main() {
	// Создание контекста и функции отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Создание приложения с использованием фреймворка fx
	app := fx.New(
		fx.Provide(rootCommandProvider),
		fx.Provide(NewConfig),
		fx.Provide(NewLogger),
		fx.Provide(NewRestyClient),
		fx.Invoke(Run),
	)
	// Запуск приложения
	if err := app.Start(ctx); err != nil {
		fmt.Println("Error starting the application:", err)
		os.Exit(1)
	}
	// Отложенный вызов остановки приложения при завершении программы
	defer func() {
		if err := app.Stop(ctx); err != nil {
			fmt.Println("Error stopping the application:", err)
		}
	}()

	os.Exit(0)
}

// rootCommandProvider предоставляет *cobra.Command в качестве зависимости для NewConfig.
func rootCommandProvider() *cobra.Command {
	// Создание корневой команды с параметрами
	rootCmd := &cobra.Command{
		Use:   "testApp",
		Short: "App",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Root command executed")
		},
	}
	// Добавление флагов к корневой команде
	rootCmd.Flags().String("url", "http://localhost:8080", "URL for requests")
	rootCmd.Flags().Int("amount", 1000, "Number of requests")
	rootCmd.Flags().Int("per_second", 10, "Requests per second")

	return rootCmd
}

// Config представляет структуру конфигурации приложения
type Config struct {
	URL      string `mapstructure:"url"`
	Requests struct {
		Amount    int `mapstructure:"amount"`
		PerSecond int `mapstructure:"per_second"`
	} `mapstructure:"requests"`
}

// NewConfig создает объект конфигурации из командной строки
func NewConfig(cmd *cobra.Command) (*Config, error) {
	// Получение значений флагов из командной строки
	url, _ := cmd.Flags().GetString("url")
	amount, _ := cmd.Flags().GetInt("amount")
	perSecond, _ := cmd.Flags().GetInt("per_second")
	// Создание объекта конфигурации
	config := &Config{
		URL: url,
		Requests: struct {
			Amount    int `mapstructure:"amount"`
			PerSecond int `mapstructure:"per_second"`
		}{
			Amount:    amount,
			PerSecond: perSecond,
		},
	}
	return config, nil
}

// NewLogger создает новый логгер
func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}
	return logger, nil
}

// NewRestyClient создает новый клиент Resty
func NewRestyClient() (*resty.Client, error) {
	return resty.New(), nil
}

// RunWithContext выполняет отправку запросов с использованием контекста
func RunWithContext(ctx context.Context, config *Config, logger *zap.Logger, client *resty.Client, done chan struct{}) {
	// Отложенный вызов close(done) после открытия канала
	defer close(done)

	for i := 1; i <= config.Requests.Amount; i++ {
		select {
		case <-ctx.Done():
			// Контекст завершен, прерываем выполнение
			return
		default:
			go func(iteration int) {
				url := config.URL
				body := map[string]interface{}{"iteration": iteration}

				resp, err := client.R().
					SetBody(body).
					Post(url)

				if err != nil {
					logger.Error("Error sending request", zap.Error(err), zap.Int("iteration", iteration))
					return
				}

				logger.Info(fmt.Sprintf("Request sent. Status: %d", resp.StatusCode()), zap.Int("iteration", iteration))
			}(i)

			if i%config.Requests.PerSecond == 0 {
				time.Sleep(time.Second)
			}
		}
	}
	// Даем некоторое время для завершения всех горутин
	time.Sleep(5 * time.Second)
}

// Run выполняет отправку запросов
func Run(config *Config, logger *zap.Logger, client *resty.Client) {
	for i := 1; i <= config.Requests.Amount; i++ {
		go func(iteration int) {
			url := config.URL
			body := map[string]interface{}{"iteration": iteration}

			resp, err := client.R().
				SetBody(body).
				Post(url)

			if err != nil {
				logger.Error("Error sending request", zap.Error(err), zap.Int("iteration", iteration))
				return
			}

			logger.Info(fmt.Sprintf("Request sent. Status: %d", resp.StatusCode()), zap.Int("iteration", iteration))
		}(i)

		if i%config.Requests.PerSecond == 0 {
			time.Sleep(time.Second)
		}
	}
	// Даем некоторое время для завершения всех горутин
	time.Sleep(5 * time.Second)
}
