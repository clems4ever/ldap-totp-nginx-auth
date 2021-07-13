package server

import (
	"embed"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/expvarhandler"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/valyala/fasthttp/pprofhandler"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/duo"
	"github.com/authelia/authelia/internal/handlers"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/middlewares"
)

//go:embed public_html
var assets embed.FS

func registerRoutes(configuration schema.Configuration, providers middlewares.Providers) fasthttp.RequestHandler {
	autheliaMiddleware := middlewares.AutheliaMiddleware(configuration, providers)
	rememberMe := strconv.FormatBool(configuration.Session.RememberMeDuration != "0")
	resetPassword := strconv.FormatBool(!configuration.AuthenticationBackend.DisableResetPassword)

	embeddedPath, _ := fs.Sub(assets, "public_html")
	embeddedFS := fasthttpadaptor.NewFastHTTPHandler(http.FileServer(http.FS(embeddedPath)))
	rootFiles := []string{"favicon.ico", "manifest.json", "robots.txt"}

	serveIndexHandler := ServeTemplatedFile(embeddedAssets, indexFile, configuration.Server.Path, rememberMe, resetPassword, configuration.Session.Name, configuration.Theme)
	serveSwaggerHandler := ServeTemplatedFile(swaggerAssets, indexFile, configuration.Server.Path, rememberMe, resetPassword, configuration.Session.Name, configuration.Theme)
	serveSwaggerAPIHandler := ServeTemplatedFile(swaggerAssets, apiFile, configuration.Server.Path, rememberMe, resetPassword, configuration.Session.Name, configuration.Theme)

	r := router.New()
	r.GET("/", serveIndexHandler)
	r.GET("/api/", serveSwaggerHandler)
	r.GET("/api/"+apiFile, serveSwaggerAPIHandler)

	for _, f := range rootFiles {
		r.GET("/"+f, embeddedFS)
	}

	r.GET("/static/{filepath:*}", embeddedFS)
	r.ANY("/api/{filepath:*}", embeddedFS)

	r.GET("/api/health", autheliaMiddleware(handlers.HealthGet))
	r.GET("/api/state", autheliaMiddleware(handlers.StateGet))

	r.GET("/api/configuration", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.ConfigurationGet)))

	r.GET("/api/verify", autheliaMiddleware(handlers.VerifyGet(configuration.AuthenticationBackend)))
	r.HEAD("/api/verify", autheliaMiddleware(handlers.VerifyGet(configuration.AuthenticationBackend)))

	r.POST("/api/firstfactor", autheliaMiddleware(handlers.FirstFactorPost(1000, true)))
	r.POST("/api/logout", autheliaMiddleware(handlers.LogoutPost))

	// Only register endpoints if forgot password is not disabled.
	if !configuration.AuthenticationBackend.DisableResetPassword {
		// Password reset related endpoints.
		r.POST("/api/reset-password/identity/start", autheliaMiddleware(
			handlers.ResetPasswordIdentityStart))
		r.POST("/api/reset-password/identity/finish", autheliaMiddleware(
			handlers.ResetPasswordIdentityFinish))
		r.POST("/api/reset-password", autheliaMiddleware(
			handlers.ResetPasswordPost))
	}

	// Information about the user.
	r.GET("/api/user/info", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.UserInfoGet)))
	r.POST("/api/user/info/2fa_method", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.MethodPreferencePost)))

	// TOTP related endpoints.
	r.POST("/api/secondfactor/totp/identity/start", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityStart)))
	r.POST("/api/secondfactor/totp/identity/finish", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityFinish)))
	r.POST("/api/secondfactor/totp", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPPost(&handlers.TOTPVerifierImpl{
			Period: uint(configuration.TOTP.Period),
			Skew:   uint(*configuration.TOTP.Skew),
		}))))

	// U2F related endpoints.
	r.POST("/api/secondfactor/u2f/identity/start", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FIdentityStart)))
	r.POST("/api/secondfactor/u2f/identity/finish", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FIdentityFinish)))

	r.POST("/api/secondfactor/u2f/register", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FRegister)))

	r.POST("/api/secondfactor/u2f/sign_request", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FSignGet)))

	r.POST("/api/secondfactor/u2f/sign", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FSignPost(&handlers.U2FVerifierImpl{}))))

	// Configure DUO api endpoint only if configuration exists.
	if configuration.DuoAPI != nil {
		var duoAPI duo.API
		if os.Getenv("ENVIRONMENT") == dev {
			duoAPI = duo.NewDuoAPI(duoapi.NewDuoApi(
				configuration.DuoAPI.IntegrationKey,
				configuration.DuoAPI.SecretKey,
				configuration.DuoAPI.Hostname, "", duoapi.SetInsecure()))
		} else {
			duoAPI = duo.NewDuoAPI(duoapi.NewDuoApi(
				configuration.DuoAPI.IntegrationKey,
				configuration.DuoAPI.SecretKey,
				configuration.DuoAPI.Hostname, ""))
		}

		r.GET("/api/secondfactor/duo_devices", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoDevicesGet(duoAPI))))

		r.POST("/api/secondfactor/duo", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoPost(duoAPI))))

		r.POST("/api/secondfactor/duo_device", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoDevicePost)))
	}

	if configuration.Server.EnablePprof {
		r.GET("/debug/pprof/{name?}", pprofhandler.PprofHandler)
	}

	if configuration.Server.EnableExpvars {
		r.GET("/debug/vars", expvarhandler.ExpvarHandler)
	}

	r.NotFound = serveIndexHandler

	handler := middlewares.LogRequestMiddleware(r.Handler)
	if configuration.Server.Path != "" {
		handler = middlewares.StripPathMiddleware(handler)
	}

	if providers.OpenIDConnect.Fosite != nil {
		handlers.RegisterOIDC(r, autheliaMiddleware)
	}

	return handler
}

// StartServer start Authelia server with the given configuration and providers.
func StartServer(configuration schema.Configuration, providers middlewares.Providers) {
	logger := logging.Logger()

	handler := registerRoutes(configuration, providers)

	server := &fasthttp.Server{
		ErrorHandler:          autheliaErrorHandler,
		Handler:               handler,
		NoDefaultServerHeader: true,
		ReadBufferSize:        configuration.Server.ReadBufferSize,
		WriteBufferSize:       configuration.Server.WriteBufferSize,
	}

	addrPattern := net.JoinHostPort(configuration.Host, strconv.Itoa(configuration.Port))

	listener, err := net.Listen("tcp", addrPattern)
	if err != nil {
		logger.Fatalf("Error initializing listener: %s", err)
	}

	// TODO(clems4ever): move that piece to a more related location, probably in the configuration package.
	if configuration.AuthenticationBackend.File != nil && configuration.AuthenticationBackend.File.Password.Algorithm == "argon2id" && runtime.GOOS == "linux" {
		f, err := ioutil.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes")
		if err != nil {
			logger.Warnf("Error reading hosts memory limit: %s", err)
		} else {
			m, _ := strconv.Atoi(strings.TrimSuffix(string(f), "\n"))
			hostMem := float64(m) / 1024 / 1024 / 1024
			argonMem := float64(configuration.AuthenticationBackend.File.Password.Memory) / 1024

			if hostMem/argonMem <= 2 {
				logger.Warnf("Authelia's password hashing memory parameter is set to: %gGB this is %g%% of the available memory: %gGB", argonMem, argonMem/hostMem*100, hostMem)
				logger.Warn("Please read https://www.authelia.com/docs/configuration/authentication/file.html#memory and tune your deployment")
			}
		}
	}

	if configuration.TLSCert != "" && configuration.TLSKey != "" {
		logger.Infof("Authelia is listening for TLS connections on %s%s", addrPattern, configuration.Server.Path)
		logger.Fatal(server.ServeTLS(listener, configuration.TLSCert, configuration.TLSKey))
	} else {
		logger.Infof("Authelia is listening for non-TLS connections on %s%s", addrPattern, configuration.Server.Path)
		logger.Fatal(server.Serve(listener))
	}
}
