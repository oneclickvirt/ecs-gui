package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.2.14"
	UpstreamECSVersion = "v0.2.13"
)

func ReleaseVersion() string {
	return "v" + Version
}
