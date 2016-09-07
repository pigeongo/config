package config

import (
    "os"
    "fmt"
    "regexp"
    "reflect"
)

type Handler interface {
    Set(key string, val string) error
    String(key, def string) string
    Strings(key string, def []string) []string
    Int(key string, def int) int
    Int64(key string, def int64) int64
    Bool(key string, def bool) bool
    Float(key string, def float64) float64
    Slice(key string, def []interface{}) []interface{}
}

type Config interface {
    Reader(filename string) (Handler, error)
    Writer(data []byte) (Handler, error)
}

var adapters = make(map[string]Config)

func Register(name string, adapter Config) {
    if adapter == nil {
        panic("config: Register adapter is nil")
    }
    if _, ok := adapters[name]; ok == true {
        panic("config: Register called twice for adapter " + name)
    }
    adapters[name] = adapter
}

func NewConfig(adapterName, filename string) (Handler, error) {
    if adapter, ok := adapters[adapterName]; ok {
        return adapter.Reader(filename)
    } else {
        return nil, fmt.Errorf("config: unknown adapter %q", adapterName)
    }
}

func EnvValue(name string) string {
    match := regexp.MustCompile(`^\$\{([a-zA-Z_][\w]+)(\|\|[./\w])*}$`).FindAllStringSubmatch(name, -1)
    if len(match) > 0 {
        switch {
        case len(match[0]) > 2:
            val := os.Getenv(match[0][1])
            if val != "" {
                return val
            }
            return match[0][2]
        case len(match[0]) > 1:
            val := os.Getenv(match[0][1])
            if val != "" {
                return val
            }
        }
    }
    return name
}

func EnvForSlice(s []interface{}) []interface{} {
    v := reflect.ValueOf(s)
    tmp := make([]interface{}, v.Len(), v.Len())
    for i := 0; i < v.Len(); i++ {
        vv := v.Index(i).Interface()
        switch vv := vv.(type) {
        case string:
            tmp[i] = EnvValue(vv)
        case []interface{}:
            tmp[i] = EnvForSlice(vv)
        case map[interface{}]interface{}:
            tmpMap := make(map[string]interface{})
            for key, value := range vv {
                tmpMap[key.(string)] = value
            }
            tmp[i] = EnvValueForMap(tmpMap)
        }
    }
    return tmp
}

func EnvValueForMap(m map[string]interface{}) map[string]interface{} {
    for k, v := range m {
        switch v := v.(type) {
        case string:
            m[k] = EnvValue(v)
        case []interface{}:
            m[k] = EnvForSlice(v)
        case map[interface{}]interface{}:
            tmpMap := make(map[string]interface{})
            for key, value := range v {
                tmpMap[key.(string)] = value
            }
            m[k] = EnvValueForMap(tmpMap)
        }
    }
    return m
}