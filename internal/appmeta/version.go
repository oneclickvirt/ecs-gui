package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.180"
	UpstreamECSVersion = "v0.1.177"
)

func ReleaseVersion() string {
	return "v" + Version
}
