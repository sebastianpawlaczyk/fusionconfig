package fusionconfig

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type Foo struct {
	ValString string
}

type Configuration struct {
	ValString  string
	ValStruct  Foo
	ValBool    bool
	ValInt     int
	ValInt8    int8
	ValInt16   int16
	ValInt32   int32
	ValInt64   int64
	ValUint    uint
	ValUint8   uint8
	ValUint16  uint16
	ValUint32  uint32
	ValUint64  uint64
	ValFloat32 float32
	ValFloat64 float64
	ValSlice   []string
}

func TestLoadConfig(t *testing.T) {
	type args struct {
		Opt []Option
	}
	type want struct {
		Configuration Configuration
		err           error
	}
	testCases := []struct {
		name     string
		load     func(c Configuration)
		args     args
		want     want
		teardown func()
	}{
		{
			name: "ok read all types with env variables",
			load: func(c Configuration) {
				setEnvs()
			},
			args: args{
				Opt: []Option{},
			},
			want: want{
				Configuration: Configuration{
					ValString: "abc",
					ValStruct: Foo{
						ValString: "abc",
					},
					ValBool:    true,
					ValInt:     42,
					ValInt8:    int8(5),
					ValInt16:   int16(22),
					ValInt32:   int32(24),
					ValInt64:   int64(-45),
					ValUint:    uint(6246),
					ValUint8:   uint8(15),
					ValUint16:  uint16(77),
					ValUint32:  uint32(2516),
					ValUint64:  uint64(156365),
					ValFloat32: float32(245.5),
					ValFloat64: float64(11111.5),
					ValSlice:   []string{"string1", "string2"},
				},
				err: nil,
			},
			teardown: func() {
				os.Clearenv()
			},
		},
		{
			name: "ok with local file",
			load: func(c Configuration) {},
			args: args{
				Opt: []Option{
					WithLocalFile("./fixtures/test-file.json"),
				},
			},
			want: want{
				Configuration: Configuration{
					ValString: "xyz",
					ValStruct: Foo{
						ValString: "abc",
					},
					ValSlice: []string{"a1", "a2"},
				},
				err: nil,
			},
			teardown: func() {
				os.Clearenv()
			},
		},
		{
			name: "ok with prefix",
			load: func(c Configuration) {
				os.Setenv("test.ValString", "abc")
			},
			args: args{
				Opt: []Option{
					WithPrefix("test"),
				},
			},
			want: want{
				Configuration: Configuration{
					ValString: "abc",
				},
				err: nil,
			},
			teardown: func() {
				os.Clearenv()
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.teardown()

			tt.load(tt.want.Configuration)

			cfg := Configuration{}
			err := LoadConfig(&cfg, tt.args.Opt...)

			assert.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.Configuration, cfg)
		})
	}
}

func TestLoadRemoteFile(t *testing.T) {
	jsonResponse := `{
		"ValString": "remote_value",
		"ValStruct": { "ValString": "remote_struct_value" },
		"ValSlice": ["item1", "item2"]
	}`

	// Create a test server that responds with the JSON.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(jsonResponse))
	}))
	defer ts.Close()

	cfg := Configuration{}

	err := LoadConfig(&cfg, WithRemoteFile(ts.URL))

	assert.NoError(t, err)
	assert.Equal(t, "remote_value", cfg.ValString)
	assert.Equal(t, "remote_struct_value", cfg.ValStruct.ValString)
	assert.Equal(t, []string{"item1", "item2"}, cfg.ValSlice)
}

func TestOverrideHierarchy(t *testing.T) {
	cfg := Configuration{}
	os.Setenv("ValString", "from_env_var")
	err := LoadConfig(&cfg)
	assert.NoError(t, err)
	assert.Equal(t, "from_env_var", cfg.ValString) // from env vars
	os.Clearenv()

	cfg = Configuration{}
	os.Setenv("ValString", "from_env_var")
	err = LoadConfig(&cfg, WithLocalFile("fixtures/test-file.json"))
	assert.NoError(t, err)
	assert.Equal(t, "xyz", cfg.ValString) // from local file
	os.Clearenv()

	jsonResponse := `{
		"ValString": "remote_value"
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(jsonResponse))
	}))
	defer ts.Close()

	cfg = Configuration{}
	os.Setenv("ValString", "from_env_var")
	err = LoadConfig(&cfg, WithLocalFile("fixtures/test-file.json"), WithRemoteFile(ts.URL))
	assert.NoError(t, err)
	assert.Equal(t, "remote_value", cfg.ValString) // from remote file
	os.Clearenv()
}

func setEnvs() {
	os.Setenv("ValString", "abc")
	os.Setenv("ValStruct.ValString", "abc")
	os.Setenv("ValBool", "true")
	os.Setenv("ValInt", "42")
	os.Setenv("ValInt8", "5")
	os.Setenv("ValInt16", "22")
	os.Setenv("ValInt32", "24")
	os.Setenv("ValInt64", "-45")
	os.Setenv("ValUint", "6246")
	os.Setenv("ValUint8", "15")
	os.Setenv("ValUint16", "77")
	os.Setenv("ValUint32", "2516")
	os.Setenv("ValUint64", "156365")
	os.Setenv("ValFloat32", "245.5")
	os.Setenv("ValFloat64", "11111.5")
	os.Setenv("ValSlice", "string1,string2")
}
