package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.170"
	UpstreamECSVersion = "v0.1.166"
)

func ReleaseVersion() string {
	return "v" + Version
}
