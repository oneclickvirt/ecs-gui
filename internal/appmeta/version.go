package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.162"
	UpstreamECSVersion = "v0.1.160"
)

func ReleaseVersion() string {
	return "v" + Version
}
