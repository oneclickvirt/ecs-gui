package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.19"
	UpstreamECSVersion = "v0.2.20"
)

func ReleaseVersion() string {
	return "v" + Version
}
