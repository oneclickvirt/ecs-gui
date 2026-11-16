package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.20"
	UpstreamECSVersion = "v0.2.24"
)

func ReleaseVersion() string {
	return "v" + Version
}
