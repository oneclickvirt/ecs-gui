package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.2"
	UpstreamECSVersion = "v0.2.1"
)

func ReleaseVersion() string {
	return "v" + Version
}
