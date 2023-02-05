// Copyright © 2022 Sloan Childers
package sink

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/cretz/bine/tor"
	bine_ed25519 "github.com/cretz/bine/torutil/ed25519"

	"github.com/go-chi/chi/v5"
	"github.com/ipsn/go-libtor"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	gorm_zerolog "github.com/wei840222/gorm-zerolog"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitLogger(level string) {
	parsedLevel, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		log.Fatal().Err(err).Msg("unable to configure logger")
	}
	zerolog.SetGlobalLevel(parsedLevel)
}

func LoadEnv(output interface{}) {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal().Msg(".env")
		return
	}

	if err := env.Parse(output); err != nil {
		log.Fatal().Err(err).Msg("environment")
	}
	// PrintEnvironment()
}

func LoadJson(fileName string, cfg interface{}) error {
	// load collector configuration
	fh, err := os.Open(fileName)
	if err != nil {
		log.Error().Err(err).Str("component", "sink").Str("file", fileName).Msg("load json open")
		return err
	}
	defer fh.Close()

	obj, err := io.ReadAll(fh)
	if err != nil {
		log.Error().Err(err).Str("component", "sink").Str("file", fileName).Msg("load json read")
		return err
	}
	err = json.Unmarshal(obj, cfg)
	if err != nil {
		log.Error().Err(err).Str("component", "sink").Str("file", fileName).Msg("load json parse")
		return err
	}
	return nil
}

func ListenAndServe(ListenAddr, SSLCertFile, SSLKeyFile string, router http.Handler) error {
	server := &http.Server{Addr: ListenAddr, Handler: router}
	var err error
	if SSLCertFile != "" {
		err = server.ListenAndServeTLS(SSLCertFile, SSLKeyFile)
	} else {
		err = server.ListenAndServe()
	}
	return err
}

// NOTE:  to use Tor you'll need to do the following
// go mod init && go mod tidy
// go get -u -a -v -x github.com/ipsn/go-libtor
// go get github.com/cretz/bine/tor@v0.2.0
// go run main.go
// go grab a soda and wait for libtor to build
func ListenAndServeTor(certFile string, auths map[string]string, router http.Handler) error {
	torServer, err := tor.Start(
		context.Background(),
		&tor.StartConf{
			DataDir:        "/tmp/",
			ProcessCreator: libtor.Creator,
			DebugWriter:    log.Logger})
	if err != nil {
		log.Error().Err(err).Str("component", "sink").Str("file", certFile).Msg("start tor")
		return err
	}
	defer torServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pk, err := GetPrivateKey(certFile)
	if err != nil {
		log.Error().Err(err).Str("component", "sink").Str("file", certFile).Msg("load private key")
		return err
	}

	onion, err := torServer.Listen(ctx, &tor.ListenConf{
		RemotePorts: []int{80},
		Key:         bine_ed25519.FromCryptoPrivateKey(pk),
		Version3:    true,
		DiscardKey:  false,
		ClientAuths: auths})
	if err != nil {
		log.Error().Err(err).Str("component", "sink").Str("file", certFile).Msg("start onion service")
		return err
	}
	log.Info().Str("component", "sink").Str("tor addr", "http://"+onion.ID+".onion")
	fmt.Println("http://" + onion.ID + ".onion")
	defer onion.Close()
	return http.Serve(onion, router)
}

func PrintEnvironment() {
	env := os.Environ()
	for _, variable := range env {
		pair := strings.Split(variable, "=")
		value := pair[1]
		if strings.Contains(pair[0], "API_KEY") || strings.Contains(pair[0], "PASS") || strings.Contains(pair[0], "USER") {
			value = "<masked>"
		}
		log.Info().Str(pair[0], value).Msg("environment")
	}
}

type PostgresConfig struct {
	PgHost     string
	PgPort     string
	PgUser     string
	PgPassword string
	PgDB       string
}

func GetDSN(cfg *PostgresConfig) string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable timezone=utc",
		cfg.PgHost,
		cfg.PgPort,
		cfg.PgUser,
		cfg.PgDB,
		cfg.PgPassword)
}

func OpenDB(cfg *PostgresConfig) *gorm.DB {
	dsn := GetDSN(cfg)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gorm_zerolog.New()})
	if err != nil {
		log.Fatal().Err(err).Str("DSN", dsn).Msg("GORM open")
	}
	return db
}

func Param(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func SendError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": err.Error(),
	})
}

func SendPrettyJSON(ctx context.Context, w http.ResponseWriter, data interface{}) {
	span, _ := tracer.StartSpanFromContext(ctx, "rendering_json")
	defer span.Finish()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(data)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Interface("data", data).Msg("unable to pretty json")
	}
}

func GetPrivateKey(fileName string) (ed25519.PrivateKey, error) {
	p, _ := decodePEMFile(fileName)
	key, err := x509.ParsePKCS8PrivateKey(p)
	if err != nil {
		return nil, err
	}
	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not ed25519 key")
	}
	return ed25519.PrivateKey(edKey), nil
}

func GetPublicKey(publicKey string) (ed25519.PublicKey, error) {
	p, _ := decodePEMFile(publicKey)
	key, err := x509.ParsePKIXPublicKey(p)
	if err != nil {
		return nil, err
	}
	edKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not ed25519 key")
	}
	return ed25519.PublicKey(edKey), nil
}

func decodePEMFile(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	p, _ := pem.Decode(buf)
	if p == nil {
		return nil, fmt.Errorf("no pem block found")
	}
	return p.Bytes, nil
}
