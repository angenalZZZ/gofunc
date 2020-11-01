package configfile

import "testing"

func TestYamlToMap(t *testing.T) {
	config, err := YamlToMap("../test/config/database.yaml")
	if err != nil {
		t.Fatal(err)
	}
	conn, ok := config["database"].(map[interface{}]interface{})
	if !ok {
		t.SkipNow()
	}
	t.Logf("database.mssql: %s", conn["mssql"])
	t.Logf("database.mysql: %s", conn["mysql"])
}
