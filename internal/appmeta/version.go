package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.156"
	UpstreamECSVersion = "v0.1.155"
)

func ReleaseVersion() string {
	return "v" + Version
}
