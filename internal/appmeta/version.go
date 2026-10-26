package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.17"
	UpstreamECSVersion = "v0.2.17"
)

func ReleaseVersion() string {
	return "v" + Version
}
