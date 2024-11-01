package consts

import "strconv"

type ErrorCode int

const (
	NotFound                  ErrorCode = 404
	BadRequest                ErrorCode = 400
	Unauthorized              ErrorCode = 401
	Forbidden                 ErrorCode = 403
	NotAllowed                ErrorCode = 405
	NotAcceptable             ErrorCode = 406
	Conflict                  ErrorCode = 409
	Gone                      ErrorCode = 410
	PreconditionFailed        ErrorCode = 412
	RequestEntityTooLarge     ErrorCode = 413
	UnsupportMediaType        ErrorCode = 415
	InternalServerError       ErrorCode = 500
	ServiceUnavailable        ErrorCode = 503
	ServiceNotImplemented     ErrorCode = 501
	HandlerUncatchedException ErrorCode = 504
	ConfigImportError         ErrorCode = 1001
	ConfigItemNotfoundError   ErrorCode = 1002
)

func (e ErrorCode) String() string {
	return strconv.Itoa(int(e))
}

const (
	DB_CONFIG_ITEM  = "databases"
	DBHOST_KEY      = "host"
	DBPWD_KEY       = "pwd"
	DBUSER_KEY      = "user"
	DBNAME_KEY      = "name"
	DBPORT_KEY      = "port"
	SF_DB_KEY       = "vmobel"
	DEFAULT_DB_KEY  = "default"
	READONLY_DB_KEY = "readonly"
)

const (
	REDIS_CONFIG_ITEM   = "redises"
	RD_HOST_KEY         = "host"
	RD_PORT_KEY         = "port"
	RD_DB_KEY           = "db"
	RD_AUTH_KEY         = "auth"
	RD_CHARSET_KEY      = "charset"
	RD_DECODE_RESPONSES = "decode_responses"
	RD_PASSWORD_KEY     = "password"
	DEFAULT_RD_KEY      = "default"
)

// ETCD
const (
	DEFAULT_ETCD_KEY      = "default"
	BACKUP_ETCD_KEY       = "backup"
	DEFAULT_ETCD_HOST     = "host"
	DEFAULT_ETCD_PORT     = "port"
	DEFAULT_ETCD_PROTOCOL = "protocol"
	DEFAULT_ETCD_USER     = "user"
	DEFAULT_ETCD_PWD      = "pwd"
)

// MQ
const (
	MQ_CONFIG_ITEM = "mqs"
	MQ_ADDR        = "MQ_ADDR"
	MQ_PORT        = "MQ_PORT"
	MQ_VHOST       = "MQ_VHOST"
	MQ_USER        = "MQ_USER"
	MQ_PWD         = "MQ_PWD"
	DEFAULT_MQ_KEY = "default"
	AGENT_MQ_KEY   = "agent"
)

// JMS
const (
	JMS_CONFIG_ITEM    = "jmss"
	DEFAULT_JMS_KEY    = "default"
	JMS_API_BASE_URL   = "jms_url"
	JMS_API_KEY_ID     = "jms_key_id"
	JMS_API_KEY_SECRET = "jms_key_secret"
)

// consul
const (
	CONSUL_CONFIG_ITEM = "consuls"
	DEFAULT_CS_KEY     = "default"
	CONSUL_HOST_KEY    = "cs_host"
	CONSUL_PORT_KEY    = "cs_port"
	CONSUL_TOKEN_KEY   = "cs_token"
	CONSUL_SCHEME_KEY  = "cs_scheme"
)

const (
	APP_NAME          = "app_name"
	LOG_PATH          = "log_path"
	LOG_BACKUP_COUNT  = "log_backup_count"
	LOG_MAX_FILE_SIZE = "log_max_filesize"
)

const (
	REQUEST_START_SIGNAL    = "request_start"
	REQUEST_FINISHED_SIGNAL = "request_finished"
)

const (
	NW_SALT      = "nw"
	ALY_SALT     = "aly"
	TX_SALT      = "tx"
	SG_SALT      = "sg"
	DEFAULT_SALT = "default"
	SALT_API     = "salt_api"
	SALT_USER    = "salt_username"
	SALT_PW      = "salt_password"
	SALT_OUT     = "salt_timeout"
)

const (
	NW_INCEPTION      = "nw"
	ALY_INCEPTION     = "aly"
	TX_INCEPTION      = "tx"
	DEFAULT_INCEPTION = "default"
)

const (
	REGION       = "cn-hangzhou"
	PRODUCT_NAME = "Dysmsapi"
	DOMAIN       = "dysmsapi.aliyuncs.com"
)

// crypto
const (
	AES_CRYPTO_KEY = "aes_crypto_key"
	// app settings
	APP_SETTINGS = "APP_SETTINGS"
	// all user info
	USERS_INFO = "USERS_INFO"
)

// API GW
const (
	WEBSITE_API_GW_URL = "api_gw"
	API_AUTH_KEY       = "settings_auth_key"
	EMAILLOGIN_DOMAIN  = "EMAILLOGIN_DOMAIN"
	EMAILLOGIN_SERVER  = "EMAILLOGIN_SERVER"
)

// email
const (
	EMAIL_SUBJECT_PREFIX = "EMAIL_SUBJECT_PREFIX"
	EMAIL_HOST           = "EMAIL_HOST"
	EMAIL_PORT           = "EMAIL_PORT"
	EMAIL_HOST_USER      = "EMAIL_HOST_USER"
	EMAIL_HOST_PASSWORD  = "EMAIL_HOST_PASSWORD"
	EMAIL_USE_SSL        = "EMAIL_USE_SSL"
	EMAIL_USE_TLS        = "EMAIL_USE_TLS"
)

// 短信配置
const (
	SMS_REGION            = "SMS_REGION"
	SMS_PRODUCT_NAME      = "SMS_PRODUCT_NAME"
	SMS_DOMAIN            = "SMS_DOMAIN"
	SMS_ACCESS_KEY_ID     = "SMS_ACCESS_KEY_ID"
	SMS_ACCESS_KEY_SECRET = "SMS_ACCESS_KEY_SECRET"
)

// 钉钉
const (
	DING_TALK_WEBHOOK = "DING_TALK_WEBHOOK"
)

// 存储
const (
	STORAGE_REGION     = "STORAGE_REGION"
	STORAGE_NAME       = "STORAGE_NAME"
	STORAGE_PATH       = "STORAGE_PATH"
	STORAGE_KEY_ID     = "STORAGE_KEY_ID"
	STORAGE_KEY_SECRET = "STORAGE_KEY_SECRET"
)

// LDAP
const (
	LDAP_SERVER_HOST    = "LDAP_SERVER_HOST"
	LDAP_SERVER_PORT    = "LDAP_SERVER_PORT"
	LDAP_ADMIN_DN       = "LDAP_ADMIN_DN"
	LDAP_ADMIN_PASSWORD = "LDAP_ADMIN_PASSWORD"
	LDAP_SEARCH_BASE    = "LDAP_SEARCH_BASE"
	LDAP_SEARCH_FILTER  = "LDAP_SEARCH_FILTER"
	LDAP_ATTRIBUTES     = "LDAP_ATTRIBUTES"
	LDAP_USE_SSL        = "LDAP_USE_SSL"
	LDAP_ENABLE         = "LDAP_ENABLE"
)

// token 超时时间
const (
	TOKEN_EXP_TIME = "TOKEN_EXP_TIME"
)

// 全局 二次认证
const (
	MFA_GLOBAL = "MFA_GLOBAL"
)

// task event 状态
const (
	STATE_NEW     = "0"  // 新建任务
	STATE_WAIT    = "1"  // 等待执行
	STATE_RUNNING = "2"  // 正在运行
	STATE_SUCCESS = "3"  // 成功完成
	STATE_ERROR   = "4"  // 发生错误
	STATE_MANUAL  = "5"  // 等待手动
	STATE_BREAK   = "6"  // 中止状态 //不区分手动和自动
	STATE_TIMING  = "7"  // 定时状态
	STATE_UNKNOWN = "8"  // 未知状态  // debug
	STATE_FAIL    = "9"  // 失败         // debug
	STATE_IGNORE  = "10" // 忽略执行
	STATE_QUEUE   = "11" // 排队中   //订单和任务节点公用
)

// 订单 状态
const (
	ORDER_STATE_WAITING          = "31" // 订单等待中
	ORDER_STATE_RUNNING          = "32" // 订单执行中
	ORDER_STATE_SUCCESS          = "33" // 订单成功
	ORDER_STATE_FAIL             = "34" // 订单失败
	ORDER_STATE_WAITING_APPROVAL = "35" // 订单等待审批
	ORDER_STATE_TERMINATED       = "39" // 订单终止
	ORDER_STATE_QUEUE            = "11" // 订单排队中
	EXEC_TIMEOUT                 = 1800
)

// 节点地址
const (
	NODE_ADDRESS      = "NODE_ADDRESS"
	EXEC_NODE_MAP_KEY = "EXEC_NODE_MAP_KEY"
	AGENT_USED_KEY    = "agent_is_used_map_mark_key"
)
