package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.192"
	UpstreamECSVersion = "v0.1.193"
)

func ReleaseVersion() string {
	return "v" + Version
}
