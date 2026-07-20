package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.151"
	UpstreamECSVersion = "v0.1.151"
)

func ReleaseVersion() string {
	return "v" + Version
}
