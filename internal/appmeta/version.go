package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.191"
	UpstreamECSVersion = "v0.1.186"
)

func ReleaseVersion() string {
	return "v" + Version
}
