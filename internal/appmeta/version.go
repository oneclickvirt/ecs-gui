package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.157"
	UpstreamECSVersion = "v0.1.156"
)

func ReleaseVersion() string {
	return "v" + Version
}
