package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.171"
	UpstreamECSVersion = "v0.1.167"
)

func ReleaseVersion() string {
	return "v" + Version
}
