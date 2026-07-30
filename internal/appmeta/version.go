package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.177"
	UpstreamECSVersion = "v0.1.174"
)

func ReleaseVersion() string {
	return "v" + Version
}
