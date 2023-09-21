// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/spf13/cobra"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb/chttp"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/config"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/log"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/output/friendly"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/output/gotmpl"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/output/json"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/output/raw"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/output/yaml"
)

type root struct {
	confFile string
	debug    bool
	log      log.Logger
	conf     *config.Config
	cmd      *cobra.Command
	fmt      *output.Formatter

	requestTimeout       string
	parsedRequestTimeout time.Duration
	retryDelay           string
	connectTimeout       string
	parsedConnectTimeout time.Duration
	retryTimeout         string
	options              map[string]interface{}
	stringOptions        map[string]string
	boolOptions          map[string]string

	trace      *chttp.ClientTrace
	dumpHeader bool
	verbose    bool

	// cl *kivik.Client

	// retry attempts
	retryCount         int
	retryDelayParsed   time.Duration
	retryTimeoutParsed time.Duration

	// resolveHome is used to resolve ~ in the default config file path
	resolveHome func(string) string
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	lg := log.New()
	root := rootCmd(lg)
	os.Exit(root.execute(ctx))
}

func (r *root) execute(ctx context.Context) int {
	ctx = chttp.WithClientTrace(ctx, r.clientTrace())
	err := r.cmd.ExecuteContext(ctx)
	if err == nil {
		return 0
	}
	return extractExitCode(err)
}

func extractExitCode(err error) int {
	if code := errors.InspectErrorCode(err); code != 0 {
		return code
	}

	// Any unhandled errors are assumed to be from Cobra, so return a "failed
	// to initialize" error
	return errors.ErrUsage
}

func formatter() *output.Formatter {
	f := output.New()
	f.Register("", friendly.New())
	f.Register("json", json.New())
	f.Register("raw", raw.New())
	f.Register("yaml", yaml.New())
	f.Register("go-template", gotmpl.New())
	return f
}

func resolveHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, path[2:])
}

func rootCmd(lg log.Logger) *root {
	r := &root{
		log:         lg,
		fmt:         formatter(),
		resolveHome: resolveHome,
	}
	r.cmd = &cobra.Command{
		Use:               "kivik",
		Short:             "kivik facilitates controlling CouchDB instances",
		Long:              `This tool makes it easier to administrate and interact with CouchDB's HTTP API`,
		PersistentPreRunE: r.init,
		RunE:              r.RunE,
	}
	r.conf = config.New(func() {
		r.cmd.SilenceUsage = true
	})

	pf := r.cmd.PersistentFlags()

	r.fmt.ConfigFlags(pf)
	pf.StringVar(&r.confFile, "config", "~/.kivik/config", "Path to config file to use for CLI requests")
	pf.BoolVar(&r.debug, "debug", false, "Enable debug output")
	pf.IntVar(&r.retryCount, "retry", 0, "In case of transient error, retry up to this many times. A negative value retries forever.")
	pf.StringToStringVarP(&r.stringOptions, "option", "O", nil, "CouchDB string option, specified as key=value. May be repeated.")
	pf.StringToStringVarP(&r.boolOptions, "option-bool", "B", nil, "CouchDb bool option, specified as key=value. May be repeated.")
	pf.BoolVarP(&r.dumpHeader, "header", "H", false, "Output response header")
	pf.BoolVarP(&r.verbose, "verbose", "v", false, "Output bi-directional network traffic")

	// Timeouts
	// Might consider adding:
	// - http.Transport.TLSHandshakeTimeout
	// - http.Transport.ResponseHeaderTimeout
	// - http.Transport.ExpectContinueTimeout (not sure this is relevant, as I'm not sure CouchDB ever uses a 100)
	// - Read timeout (would have to be an HTTP transport that wraps the Response.Body reader with a context-aware reader that extends the timeout every time more data is read)
	pf.StringVar(&r.requestTimeout, "request-timeout", "", "The time limit for each request.")
	pf.StringVar(&r.retryDelay, "retry-delay", "", "Delay between retry attempts. Disables the default exponential backoff algorithm.")
	pf.StringVar(&r.connectTimeout, "connect-timeout", "", "Limits the time spent establishing a TCP connection.")
	pf.StringVar(&r.retryTimeout, "retry-timeout", "", "When used with --retry, no more retries will be attempted after this timeout.")

	r.cmd.AddCommand(getCmd(r))
	r.cmd.AddCommand(pingCmd(r))
	r.cmd.AddCommand(putCmd(r))
	r.cmd.AddCommand(descrCmd(r))
	r.cmd.AddCommand(versionCmd(r))
	r.cmd.AddCommand(deleteCmd(r))
	r.cmd.AddCommand(postCmd(r))
	r.cmd.AddCommand(postViewCleanupCmd(r))
	r.cmd.AddCommand(postFlushCmd(r))
	r.cmd.AddCommand(postCompactCmd(r))
	r.cmd.AddCommand(postCompactViewsCmd(r))
	r.cmd.AddCommand(postPurgeRootCmd(r))
	r.cmd.AddCommand(copyCmd(r))
	r.cmd.AddCommand(replicateCmd(r))

	return r
}

func parseDuration(val string) (time.Duration, error) {
	if val == "" {
		return 0, nil
	}
	if d, err := strconv.ParseFloat(val, 64); err == nil {
		if d < 0 {
			return 0, errors.Code(errors.ErrUsage, "negative timeout not permitted")
		}
		return time.Duration(d) * time.Second, nil
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, errors.Code(errors.ErrUsage, err)
	}
	if d < 0 {
		return 0, errors.Code(errors.ErrUsage, "negative timeout not permitted")
	}
	return d, nil
}

func (r *root) init(cmd *cobra.Command, args []string) error {
	r.log.SetOut(cmd.OutOrStdout())
	r.log.SetErr(cmd.ErrOrStderr())
	r.log.SetDebug(r.debug)

	r.log.Debug("Debug mode enabled")

	var err error
	r.parsedRequestTimeout, err = parseDuration(r.requestTimeout)
	if err != nil {
		return err
	}
	r.parsedConnectTimeout, err = parseDuration(r.connectTimeout)
	if err != nil {
		return err
	}
	r.retryDelayParsed, err = parseDuration(r.retryDelay)
	if err != nil {
		return err
	}
	r.retryTimeoutParsed, err = parseDuration(r.retryTimeout)
	if err != nil {
		return err
	}

	if err := r.conf.Read(r.resolveHome(r.confFile), r.log); err != nil {
		return err
	}

	if r.options == nil {
		r.options = map[string]interface{}{}
	}
	if len(args) > 0 {
		opts, err := r.conf.SetURL(args[0])
		if err != nil {
			return err
		}
		for k, v := range opts {
			if _, ok := r.options[k]; !ok {
				r.options[k] = v
			}
		}
	}
	for k, v := range r.stringOptions {
		if _, ok := r.options[k]; !ok {
			r.options[k] = v
		}
	}
	for k, v := range r.boolOptions {
		if _, ok := r.options[k]; !ok {
			switch strings.ToLower(v) {
			case "true", "t":
				r.options[k] = true
			case "false", "f":
				r.options[k] = false
			default:
				return errors.Codef(errors.ErrUsage, "invalid boolean value: %s", v)
			}
		}
	}

	if len(r.options) > 0 {
		r.log.Debug("CouchDB options: %v", r.options)
	}

	r.setTrace()

	return nil
}

func (r *root) client() (*kivik.Client, error) {
	cx, err := r.conf.CurrentCx()
	if err != nil {
		return nil, err
	}
	return cx.KivikClient(r.parsedConnectTimeout, r.parsedRequestTimeout)
}

func (r *root) RunE(cmd *cobra.Command, args []string) error {
	if _, err := r.client(); err != nil {
		return err
	}
	cx, err := r.conf.DSN()
	if err != nil {
		return err
	}
	r.log.Debugf("DSN: %s from %q", cx, r.conf.CurrentContext)

	return nil
}

func (r *root) retry(fn func() error) error {
	if r.retryCount == 0 {
		return fn()
	}
	var bo backoff.BackOff
	switch {
	case r.retryDelayParsed == 0 && r.retryDelay != "": // Disables retry delay
		bo = &backoff.ZeroBackOff{}
	case r.retryDelayParsed != 0:
		bo = backoff.NewConstantBackOff(r.retryDelayParsed)
	default:
		bo = backoff.NewExponentialBackOff()
	}
	if r.retryCount >= 0 {
		// WithMaxRetries really means WithMaxTries, so +1
		bo = backoff.WithMaxRetries(bo, uint64(r.retryCount+1))
	}
	if r.retryTimeoutParsed > 0 {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, r.retryTimeoutParsed)
		defer cancel()
		bo = backoff.WithContext(bo, ctx)
	}
	var count int
	var err error
	return backoff.Retry(func() error {
		if count > 0 {
			msg := fmt.Sprintf("Warning: Transient problem: %s.", err)
			switch nbo := bo.NextBackOff(); nbo {
			case backoff.Stop, 0:
			default:
				msg += fmt.Sprintf(" Will retry in %s.", fmtDuration(nbo))
			}
			if remain := r.retryCount - count; remain > 0 {
				msg += fmt.Sprintf(" %d retries left.", remain)
			}
			r.log.Info(msg)
		}
		count++
		err = fn()
		return err
	}, bo)
}

// nolint:gomnd
func fmtDuration(dur time.Duration) string {
	s := dur.Seconds()
	if s < 60 {
		return fmt.Sprintf("%0.2fs", s)
	}
	m := int(s / 60)
	s -= float64(m) * 60
	if m < 60 {
		return fmt.Sprintf("%dm%ds", m, int(s))
	}
	h := m / 60
	m -= h * 60
	if h < 24 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	d := h / 24
	h -= d * 24
	return fmt.Sprintf("%dd%dh%dm", d, h, m)
}

// opts returns the kivik options gathered from the command line.
func (r *root) opts() kivik.Option {
	return kivik.Params(r.options)
}
