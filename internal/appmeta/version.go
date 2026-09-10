package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.12"
	UpstreamECSVersion = "v0.2.11"
)

func ReleaseVersion() string {
	return "v" + Version
}
