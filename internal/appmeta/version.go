package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.147"
	UpstreamECSVersion = "v0.1.147"
)

func ReleaseVersion() string {
	return "v" + Version
}
