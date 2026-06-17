package app

const (
	DefaultAddress         = ":8080"
	DefaultDataDir         = "."
	DefaultConfigDirName   = "config"
	DefaultUploadSecret    = "upload_secret"
	DefaultUploadToken     = "upload_token"
	DefaultSubSecret       = "sub_secret"
	DefaultSubscribe       = "subscribe"
	DefaultBackendURL      = "https://subc.020.name/sub?"
	DefaultRemoteConfigURL = "https://github.com/AoEiuV020/SubConfig/raw/main/subconverter.ini"
	DefaultUploadIVBase64  = "EJwC9OfO/fkuTvPax7YHeQ=="
	DefaultHTTPTimeoutSec  = 30
)

const (
	EnvAddress         = "CONFIG_DEPOT_ADDR"
	EnvDataDir         = "CONFIG_DEPOT_DATA_DIR"
	EnvConfigDir       = "CONFIG_DEPOT_CONFIG_DIR"
	EnvUploadSecret    = "CONFIG_DEPOT_UPLOAD_SECRET_FILE"
	EnvUploadToken     = "CONFIG_DEPOT_UPLOAD_TOKEN_FILE"
	EnvSubSecret       = "CONFIG_DEPOT_SUB_SECRET_FILE"
	EnvSubscribe       = "CONFIG_DEPOT_SUBSCRIBE_FILE"
	EnvBackendURL      = "CONFIG_DEPOT_BACKEND_URL"
	EnvRemoteConfigURL = "CONFIG_DEPOT_REMOTE_CONFIG_URL"
)
