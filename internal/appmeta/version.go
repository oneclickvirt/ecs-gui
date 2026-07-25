package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.175"
	UpstreamECSVersion = "v0.1.172"
)

func ReleaseVersion() string {
	return "v" + Version
}
