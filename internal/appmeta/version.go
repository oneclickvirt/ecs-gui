package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.189"
	UpstreamECSVersion = "v0.1.185"
)

func ReleaseVersion() string {
	return "v" + Version
}
