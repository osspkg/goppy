package search

const (
	indexFilename = "/store/root.bolt"

	customAnalyzerName = "goppy_analyzer"
	customFilterName   = "goppy_filter"

	structScoreField = "Score"

	FieldText = "text"
	FieldDate = "date"
)

type (
	Config struct {
		Search ConfigItem `yaml:"search"`
	}
	ConfigItem struct {
		Folder  string        `yaml:"folder"`
		Indexes []ConfigIndex `yaml:"indexes"`
	}
	ConfigIndex struct {
		Name   string             `yaml:"name"`
		Fields []ConfigIndexField `yaml:"fields"`
	}
	ConfigIndexField struct {
		Name string `yaml:"name"`
		Type string `yaml:"type"`
	}
)
