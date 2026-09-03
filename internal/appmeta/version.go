package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.197"
	UpstreamECSVersion = "v0.1.197"
)

func ReleaseVersion() string {
	return "v" + Version
}
