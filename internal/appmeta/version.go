package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.155"
	UpstreamECSVersion = "v0.1.154"
)

func ReleaseVersion() string {
	return "v" + Version
}
