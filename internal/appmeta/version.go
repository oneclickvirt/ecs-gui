package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.0"
	UpstreamECSVersion = "v0.2.0"
)

func ReleaseVersion() string {
	return "v" + Version
}
