package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.165"
	UpstreamECSVersion = "v0.1.162"
)

func ReleaseVersion() string {
	return "v" + Version
}
