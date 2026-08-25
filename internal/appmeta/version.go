package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.187"
	UpstreamECSVersion = "v0.1.183"
)

func ReleaseVersion() string {
	return "v" + Version
}
