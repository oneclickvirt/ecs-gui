package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.185"
	UpstreamECSVersion = "v0.1.182"
)

func ReleaseVersion() string {
	return "v" + Version
}
