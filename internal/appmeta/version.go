package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.172"
	UpstreamECSVersion = "v0.1.169"
)

func ReleaseVersion() string {
	return "v" + Version
}
