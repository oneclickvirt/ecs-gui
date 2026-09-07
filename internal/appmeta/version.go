package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.6"
	UpstreamECSVersion = "v0.2.4"
)

func ReleaseVersion() string {
	return "v" + Version
}
