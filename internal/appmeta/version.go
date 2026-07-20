package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.148"
	UpstreamECSVersion = "v0.1.148"
)

func ReleaseVersion() string {
	return "v" + Version
}
