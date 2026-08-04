package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.178"
	UpstreamECSVersion = "v0.1.175"
)

func ReleaseVersion() string {
	return "v" + Version
}
