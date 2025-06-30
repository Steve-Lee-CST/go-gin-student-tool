package config

var Base = struct {
	ServiceName string
	Protocol    string
	Domain      string
	Port        string
}{
	ServiceName: "stu-tool", // Service name for the application
	Protocol:    "http",
	Domain:      "127.0.0.1",
	Port:        "80",
}
