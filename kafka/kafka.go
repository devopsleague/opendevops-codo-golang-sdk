package kafka

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/dnwe/otelsarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// KafkaConfigOption 用于配置 Kafka 配置项的函数类型
type KafkaConfigOption func(c *KafkaConfig)

// KafkaConfig 包含 Kafka 连接配置
type KafkaConfig struct {
	// Kafka Broker 地址列表
	BootstrapServers []string `json:"bootstrap_servers" yaml:"bootstrap_servers" env:"DEFAULT_KAFKA_BOOTSTRAP_SERVERS"`

	// Kafka Consumer Group Id
	GroupId string `json:"group_id" yaml:"group_id" env:"DEFAULT_KAFKA_GROUP_ID"`

	// SASL 用户名（如果需要使用 SASL 认证）
	SASLUsername string `json:"sasl_username" yaml:"sasl_username" env:"DEFAULT_KAFKA_SASL_USERNAME"`

	// SASL 密码（如果需要使用 SASL 认证）
	SASLPassword string `json:"sasl_password" yaml:"sasl_password" env:"DEFAULT_KAFKA_SASL_PASSWORD"`

	// SASL 认证机制（如 PLAIN）
	SASLMechanism string `json:"sasl_mechanism" yaml:"sasl_mechanism" env:"DEFAULT_KAFKA_SASL_MECHANISM"`

	// Kafka 最大请求空闲时间（单位：秒）
	MaxOpenRequests uint32 `json:"max_open_requests" yaml:"max_open_requests" env:"DEFAULT_KAFKA_MAX_OPEN_REQUESTS"`

	// DialTimeout   最大连接超时时间（单位：秒）
	DialTimeout uint32 `json:"dial_read_timeout" yaml:"dial_read_timeout" env:"DEFAULT_KAFKA_DIAL_READ_TIMEOUT"`

	// ReadTmout 最大读取超时时间（单位：秒）
	ReadTimeout uint32 `json:"read_timeout" yaml:"read_timeout" env:"DEFAULT_KAFKA_READ_TIMEOUT"`

	// WriteTimeout 最大写入超时时间（单位：秒）
	WriteTimeout uint32 `json:"write_timeout" yaml:"write_timeout" env:"DEFAULT_KAFKA_WRITE_TIMEOUT"`

	// OpenTelemetry 跟踪提供者
	TracerProvider trace.TracerProvider `json:"-"`

	// OpenTelemetry 度量提供者
	MeterProvider metric.MeterProvider `json:"-"`

	// OpenTelemetry 传播器
	Propagator propagation.TextMapPropagator `json:"-"`
}

func defaultConfig() KafkaConfig {
	emptyOr := func(str string, defaultVal string) string {
		str = strings.TrimSpace(str)
		if str == "" {
			return defaultVal
		}
		return str
	}
	bootstrapServers := emptyOr(os.Getenv("DEFAULT_KAFKA_BOOTSTRAP_SERVERS"), "localhost:9092")
	groupId := emptyOr(os.Getenv("DEFAULT_KAFKA_GROUP_ID"), "default-group")
	saslUserName := emptyOr(os.Getenv("DEFAULT_KAFKA_SASL_USERNAME"), "")
	saslPassword := emptyOr(os.Getenv("DEFAULT_KAFKA_SASL_PASSWORD"), "")
	saslMechanism := emptyOr(os.Getenv("DEFAULT_KAFKA_SASL_MECHANISM"), "")
	dialTimeout := emptyOr(os.Getenv("DEFAULT_KAFKA_DIAL_READ_TIMEOUT"), "10")
	readTimeout := emptyOr(os.Getenv("DEFAULT_KAFKA_READ_TIMEOUT"), "10")
	writeTimeout := emptyOr(os.Getenv("DEFAULT_KAFKA_WRITE_TIMEOUT"), "10")
	maxOpenRequests := emptyOr(os.Getenv("DEFAULT_KAFKA_MAX_OPEN_REQUESTS"), "10")
	bootstrapServersList := strings.Split(bootstrapServers, ",")
	i64dialTimeout, _ := strconv.Atoi(dialTimeout)
	i64reeadTimeout, _ := strconv.Atoi(readTimeout)
	i64writeTimeout, _ := strconv.Atoi(writeTimeout)
	i64maxOpenRequests, _ := strconv.Atoi(maxOpenRequests)

	return KafkaConfig{
		BootstrapServers: bootstrapServersList,
		GroupId:          groupId,
		SASLUsername:     saslUserName,
		SASLPassword:     saslPassword,
		SASLMechanism:    saslMechanism,
		DialTimeout:      uint32(i64dialTimeout),
		ReadTimeout:      uint32(i64reeadTimeout),
		WriteTimeout:     uint32(i64writeTimeout),
		MaxOpenRequests:  uint32(i64maxOpenRequests),
		TracerProvider:   otel.GetTracerProvider(),
		MeterProvider:    otel.GetMeterProvider(),
		Propagator:       otel.GetTextMapPropagator(),
	}
}

// WithBrokerAddrs 配置 Kafka BootstrapServer 地址列表
func WithBootstrapServers(bootstrapServers string) KafkaConfigOption {
	bootstrapServersList := strings.Split(bootstrapServers, ",")
	return func(c *KafkaConfig) {
		c.BootstrapServers = bootstrapServersList
	}
}

// WithGroupID 配置 Kafka Consumer Group ID
func WithGroupID(groupID string) KafkaConfigOption {
	return func(c *KafkaConfig) {
		c.GroupId = groupID
	}
}

// WithSASLUsername 配置 SASL 用户名
func WithSASLUsername(username string) KafkaConfigOption {
	return func(c *KafkaConfig) {
		c.SASLUsername = username
	}
}

// WithSASLPassword 配置 SASL 密码
func WithSASLPassword(password string) KafkaConfigOption {
	return func(c *KafkaConfig) {
		c.SASLPassword = password
	}
}

// WithSASLMechanism 配置 SASL 认证机制
func WithSASLMechanism(mechanism string) KafkaConfigOption {
	return func(c *KafkaConfig) {
		c.SASLMechanism = mechanism
	}
}

// WithConnMaxLifetime 配置 Kafka 最大写入超时时间
func WithWriteTimeout(t uint32) KafkaConfigOption {
	return func(c *KafkaConfig) {
		c.WriteTimeout = t
	}
}

// WithMaxIdleConns 配置 Kafka 最大读取超时时间
func WithReadTimeout(t uint32) KafkaConfigOption {
	return func(c *KafkaConfig) {
		c.ReadTimeout = t
	}
}

// WithMaxOpenConns 配置 Kafka 最大连接超时时间
func WithDialTimeout(t uint32) KafkaConfigOption {
	return func(c *KafkaConfig) {
		c.DialTimeout = t
	}
}

func NewProducer(opts ...KafkaConfigOption) (sarama.AsyncProducer, func(), error) {
	kafkaConf := defaultConfig()
	for _, opt := range opts {
		opt(&kafkaConf)
	}
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0
	// So we can know the partition and offset of messages.
	config.Producer.Return.Successes = true

	// 配置 SASL 认证
	if kafkaConf.SASLUsername != "" && kafkaConf.SASLPassword != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = kafkaConf.SASLUsername
		config.Net.SASL.Password = kafkaConf.SASLPassword
		config.Net.SASL.Mechanism = sarama.SASLMechanism(kafkaConf.SASLMechanism)
	}
	config.Net.MaxOpenRequests = int(kafkaConf.MaxOpenRequests)
	config.Net.DialTimeout = time.Second * time.Duration(kafkaConf.DialTimeout)
	config.Net.ReadTimeout = time.Second * time.Duration(kafkaConf.ReadTimeout)
	config.Net.WriteTimeout = time.Second * time.Duration(kafkaConf.WriteTimeout)

	producer, err := sarama.NewAsyncProducer(kafkaConf.BootstrapServers, config)
	if err != nil {
		return nil, nil, fmt.Errorf("starting Sarama producer: %w", err)
	}

	// Wrap instrumentation
	producer = otelsarama.WrapAsyncProducer(config, producer,
		otelsarama.WithTracerProvider(kafkaConf.TracerProvider),
		otelsarama.WithPropagators(kafkaConf.Propagator),
	)
	ctx, cancel := context.WithCancel(context.Background())

	// We will log to STDOUT if we're not able to produce messages.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic: %v\n, Stack trace: %s", r, string(debug.Stack()))
			}
		}()
		for {
			select {
			case err := <-producer.Errors():
				fmt.Printf("Failed to write message: %v\n, Stack trace: %s", err, string(debug.Stack()))

			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		for {
			select {
			case <-producer.Successes():
			case <-ctx.Done():
				return
			}
		}
	}()

	return producer, func() {
		cancel()
		producer.Close()
	}, nil
}

func NewConsumerGroup(opts ...KafkaConfigOption) (sarama.ConsumerGroup, func(), error) {
	kafkaConf := defaultConfig()
	for _, opt := range opts {
		opt(&kafkaConf)
	}
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// 配置 SASL 认证
	if kafkaConf.SASLUsername != "" && kafkaConf.SASLPassword != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = kafkaConf.SASLUsername
		config.Net.SASL.Password = kafkaConf.SASLPassword
		config.Net.SASL.Mechanism = sarama.SASLMechanism(kafkaConf.SASLMechanism)
	}
	config.Net.MaxOpenRequests = int(kafkaConf.MaxOpenRequests)
	config.Net.DialTimeout = time.Second * time.Duration(kafkaConf.DialTimeout)
	config.Net.ReadTimeout = time.Second * time.Duration(kafkaConf.ReadTimeout)
	config.Net.WriteTimeout = time.Second * time.Duration(kafkaConf.WriteTimeout)

	consumerGroup, err := sarama.NewConsumerGroup(kafkaConf.BootstrapServers, kafkaConf.GroupId, config)
	if err != nil {
		return nil, nil, fmt.Errorf("starting consumer group: %w\n Stack trace: %s", err, string(debug.Stack()))
	}

	return consumerGroup, func() {
		consumerGroup.Close()
	}, nil
}
