package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.3"
	UpstreamECSVersion = "v0.2.2"
)

func ReleaseVersion() string {
	return "v" + Version
}
