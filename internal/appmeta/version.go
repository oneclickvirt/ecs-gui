package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.164"
	UpstreamECSVersion = "v0.1.161"
)

func ReleaseVersion() string {
	return "v" + Version
}
