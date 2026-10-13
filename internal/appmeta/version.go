package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.15"
	UpstreamECSVersion = "v0.2.15"
)

func ReleaseVersion() string {
	return "v" + Version
}
