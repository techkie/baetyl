package master

import (
	"reflect"
	"testing"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/stretchr/testify/assert"
)

var cfgV1 = `
version: V1
services:
  - name: a
    image: 'baetyl-a:latest'
    replica: 1
    mounts:
      - name: a-conf-V1
        path: etc/baetyl
        readonly: true
  - name: b
    image: 'baetyl-b:latest'
    replica: 1
    mounts:
      - name: b-conf-V1
        path: etc/baetyl
        readonly: true
  - name: c
    image: 'baetyl-c:latest'
    replica: 1
    mounts:
      - name: c-conf-V1
        path: etc/baetyl
        readonly: true
volumes:
  - name: a-conf-V1
    path: a-conf/V1
  - name: b-conf-V1
    path: b-conf/V1
  - name: c-conf-V1
    path: c-conf/V1
`

var cfgV2 = `
version: V2
services:
  - name: a
    image: 'baetyl-a:latest'
    replica: 1
    mounts:
      - name: a-conf-V1
        path: etc/baetyl
        readonly: true
  - name: b
    image: 'baetyl-b:0.1.4'
    replica: 1
    mounts:
      - name: b-conf-V1
        path: etc/baetyl
        readonly: true
  - name: d
    image: 'baetyl-d:latest'
    replica: 1
    mounts:
      - name: d-conf-V1
        path: etc/baetyl
        readonly: true
volumes:
  - name: a-conf-V1
    path: a-conf/V1
  - name: b-conf-V1
    path: b-conf/V1
  - name: d-conf-V1
    path: d-conf/V1
`

var cfgV3 = `
version: V3
services:
  - name: a
    image: 'baetyl-a:latest'
    replica: 0
    mounts:
      - name: a-conf-V1
        path: etc/baetyl
        readonly: true
  - name: b
    image: 'baetyl-b:0.1.4'
    replica: 1
    mounts:
      - name: b-conf-V1
        path: etc/baetyl
        readonly: true
      - name: b-data-V1
        path: var/db/baetyl/data
  - name: d
    image: 'baetyl-d:latest'
    replica: 1
    mounts:
      - name: d-conf-V1
        path: etc/baetyl
        readonly: true
volumes:
  - name: a-conf-V1
    path: a-conf/V1
  - name: b-conf-V1
    path: b-conf/V1
  - name: d-conf-V1
    path: d-conf/V22
`

var cfgV4 = `
version: V4
`

func Test_diffServices(t *testing.T) {
	var V1 baetyl.AppConfig
	err := utils.UnmarshalYAML([]byte(cfgV1), &V1)
	assert.NoError(t, err)

	var V2 baetyl.AppConfig
	err = utils.UnmarshalYAML([]byte(cfgV2), &V2)
	assert.NoError(t, err)

	var V3 baetyl.AppConfig
	err = utils.UnmarshalYAML([]byte(cfgV3), &V3)
	assert.NoError(t, err)

	var V4 baetyl.AppConfig
	err = utils.UnmarshalYAML([]byte(cfgV4), &V4)
	assert.NoError(t, err)

	type args struct {
		cur baetyl.AppConfig
		old baetyl.AppConfig
	}
	tests := []struct {
		name string
		args args
		want map[string]struct{}
	}{
		{
			name: "a,b,c --> a,b',d",
			args: args{
				cur: V2,
				old: V1,
			},
			want: map[string]struct{}{
				"a": struct{}{},
			},
		},
		{
			name: "a,b,d --> a',b',d'",
			args: args{
				cur: V3,
				old: V2,
			},
			want: map[string]struct{}{},
		},
		{
			name: "a,b,d --> nil",
			args: args{
				cur: V4,
				old: V3,
			},
			want: map[string]struct{}{},
		},
		{
			name: "nil --> a,b,d",
			args: args{
				cur: V3,
				old: V4,
			},
			want: map[string]struct{}{},
		},
		{
			name: "a,b,d --> a,b,d",
			args: args{
				cur: V3,
				old: V3,
			},
			want: map[string]struct{}{
				"a": struct{}{},
				"b": struct{}{},
				"d": struct{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ccur := tt.args.cur.ToComposeAppConfig()
			cold := tt.args.old.ToComposeAppConfig()
			if got := diffServices(ccur, cold); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("diffServices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceSort(t *testing.T) {
	services := map[string]baetyl.ComposeService{}
	services["a"] = baetyl.ComposeService{
		DependsOn: []string{},
	}
	services["b"] = baetyl.ComposeService{
		DependsOn: []string{"a"},
	}
	services["c"] = baetyl.ComposeService{
		DependsOn: []string{"a", "b"},
	}
	services["d"] = baetyl.ComposeService{
		DependsOn: []string{"b", "c"},
	}
	services["e"] = baetyl.ComposeService{
		DependsOn: []string{"c", "a", "b"},
	}
	services["f"] = baetyl.ComposeService{
		DependsOn: []string{"b", "c"},
	}
	services["h"] = baetyl.ComposeService{
		DependsOn: []string{"d", "f"},
	}
	order := ServiceSort(services)
	om := map[string]int{}
	for i, o := range order {
		om[o] = i
	}
	// order of depended services are less than the service
	assert.True(t, om["a"] < om["b"])
	assert.True(t, om["a"] < om["c"])
	assert.True(t, om["b"] < om["c"])
	assert.True(t, om["b"] < om["d"])
	assert.True(t, om["c"] < om["d"])
	assert.True(t, om["a"] < om["e"])
	assert.True(t, om["b"] < om["e"])
	assert.True(t, om["c"] < om["e"])
	assert.True(t, om["b"] < om["f"])
	assert.True(t, om["c"] < om["f"])
	assert.True(t, om["d"] < om["h"])
	assert.True(t, om["f"] < om["h"])
}
