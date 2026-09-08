package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.10"
	UpstreamECSVersion = "v0.2.9"
)

func ReleaseVersion() string {
	return "v" + Version
}
