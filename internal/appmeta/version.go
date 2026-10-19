package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.16"
	UpstreamECSVersion = "v0.2.16"
)

func ReleaseVersion() string {
	return "v" + Version
}
