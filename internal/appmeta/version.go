package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.9"
	UpstreamECSVersion = "v0.2.7"
)

func ReleaseVersion() string {
	return "v" + Version
}
