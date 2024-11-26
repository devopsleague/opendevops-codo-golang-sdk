package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

type ConfigOptions struct {
	loadFromEnv bool
	envPrefix   string

	loadFromYaml bool
	yamlPath     string

	loadFromFlag bool
	flagSet      *flag.FlagSet
	flagArgs     []string

	loadFromPFlag bool
	pflagSet      *pflag.FlagSet
	pflagArgs     []string
}

type ConfigOption func(*ConfigOptions)

func defaultConfigOptions() ConfigOptions {
	return ConfigOptions{
		loadFromEnv:  true,
		envPrefix:    "CODO",
		loadFromYaml: false,
		yamlPath:     "",
	}
}

// WithEnv 从环境变量加载配置
func WithEnv(prefix string) ConfigOption {
	return func(options *ConfigOptions) {
		options.loadFromEnv = true
		options.envPrefix = prefix
	}
}

// WithYaml 从 YAML 加载配置, 依赖 json flag , 需要定义正确的数据类型
func WithYaml(filepath string) ConfigOption {
	return func(options *ConfigOptions) {
		options.loadFromYaml = true
		options.yamlPath = filepath
	}
}

// WithFlag 从 flag 加载配置
func WithFlag(flagSet *flag.FlagSet, args []string) ConfigOption {
	return func(options *ConfigOptions) {
		options.loadFromFlag = true
		options.flagSet = flagSet
		options.flagArgs = args
	}
}

// WithPFlag 从 pflag 加载配置
func WithPFlag(pflagSet *pflag.FlagSet, args []string) ConfigOption {
	return func(options *ConfigOptions) {
		options.loadFromPFlag = true
		options.pflagSet = pflagSet
		options.pflagArgs = args
	}
}

// LoadConfig 加载配置, 优先级为 flag > env > yaml
func LoadConfig(dst interface{}, opts ...ConfigOption) error {
	c := defaultConfigOptions()
	for _, opt := range opts {
		opt(&c)
	}
	if c.loadFromYaml {
		err := LoadYaml(c.yamlPath, dst)
		if err != nil {
			return err
		}
	}
	if c.loadFromEnv {
		err := LoadEnv(c.envPrefix, dst)
		if err != nil {
			return err
		}
	}
	if c.loadFromFlag {
		err := LoadFlag(c.flagSet, c.flagArgs, dst)
		if err != nil {
			return err
		}
	}
	if c.loadFromPFlag {
		err := LoadPFlag(c.pflagSet, c.pflagArgs, dst)
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadYaml(filepath string, conf interface{}) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	var dst map[string]interface{}
	err = yaml.NewDecoder(f).Decode(&dst)
	if err != nil {
		return err
	}
	bs, _ := Marshal(dst)
	return Unmarshal(bs, conf)
}

// LoadEnv 解析环境变量到结构体
func LoadEnv(prefix string, v interface{}) error {
	_, err := parseEnv(prefix, reflect.ValueOf(v))
	return err
}

func parseEnv(prefix string, v reflect.Value) (bool, error) {
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return false, fmt.Errorf("invalid value: must be a non-nil pointer")
	}

	v = v.Elem()
	t := v.Type()

	var isSetFields bool
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		keys := []string{
			strings.ToUpper(field.Tag.Get("yaml")),
			strings.ToUpper(field.Tag.Get("json")),
			strings.ToUpper(field.Name),
		}

		var envKeyPrefix string
		for _, key := range keys {
			if key != "" {
				envKeyPrefix = key
				if prefix != "" {
					envKeyPrefix = prefix + "_" + envKeyPrefix
				}
				break
			}
		}

		// 如果 ENV 是 空, 则使用 envKeyPrefix 作为 KEY
		envKey := field.Tag.Get("env")
		if envKey == "" {
			envKey = envKeyPrefix
		}

		var valueElement = value
		if value.Kind() == reflect.Pointer && value.CanSet() {
			if value.IsNil() {
				valueElement = reflect.New(value.Type().Elem()).Elem()
			} else {
				valueElement = value.Elem()
			}
		}

		switch valueElement.Kind() {
		case reflect.Struct:
			ok, err := parseEnv(envKeyPrefix, valueElement.Addr())
			if err != nil {
				return false, err
			}
			if ok {
				if value.Kind() == reflect.Pointer && value.CanSet() {
					value.Set(valueElement.Addr())
				} else {
					value.Set(valueElement)
				}
				isSetFields = true
			}
		case reflect.Slice:
			if err := setSliceField(valueElement, envKey); err != nil {
				return false, err
			}
		default:
			ok, err := setField(valueElement, envKey)
			if err != nil {
				return false, err
			}
			if ok {
				isSetFields = true
			}
		}
	}

	return isSetFields, nil
}

func LoadFlag(flagSet *flag.FlagSet, args []string, conf interface{}) error {
	err := parseFlagStruct(flagSet, reflect.ValueOf(conf))
	if err != nil {
		return err
	}
	return flagSet.Parse(args)
}

func parseFlagStruct(flagSet *flag.FlagSet, v reflect.Value) error {
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("invalid value: must be a non-nil pointer, current=%s", v.Kind())
	}
	v = v.Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		valueElement := fieldValue
		if field.Type.Kind() == reflect.Struct {
			err := parseFlagStruct(flagSet, valueElement.Addr())
			if err != nil {
				return err
			}
			continue
		}

		flagName := field.Tag.Get("flag")
		if flagName == "" {
			continue
		}

		var flagShot string
		flags := strings.Split(flagName, "|")
		if len(flags) > 1 {
			for _, s := range flags {
				if s == "" {
					continue
				}
				if len(s) == 1 {
					flagShot = s
				} else {
					flagName = s
				}
			}
		}
		_ = flagShot

		usage := field.Tag.Get("usage")

		switch fieldValue.Kind() {
		case reflect.String:
			ptr := fieldValue.Addr().Interface().(*string)
			var defaultValue string
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.StringVar(ptr, flagName, defaultValue, usage)
		case reflect.Int:
			ptr := fieldValue.Addr().Interface().(*int)
			var defaultValue int
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.IntVar(ptr, flagName, defaultValue, usage)
		case reflect.Float64:
			ptr := fieldValue.Addr().Interface().(*float64)
			var defaultValue float64
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.Float64Var(ptr, flagName, defaultValue, usage)
		case reflect.Int64:
			ptr := fieldValue.Addr().Interface().(*int64)
			var defaultValue int64
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.Int64Var(ptr, flagName, defaultValue, usage)
		case reflect.Uint:
			ptr := fieldValue.Addr().Interface().(*uint)
			var defaultValue uint
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.UintVar(ptr, flagName, defaultValue, usage)
		case reflect.Uint64:
			ptr := fieldValue.Addr().Interface().(*uint64)
			var defaultValue uint64
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.Uint64Var(ptr, flagName, defaultValue, usage)
		case reflect.Bool:
			ptr := fieldValue.Addr().Interface().(*bool)
			var defaultValue bool
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.BoolVar(ptr, flagName, defaultValue, usage)
		case reflect.Slice:
			switch fieldValue.Type().Elem().Kind() {
			case reflect.String:
				ptr := fieldValue.Addr().Interface().(*[]string)
				flagSet.Var((*sliceValue[string])(ptr), flagName, usage)
			case reflect.Int:
				ptr := fieldValue.Addr().Interface().(*[]int)
				flagSet.Var((*sliceValue[int])(ptr), flagName, usage)
			case reflect.Float64:
				ptr := fieldValue.Addr().Interface().(*[]float64)
				flagSet.Var((*sliceValue[float64])(ptr), flagName, usage)
			case reflect.Int64:
				ptr := fieldValue.Addr().Interface().(*[]int64)
				flagSet.Var((*sliceValue[int64])(ptr), flagName, usage)
			case reflect.Uint:
				ptr := fieldValue.Addr().Interface().(*[]uint)
				flagSet.Var((*sliceValue[uint])(ptr), flagName, usage)
			case reflect.Uint64:
				ptr := fieldValue.Addr().Interface().(*[]uint64)
				flagSet.Var((*sliceValue[uint64])(ptr), flagName, usage)
			case reflect.Bool:
				ptr := fieldValue.Addr().Interface().(*[]bool)
				flagSet.Var((*sliceValue[bool])(ptr), flagName, usage)
			default:
				return fmt.Errorf("unsupported slice type === %s", fieldValue.Type().Elem().Kind())
			}
		default:
			return fmt.Errorf("unsupported type === %s", fieldValue.Kind())
		}
	}

	return nil
}

func LoadPFlag(flagSet *pflag.FlagSet, args []string, conf interface{}) error {
	err := parsePFlagStruct(flagSet, reflect.ValueOf(conf))
	if err != nil {
		return err
	}
	return flagSet.Parse(args)
}

func parsePFlagStruct(flagSet *pflag.FlagSet, v reflect.Value) error {
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("invalid value: must be a non-nil pointer, current=%s", v.Kind())
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		valueElement := fieldValue
		if field.Type.Kind() == reflect.Struct {
			err := parsePFlagStruct(flagSet, valueElement.Addr())
			if err != nil {
				return err
			}
			continue
		}

		flagName := field.Tag.Get("flag")
		if flagName == "" {
			continue
		}

		var flagShot string
		flags := strings.Split(flagName, "|")
		if len(flags) > 1 {
			for _, s := range flags {
				if s == "" {
					continue
				}
				if len(s) == 1 {
					flagShot = s
				} else {
					flagName = s
				}
			}
		}

		usage := field.Tag.Get("usage")

		switch fieldValue.Kind() {
		case reflect.String:
			ptr := fieldValue.Addr().Interface().(*string)
			var defaultValue string
			if ptr != nil {
				defaultValue = *ptr
			}

			flagSet.StringVarP(ptr, flagName, flagShot, defaultValue, usage)
		case reflect.Int:
			ptr := fieldValue.Addr().Interface().(*int)
			var defaultValue int
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.IntVarP(ptr, flagName, flagShot, defaultValue, usage)
		case reflect.Float64:
			ptr := fieldValue.Addr().Interface().(*float64)
			var defaultValue float64
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.Float64VarP(ptr, flagName, flagShot, defaultValue, usage)
		case reflect.Int64:
			ptr := fieldValue.Addr().Interface().(*int64)
			var defaultValue int64
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.Int64VarP(ptr, flagName, flagShot, defaultValue, usage)
		case reflect.Uint:
			ptr := fieldValue.Addr().Interface().(*uint)
			var defaultValue uint
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.UintVarP(ptr, flagName, flagShot, defaultValue, usage)
		case reflect.Uint64:
			ptr := fieldValue.Addr().Interface().(*uint64)
			var defaultValue uint64
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.Uint64VarP(ptr, flagName, flagShot, defaultValue, usage)
		case reflect.Bool:
			ptr := fieldValue.Addr().Interface().(*bool)
			var defaultValue bool
			if ptr != nil {
				defaultValue = *ptr
			}
			flagSet.BoolVarP(ptr, flagName, flagShot, defaultValue, usage)
		case reflect.Slice:
			switch fieldValue.Type().Elem().Kind() {
			case reflect.String:
				ptr := fieldValue.Addr().Interface().(*[]string)
				flagSet.VarP((*sliceValue[string])(ptr), flagName, flagShot, usage)
			case reflect.Int:
				ptr := fieldValue.Addr().Interface().(*[]int)
				flagSet.VarP((*sliceValue[int])(ptr), flagName, flagShot, usage)
			case reflect.Float64:
				ptr := fieldValue.Addr().Interface().(*[]float64)
				flagSet.VarP((*sliceValue[float64])(ptr), flagName, flagShot, usage)
			case reflect.Int64:
				ptr := fieldValue.Addr().Interface().(*[]int64)
				flagSet.VarP((*sliceValue[int64])(ptr), flagName, flagShot, usage)
			case reflect.Uint:
				ptr := fieldValue.Addr().Interface().(*[]uint)
				flagSet.VarP((*sliceValue[uint])(ptr), flagName, flagShot, usage)
			case reflect.Uint64:
				ptr := fieldValue.Addr().Interface().(*[]uint64)
				flagSet.VarP((*sliceValue[uint64])(ptr), flagName, flagShot, usage)
			case reflect.Bool:
				ptr := fieldValue.Addr().Interface().(*[]bool)
				flagSet.VarP((*sliceValue[bool])(ptr), flagName, flagShot, usage)
			default:
				return fmt.Errorf("unsupported slice type === %s", fieldValue.Type().Elem().Kind())
			}
		default:
			return fmt.Errorf("unsupported type === %s", fieldValue.Kind())
		}
	}

	return nil
}

type valueRange interface {
	~int | ~uint |
		~int64 | ~uint64 |
		~int32 | ~uint32 |
		~int16 | ~uint16 |
		~int8 | ~uint8 |
		~string |
		~float64 | ~float32 | bool
}

type sliceValue[T valueRange] []T

func (x *sliceValue[T]) String() string {
	return fmt.Sprintf("%v", *x)
}

func (x *sliceValue[T]) Type() string {
	var t T
	return fmt.Sprintf("slice_%T", t)
}

func (x *sliceValue[T]) Set(value string) error {
	strs := strings.Split(value, ",")
	var zero T
	switch interface{}(zero).(type) {
	case string:
		*x = *(*sliceValue[T])(unsafe.Pointer(&strs))
	case int:
		data := make([]int, 0, len(strs))
		for _, str := range strs {
			i64, _ := strconv.ParseInt(str, 10, 64)
			data = append(data, int(i64))
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case int64:
		data := make([]int64, 0, len(strs))
		for _, str := range strs {
			i64, _ := strconv.ParseInt(str, 10, 64)
			data = append(data, i64)
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case int32:
		data := make([]int32, 0, len(strs))
		for _, str := range strs {
			i64, _ := strconv.ParseInt(str, 10, 32)
			data = append(data, int32(i64))
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case int16:
		data := make([]int16, 0, len(strs))
		for _, str := range strs {
			i64, _ := strconv.ParseInt(str, 10, 16)
			data = append(data, int16(i64))
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case int8:
		data := make([]int8, 0, len(strs))
		for _, str := range strs {
			i64, _ := strconv.ParseInt(str, 10, 8)
			data = append(data, int8(i64))
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case uint:
		data := make([]uint, 0, len(strs))
		for _, str := range strs {
			u64, _ := strconv.ParseUint(str, 10, 64)
			data = append(data, uint(u64))
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case uint64:
		data := make([]uint64, 0, len(strs))
		for _, str := range strs {
			u64, _ := strconv.ParseUint(str, 10, 64)
			data = append(data, u64)
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case uint32:
		data := make([]uint32, 0, len(strs))
		for _, str := range strs {
			u64, _ := strconv.ParseUint(str, 10, 32)
			data = append(data, uint32(u64))
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case uint16:
		data := make([]uint16, 0, len(strs))
		for _, str := range strs {
			u64, _ := strconv.ParseUint(str, 10, 16)
			data = append(data, uint16(u64))
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case uint8:
		data := make([]uint8, 0, len(strs))
		for _, str := range strs {
			u64, _ := strconv.ParseUint(str, 10, 8)
			data = append(data, uint8(u64))
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case float64:
		data := make([]float64, 0, len(strs))
		for _, str := range strs {
			f64, _ := strconv.ParseFloat(str, 64)
			data = append(data, f64)
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case float32:
		data := make([]float32, 0, len(strs))
		for _, str := range strs {
			f64, _ := strconv.ParseFloat(str, 32)
			data = append(data, float32(f64))
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	case bool:
		data := make([]bool, 0, len(strs))
		for _, str := range strs {
			b, _ := strconv.ParseBool(str)
			data = append(data, b)
		}
		*x = *(*sliceValue[T])(unsafe.Pointer(&data))
	default:
		log.Printf("sliceValue unsupported type === %T", zero)
	}
	return nil
}

func setField(value reflect.Value, envKey string) (bool, error) {
	envValue := os.Getenv(envKey)
	if envValue == "" {
		return false, nil
	}

	switch value.Kind() {
	case reflect.String:
		value.SetString(strings.TrimSpace(envValue))
	case reflect.Float64, reflect.Float32:
		floatValue, err := strconv.ParseFloat(envValue, 64)
		if err != nil {
			return false, fmt.Errorf("invalid float value for %s: %s", envKey, envValue)
		}
		value.SetFloat(floatValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, err := strconv.ParseInt(envValue, 10, 64)
		if err != nil {
			return false, fmt.Errorf("invalid integer value for %s: %s", envKey, envValue)
		}
		value.SetInt(i64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u64, err := strconv.ParseUint(envValue, 10, 64)
		if err != nil {
			return false, fmt.Errorf("invalid unsigned integer value for %s: %s", envKey, envValue)
		}
		value.SetUint(u64)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(envValue)
		if err != nil {
			return false, fmt.Errorf("invalid boolean value for %s: %s", envKey, envValue)
		}
		value.SetBool(boolValue)
	default:
		return false, fmt.Errorf("unsupported type for %s, %s", envKey, value.Type())
	}

	return true, nil
}

func setSliceField(value reflect.Value, envKey string) error {
	envValue := os.Getenv(envKey)
	if envValue == "" {
		return nil
	}

	elements := strings.Split(envValue, ",")
	sliceType := value.Type().Elem()
	slice := reflect.MakeSlice(value.Type(), len(elements), len(elements))

	for i, element := range elements {
		element = strings.TrimSpace(element)
		switch sliceType.Kind() {
		case reflect.String:
			slice.Index(i).SetString(element)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(element, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value in slice for %s: %s", envKey, element)
			}
			slice.Index(i).SetInt(intValue)
			// uint
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintValue, err := strconv.ParseUint(element, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unsigned integer value in slice for %s: %s", envKey, element)
			}
			slice.Index(i).SetUint(uintValue)
		case reflect.Float64, reflect.Float32:
			floatValue, err := strconv.ParseFloat(element, 64)
			if err != nil {
				return fmt.Errorf("invalid float value in slice for %s: %s", envKey, element)
			}
			slice.Index(i).SetFloat(floatValue)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(element)
			if err != nil {
				return fmt.Errorf("invalid boolean value in slice for %s: %s", envKey, element)
			}
			slice.Index(i).SetBool(boolValue)
		default:
			return fmt.Errorf("unsupported slice type for %s", envKey)
		}
	}

	value.Set(slice)
	return nil
}

var (
	// MarshalOptions is a configurable JSON format marshaller.
	MarshalOptions = protojson.MarshalOptions{
		UseProtoNames:     true,
		EmitUnpopulated:   true,
		EmitDefaultValues: true,
	}
	// UnmarshalOptions is a configurable JSON format parser.
	UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

func Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case json.Marshaler:
		return m.MarshalJSON()
	case proto.Message:
		return MarshalOptions.Marshal(m)
	default:
		return json.Marshal(m)
	}
}

func Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case json.Unmarshaler:
		return m.UnmarshalJSON(data)
	case proto.Message:
		return UnmarshalOptions.Unmarshal(data, m)
	default:
		rv := reflect.ValueOf(v)
		for rv := rv; rv.Kind() == reflect.Ptr; {
			if rv.IsNil() {
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			rv = rv.Elem()
		}
		if m, ok := reflect.Indirect(rv).Interface().(proto.Message); ok {
			return UnmarshalOptions.Unmarshal(data, m)
		}
		return json.Unmarshal(data, m)
	}
}
