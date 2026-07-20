package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.153"
	UpstreamECSVersion = "v0.1.153"
)

func ReleaseVersion() string {
	return "v" + Version
}
