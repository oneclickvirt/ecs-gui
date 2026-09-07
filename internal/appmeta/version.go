package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.5"
	UpstreamECSVersion = "v0.2.3"
)

func ReleaseVersion() string {
	return "v" + Version
}
