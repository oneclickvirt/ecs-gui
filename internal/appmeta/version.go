package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.8"
	UpstreamECSVersion = "v0.2.6"
)

func ReleaseVersion() string {
	return "v" + Version
}
