package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.193"
	UpstreamECSVersion = "v0.1.194"
)

func ReleaseVersion() string {
	return "v" + Version
}
