package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.11"
	UpstreamECSVersion = "v0.2.10"
)

func ReleaseVersion() string {
	return "v" + Version
}
