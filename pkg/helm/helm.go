package helm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/strvals"
)

var settings *cli.EnvSettings

func Install(namespace, url, repoName, chartName, releaseName, version string, wait bool, args map[string]string) error {

	// Set the namespace
	os.Setenv("HELM_NAMESPACE", namespace)

	settings = cli.New()

	// No need to add/update if using OCI
	if !strings.HasPrefix(url, "oci://") {

		// Add helm repo
		if err := RepoAdd(repoName, url); err != nil {
			return err
		}

		// Update charts from the helm repo
		if err := RepoUpdate(); err != nil {
			return err
		}
	}

	// Install charts
	if err := InstallChart(releaseName, repoName, chartName, version, url, wait, args); err != nil {
		return err
	}

	// if we are here, everything is ok
	return nil
}

// RepoAdd adds repo with given name and url
func RepoAdd(name, url string) error {
	repoFile := settings.RepositoryConfig

	//Ensure the file directory exists as it is required for file locking
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Acquire a file lock for process synchronization
	fileLock := flock.New(strings.Replace(repoFile, filepath.Ext(repoFile), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		return err
	}

	b, err := os.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	if f.Has(name) {
		return nil
	}

	c := repo.Entry{
		Name: name,
		URL:  url,
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return err
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		err := errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", url)
		return err
	}

	f.Update(&c)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		return err
	}

	// if we are here, everything is ok
	return nil
}

// RepoUpdate updates charts for all helm repos
func RepoUpdate() error {
	repoFile := settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if os.IsNotExist(errors.Cause(err)) || len(f.Repositories) == 0 {
		return err
	}
	var repos []*repo.ChartRepository
	for _, cfg := range f.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			return err
		}
		repos = append(repos, r)
	}

	var wg sync.WaitGroup
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if _, err := re.DownloadIndexFile(); err != nil {
				log.Infof("...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
			}
		}(re)
	}
	wg.Wait()

	// if we are here, everything is ok
	return nil
}

// InstallChart
func InstallChart(name, repo, chart, version, url string, wait bool, args map[string]string) error {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), debug); err != nil {
		return err
	}
	client := action.NewInstall(actionConfig)

	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}

	// Set version if provided
	if version != "" {
		client.Version = version
	}

	client.ReleaseName = name

	// Get the chart path
	cp, err := getChartPath(url, repo, chart, client, settings)
	if err != nil {
		return err
	}

	p := getter.All(settings)
	valueOpts := &values.Options{}
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return err
	}

	// Add args
	if err := strvals.ParseInto(args["set"], vals); err != nil {
		return err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return err
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		return err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              os.Stdout,
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					RegistryClient:   client.GetRegistryClient(),
				}
				if err := man.Update(); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	// set and have helm create the namespace
	client.Namespace = settings.Namespace()
	client.CreateNamespace = true
	client.Wait = wait
	// TODO: Make this configurable
	client.Timeout = 180 * time.Second

	_, err = client.Run(chartRequested, vals)
	if err != nil {
		return err
	}

	// if we are here, everything is ok
	return nil
}

func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
}

// getChartPath returns the path to the chart taking OCI into account
func getChartPath(url, repo, chart string, client *action.Install, settings *cli.EnvSettings) (string, error) {
	if strings.HasPrefix(url, "oci://") {
		rc, err := registry.NewClient()
		if err != nil {
			return "", err
		}
		client.SetRegistryClient(rc)
		return client.ChartPathOptions.LocateChart(url, settings)
	} else {
		return client.ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", repo, chart), settings)
	}

}
