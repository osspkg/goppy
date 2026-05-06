# Config Resolver

Updating the config through resolver variables.

## Config example

update config via ENV

```text
@env(key#defaut value)
```

```yaml
envs:
  home: "@env(HOME#/tmp/home)"
  path: "@env(PATH#/usr/local/bin)"
```

```go
import (
  "go.osspkg.com/goppy/v3/pkg/config"
)

type (
  ConfigItem struct {
    Home string `yaml:"home"`
    Path string `yaml:"path"`
  }
  Config struct {
    Envs ConfigItem `yaml:"envs"`
  }
)

func main() {
  conf := Config{}
  
  res := config.New(
    config.NewEnvResolver(),
  )
  res.OpenFile("./config.yaml") // open config file
  res.Build() // prepare config with resolvers
  res.Decode(&conf) // decoding config
  
  fmt.Println(conf.Envs.Home)
  fmt.Println(conf.Envs.Path)
}

```