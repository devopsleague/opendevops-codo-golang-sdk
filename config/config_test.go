package config

import (
	"flag"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/opendevops-cn/codo-golang-sdk/config/testdata"
	"github.com/spf13/pflag"
)

func TestLoadConfig(t *testing.T) {
	type testConfig struct {
		Host          string  `json:"host" env:"HOST"`
		NoEnv         uint32  `json:"noEnv"`
		InJSON        float32 `json:"in_json"`
		OnlyNameSlice []float32
		Bool          bool

		TestFlag  uint   `flag:"test-flag|t" usage:"测试 flag 使用"`
		TestFlags []bool `flag:"test-bools" usage:"测试 flag 数组"`
	}

	var dst testConfig

	os.Setenv("TEST_HOST", "123")
	os.Setenv("TEST_NOENV", "456")
	os.Setenv("TEST_IN_JSON", "123.456")
	os.Setenv("TEST_ONLYNAMESLICE", "123.456,1.1,2.2")
	os.Setenv("TEST_BOOL", "true")

	type args struct {
		dst  interface{}
		opts []ConfigOption
	}
	tests := []struct {
		name string
		args args
		want *testConfig
	}{
		{
			name: "1",
			args: args{
				dst: &dst,
				opts: []ConfigOption{
					WithEnv("TEST"),
					WithFlag(flag.NewFlagSet("test", flag.ExitOnError), []string{"--test-flag=2", "--test-bools=true,false,1,0"}),
				},
			},
			want: &testConfig{
				Host:          "",
				NoEnv:         456,
				InJSON:        123.456,
				OnlyNameSlice: []float32{123.456, 1.1, 2.2},
				Bool:          true,
				TestFlag:      2,
				TestFlags:     []bool{true, false, true, false},
			},
		},
		{
			name: "2",
			args: args{
				dst: &dst,
				opts: []ConfigOption{
					WithEnv("TEST"),
					WithPFlag(pflag.NewFlagSet("ptest", pflag.ExitOnError), []string{"-t=3", "--test-bools=true,false,1,0"}),
				},
			},
			want: &testConfig{
				Host:          "",
				NoEnv:         456,
				InJSON:        123.456,
				OnlyNameSlice: []float32{123.456, 1.1, 2.2},
				Bool:          true,
				TestFlag:      3,
				TestFlags:     []bool{true, false, true, false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LoadConfig(tt.args.dst, tt.args.opts...)
			if err != nil {
				panic(err)
			}
			if !reflect.DeepEqual(tt.args.dst, tt.want) {
				t.Errorf("%+v\n", tt.args.dst)
				return
			}
			t.Logf("test config === %+v\n", tt.args.dst)
		})
	}
}

func TestLoadConfig2(t *testing.T) {
	os.Setenv("CODO_PROMETHEUS_PATH", "/metrics-test-env")
	var cfg testdata.Bootstrap
	err := LoadConfig(&cfg,
		WithYaml("testdata/config.yaml"),
	)
	if err != nil {
		panic(err)
	}
	log.Println(cfg)
	if cfg.PROMETHEUS.PATH != "/metrics-test-env" {
		t.Errorf("load config error")
	}
}
