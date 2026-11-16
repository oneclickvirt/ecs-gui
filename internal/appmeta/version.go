package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.21"
	UpstreamECSVersion = "v0.2.27"
)

func ReleaseVersion() string {
	return "v" + Version
}
