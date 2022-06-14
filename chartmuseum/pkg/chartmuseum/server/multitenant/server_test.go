package multitenant

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	pathutil "path"
	"strings"
	"testing"
	"time"

	cm_logger "github.com/kubernetes-helm/chartmuseum/pkg/chartmuseum/logger"
	cm_router "github.com/kubernetes-helm/chartmuseum/pkg/chartmuseum/router"
	"github.com/kubernetes-helm/chartmuseum/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/kubernetes-helm/chartmuseum/pkg/repo"
	"github.com/stretchr/testify/suite"
)

var maxUploadSize = 1024 * 1024 * 20

// These are generated from scripts/setup_test_environment.sh
var testTarballPath = "../../../../testdata/charts/mychart/mychart-0.1.0.tgz"
var testTarballPathV2 = "../../../../testdata/charts/mychart/mychart-0.2.0.tgz"
var testProvfilePath = "../../../../testdata/charts/mychart/mychart-0.1.0.tgz.prov"
var otherTestTarballPath = "../../../../testdata/charts/otherchart/otherchart-0.1.0.tgz"
var otherTestProvfilePath = "../../../../testdata/charts/otherchart/otherchart-0.1.0.tgz.prov"

type MultiTenantServerTestSuite struct {
	suite.Suite
	Depth0Server         *MultiTenantServer
	Depth1Server         *MultiTenantServer
	Depth2Server         *MultiTenantServer
	Depth3Server         *MultiTenantServer
	DisabledAPIServer    *MultiTenantServer
	OverwriteServer      *MultiTenantServer
	ChartURLServer       *MultiTenantServer
	MaxObjectsServer     *MultiTenantServer
	MaxUploadSizeServer  *MultiTenantServer
	TempDirectory        string
	TestTarballFilename  string
	TestProvfileFilename string
	StorageDirectory     map[string]map[string][]string
	LastCrashMessage     string
	LastPrinted          string
	LastExitCode         int
}

func (suite *MultiTenantServerTestSuite) doRequest(stype string, method string, urlStr string, body io.Reader, contentType string, output ...*bytes.Buffer) gin.ResponseWriter {
	recorder := httptest.NewRecorder()
	if len(output) > 0 {
		recorder.Body = output[0]
	}
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest(method, urlStr, body)
	if contentType != "" {
		c.Request.Header.Set("Content-Type", contentType)
	}

	switch stype {
	case "depth0":
		suite.Depth0Server.Router.HandleContext(c)
	case "depth1":
		suite.Depth1Server.Router.HandleContext(c)
	case "depth2":
		suite.Depth2Server.Router.HandleContext(c)
	case "depth3":
		suite.Depth3Server.Router.HandleContext(c)
	case "disabled":
		suite.DisabledAPIServer.Router.HandleContext(c)
	case "overwrite":
		suite.OverwriteServer.Router.HandleContext(c)
	case "charturl":
		suite.ChartURLServer.Router.HandleContext(c)
	case "maxobjects":
		suite.MaxObjectsServer.Router.HandleContext(c)
	case "maxuploadsize":
		suite.MaxUploadSizeServer.Router.HandleContext(c)
	}

	return c.Writer
}

func (suite *MultiTenantServerTestSuite) copyTestFilesTo(dir string) {
	srcFileTarball, err := os.Open(testTarballPath)
	suite.Nil(err, "no error opening test tarball")
	defer srcFileTarball.Close()

	destFileTarball, err := os.Create(pathutil.Join(dir, "mychart-0.1.0.tgz"))
	suite.Nil(err, fmt.Sprintf("no error creating new tarball in %s", dir))
	defer destFileTarball.Close()

	_, err = io.Copy(destFileTarball, srcFileTarball)
	suite.Nil(err, fmt.Sprintf("no error copying test testball to %s", dir))

	err = destFileTarball.Sync()
	suite.Nil(err, fmt.Sprintf("no error syncing temp tarball in %s", dir))

	srcFileProvfile, err := os.Open(testProvfilePath)
	suite.Nil(err, "no error opening test provenance file")
	defer srcFileProvfile.Close()

	destFileProvfile, err := os.Create(pathutil.Join(dir, "mychart-0.1.0.tgz.prov"))
	suite.Nil(err, fmt.Sprintf("no error creating new provenance file in %s", dir))
	defer destFileProvfile.Close()

	_, err = io.Copy(destFileProvfile, srcFileProvfile)
	suite.Nil(err, fmt.Sprintf("no error copying test provenance file to %s", dir))

	err = destFileProvfile.Sync()
	suite.Nil(err, fmt.Sprintf("no error syncing temp provenance file in %s", dir))
}

func (suite *MultiTenantServerTestSuite) populateOrgTeamRepoDirectory(org string, team string, repo string) {
	testPrefix := fmt.Sprintf("%s/%s/%s", org, team, repo)
	newDir := pathutil.Join(suite.TempDirectory, testPrefix)
	os.MkdirAll(newDir, os.ModePerm)
	suite.copyTestFilesTo(newDir)
	suite.copyTestFilesTo(pathutil.Join(newDir, ".."))
	suite.copyTestFilesTo(pathutil.Join(newDir, "../.."))
}

func (suite *MultiTenantServerTestSuite) SetupSuite() {
	echo = func(v ...interface{}) (int, error) {
		suite.LastPrinted = fmt.Sprint(v...)
		return 0, nil
	}

	exit = func(code int) {
		suite.LastExitCode = code
		suite.LastCrashMessage = fmt.Sprintf("exited %d", code)
	}

	timestamp := time.Now().Format("20060102150405")
	suite.TempDirectory = fmt.Sprintf("../../../../.test/chartmuseum-multitenant-server/%s", timestamp)
	os.MkdirAll(suite.TempDirectory, os.ModePerm)
	suite.copyTestFilesTo(suite.TempDirectory)

	srcFileTarball, err := os.Open(testTarballPath)
	suite.Nil(err, "no error opening test tarball")
	defer srcFileTarball.Close()

	suite.TestTarballFilename = pathutil.Join(suite.TempDirectory, "mychart-0.1.0.tgz")
	destFileTarball, err := os.Create(suite.TestTarballFilename)
	suite.Nil(err, "no error creating new tarball in temp dir")
	defer destFileTarball.Close()

	_, err = io.Copy(destFileTarball, srcFileTarball)
	suite.Nil(err, "no error copying test testball to temp tarball")

	err = destFileTarball.Sync()
	suite.Nil(err, "no error syncing temp tarball")

	suite.StorageDirectory = map[string]map[string][]string{
		"org1": {
			"team1": {"repo1", "repo2", "repo3"},
			"team2": {"repo1", "repo2", "repo3"},
			"team3": {"repo1", "repo2", "repo3"},
		},
		"org2": {
			"team1": {"repo1", "repo2", "repo3"},
			"team2": {"repo1", "repo2", "repo3"},
			"team3": {"repo1", "repo2", "repo3"},
		},
		"org3": {
			"team1": {"repo1", "repo2", "repo3"},
			"team2": {"repo1", "repo2", "repo3"},
			"team3": {"repo1", "repo2", "repo3"},
		},
	}

	// Scaffold out test storage directory structure
	for org, teams := range suite.StorageDirectory {
		for team, repos := range teams {
			for _, repo := range repos {
				suite.populateOrgTeamRepoDirectory(org, team, repo)
			}
		}
	}

	backend := storage.Backend(storage.NewLocalFilesystemBackend(suite.TempDirectory))

	logger, err := cm_logger.NewLogger(cm_logger.LoggerOptions{
		Debug: true,
	})
	suite.Nil(err, "no error creating logger")

	router := cm_router.NewRouter(cm_router.RouterOptions{
		Logger:        logger,
		Depth:         0,
		EnableMetrics: true,
		MaxUploadSize: maxUploadSize,
	})
	server, err := NewMultiTenantServer(MultiTenantServerOptions{
		Logger:                 logger,
		Router:                 router,
		StorageBackend:         backend,
		EnableAPI:              true,
		ChartPostFormFieldName: "chart",
		ProvPostFormFieldName:  "prov",
		IndexLimit:             1,
	})
	suite.NotNil(server)
	suite.Nil(err, "no error creating new multitenant (depth=0) server")
	suite.Depth0Server = server

	router = cm_router.NewRouter(cm_router.RouterOptions{
		Logger:        logger,
		Depth:         1,
		EnableMetrics: true,
		MaxUploadSize: maxUploadSize,
	})
	server, err = NewMultiTenantServer(MultiTenantServerOptions{
		Logger:                 logger,
		Router:                 router,
		StorageBackend:         backend,
		EnableAPI:              true,
		ChartPostFormFieldName: "chart",
		ProvPostFormFieldName:  "prov",
	})
	suite.NotNil(server)
	suite.Nil(err, "no error creating new multitenant (depth=1) server")
	suite.Depth1Server = server

	router = cm_router.NewRouter(cm_router.RouterOptions{
		Logger:        logger,
		Depth:         2,
		MaxUploadSize: maxUploadSize,
	})
	server, err = NewMultiTenantServer(MultiTenantServerOptions{
		Logger:                 logger,
		Router:                 router,
		StorageBackend:         backend,
		EnableAPI:              true,
		ChartPostFormFieldName: "chart",
		ProvPostFormFieldName:  "prov",
	})
	suite.NotNil(server)
	suite.Nil(err, "no error creating new multitenant (depth=2) server")
	suite.Depth2Server = server

	router = cm_router.NewRouter(cm_router.RouterOptions{
		Logger:        logger,
		Depth:         3,
		MaxUploadSize: maxUploadSize,
	})
	server, err = NewMultiTenantServer(MultiTenantServerOptions{
		Logger:                 logger,
		Router:                 router,
		StorageBackend:         backend,
		EnableAPI:              true,
		ChartPostFormFieldName: "chart",
		ProvPostFormFieldName:  "prov",
	})
	suite.NotNil(server)
	suite.Nil(err, "no error creating new multitenant (depth=3) server")
	suite.Depth3Server = server

	router = cm_router.NewRouter(cm_router.RouterOptions{
		Logger:        logger,
		Depth:         0,
		MaxUploadSize: maxUploadSize,
	})

	server, err = NewMultiTenantServer(MultiTenantServerOptions{
		Logger:         logger,
		Router:         router,
		StorageBackend: backend,
		EnableAPI:      false,
	})
	suite.NotNil(server)
	suite.Nil(err, "no error creating new disabled server")
	suite.DisabledAPIServer = server

	router = cm_router.NewRouter(cm_router.RouterOptions{
		Logger:        logger,
		Depth:         0,
		MaxUploadSize: maxUploadSize,
	})
	server, err = NewMultiTenantServer(MultiTenantServerOptions{
		Logger:                 logger,
		Router:                 router,
		StorageBackend:         backend,
		EnableAPI:              true,
		AllowOverwrite:         true,
		ChartPostFormFieldName: "chart",
		ProvPostFormFieldName:  "prov",
	})
	suite.NotNil(server)
	suite.Nil(err, "no error creating new overwrite server")
	suite.OverwriteServer = server

	router = cm_router.NewRouter(cm_router.RouterOptions{
		Logger:        logger,
		Depth:         0,
		MaxUploadSize: maxUploadSize,
	})
	server, err = NewMultiTenantServer(MultiTenantServerOptions{
		Logger:                 logger,
		Router:                 router,
		StorageBackend:         backend,
		EnableAPI:              true,
		ChartPostFormFieldName: "chart",
		ProvPostFormFieldName:  "prov",
		ChartURL:               "https://chartmuseum.com",
	})
	suite.NotNil(server)
	suite.Nil(err, "no error creating new custom chart URL server")
	suite.ChartURLServer = server

	router = cm_router.NewRouter(cm_router.RouterOptions{
		Logger:        logger,
		Depth:         0,
		MaxUploadSize: maxUploadSize,
	})
	server, err = NewMultiTenantServer(MultiTenantServerOptions{
		Logger:                 logger,
		Router:                 router,
		StorageBackend:         backend,
		EnableAPI:              true,
		AllowOverwrite:         true,
		ChartPostFormFieldName: "chart",
		ProvPostFormFieldName:  "prov",
		MaxStorageObjects:      1,
	})
	suite.NotNil(server)
	suite.Nil(err, "no error creating new max objects server")
	suite.MaxObjectsServer = server

	router = cm_router.NewRouter(cm_router.RouterOptions{
		Logger:        logger,
		Depth:         0,
		MaxUploadSize: 1, // intentionally small
	})
	server, err = NewMultiTenantServer(MultiTenantServerOptions{
		Logger:                 logger,
		Router:                 router,
		StorageBackend:         backend,
		EnableAPI:              true,
		AllowOverwrite:         true,
		ChartPostFormFieldName: "chart",
		ProvPostFormFieldName:  "prov",
	})
	suite.NotNil(server)
	suite.Nil(err, "no error creating new max upload size server")
	suite.MaxUploadSizeServer = server
}

func (suite *MultiTenantServerTestSuite) TearDownSuite() {
	err := os.RemoveAll(suite.TempDirectory)
	suite.Nil(err, "no error deleting temp directory for local storage")
}

func (suite *MultiTenantServerTestSuite) TestRegenerateRepositoryIndex() {
	server := suite.Depth0Server

	log := server.Logger.ContextLoggingFn(&gin.Context{})

	entry, err := server.initCacheEntry(log, "")
	suite.Nil(err, "no error on init cache entry")

	objects, err := server.fetchChartsInStorage(log, "")
	diff := storage.GetObjectSliceDiff(server.getRepoObjectSlice(entry), objects)
	_, err = server.regenerateRepositoryIndexWorker(log, entry, diff)
	suite.Nil(err, "no error regenerating repo index")

	newtime := time.Now().Add(1 * time.Hour)
	err = os.Chtimes(suite.TestTarballFilename, newtime, newtime)
	suite.Nil(err, "no error changing modtime on temp file")

	objects, err = server.fetchChartsInStorage(log, "")
	diff = storage.GetObjectSliceDiff(server.getRepoObjectSlice(entry), objects)
	_, err = server.regenerateRepositoryIndexWorker(log, entry, diff)
	suite.Nil(err, "no error regenerating repo index with tarball updated")

	brokenTarballFilename := pathutil.Join(suite.TempDirectory, "brokenchart.tgz")
	destFile, err := os.Create(brokenTarballFilename)
	suite.Nil(err, "no error creating new broken tarball in temp dir")
	defer destFile.Close()
	objects, err = server.fetchChartsInStorage(log, "")
	diff = storage.GetObjectSliceDiff(server.getRepoObjectSlice(entry), objects)
	_, err = server.regenerateRepositoryIndexWorker(log, entry, diff)
	suite.Nil(err, "error not returned with broken tarball added")

	err = os.Chtimes(brokenTarballFilename, newtime, newtime)
	suite.Nil(err, "no error changing modtime on broken tarball")
	objects, err = server.fetchChartsInStorage(log, "")
	diff = storage.GetObjectSliceDiff(server.getRepoObjectSlice(entry), objects)
	_, err = server.regenerateRepositoryIndexWorker(log, entry, diff)
	suite.Nil(err, "error not returned with broken tarball updated")

	err = os.Remove(brokenTarballFilename)
	suite.Nil(err, "no error removing broken tarball")
	objects, err = server.fetchChartsInStorage(log, "")
	diff = storage.GetObjectSliceDiff(server.getRepoObjectSlice(entry), objects)
	_, err = server.regenerateRepositoryIndexWorker(log, entry, diff)
	suite.Nil(err, "error not returned with broken tarball removed")
}

func (suite *MultiTenantServerTestSuite) TestGenIndex() {
	logger, err := cm_logger.NewLogger(cm_logger.LoggerOptions{
		Debug:   true,
		LogJSON: true,
	})
	suite.Nil(err, "no error creating logger")

	router := cm_router.NewRouter(cm_router.RouterOptions{
		Logger: logger,
	})

	NewMultiTenantServer(MultiTenantServerOptions{
		Logger:         logger,
		Router:         router,
		StorageBackend: suite.Depth0Server.StorageBackend,
		GenIndex:       true,
	})
	suite.Equal("exited 0", suite.LastCrashMessage, "no error with --gen-index")
	suite.Equal(0, suite.LastExitCode, "--gen-index flag exits 0")
	suite.Contains(suite.LastPrinted, "apiVersion:", "--gen-index prints yaml")
}

func (suite *MultiTenantServerTestSuite) TestStatefiles() {
	logger, err := cm_logger.NewLogger(cm_logger.LoggerOptions{
		Debug:   true,
		LogJSON: true,
	})
	suite.Nil(err, "no error creating logger")

	router := cm_router.NewRouter(cm_router.RouterOptions{
		Logger: logger,
	})

	// add an example index-cache.yaml
	indexCacheFilePath := pathutil.Join(suite.TempDirectory, repo.StatefileFilename)
	content := []byte(`apiVersion: v1
entries:
  acs-engine-autoscaler:
  - name: acs-engine-autoscaler
    urls:
    - charts/acs-engine-autoscaler-2.1.2.tgz
    version: 2.1.2
generated: "2018-05-23T15:14:46-05:00"`)
	err = ioutil.WriteFile(indexCacheFilePath, content, 0644)
	suite.Nil(err, "no error creating test index-cache.yaml")
	defer os.Remove(indexCacheFilePath)

	NewMultiTenantServer(MultiTenantServerOptions{
		Logger:         logger,
		Router:         router,
		StorageBackend: suite.Depth0Server.StorageBackend,
		UseStatefiles:  true,
		GenIndex:       true,
	})
	suite.Equal("exited 0", suite.LastCrashMessage, "no error with --gen-index")
	suite.Equal(0, suite.LastExitCode, "--gen-index flag exits 0")
	suite.Contains(suite.LastPrinted, "apiVersion:", "--gen-index prints yaml")

	// remove index-cache.yaml
	err = os.Remove(indexCacheFilePath)
	suite.Nil(err, "no error deleting test index-cache.yaml")

	// invalid, unparsable index-cache.yaml
	indexCacheFilePath = pathutil.Join(suite.TempDirectory, repo.StatefileFilename)
	content = []byte(`is this valid yaml? maybe. but its definitely not a valid index.yaml!`)
	err = ioutil.WriteFile(indexCacheFilePath, content, 0644)
	suite.Nil(err, "no error creating test index-cache.yaml")

	NewMultiTenantServer(MultiTenantServerOptions{
		Logger:         logger,
		Router:         router,
		StorageBackend: suite.Depth0Server.StorageBackend,
		UseStatefiles:  true,
		GenIndex:       true,
	})
	suite.Equal("exited 0", suite.LastCrashMessage, "no error with --gen-index")
	suite.Equal(0, suite.LastExitCode, "--gen-index flag exits 0")
	suite.Contains(suite.LastPrinted, "apiVersion:", "--gen-index prints yaml")

	// remove index-cache.yaml
	err = os.Remove(indexCacheFilePath)
	suite.Nil(err, "no error deleting test index-cache.yaml")

	// no index-cache.yaml
	NewMultiTenantServer(MultiTenantServerOptions{
		Logger:         logger,
		Router:         router,
		StorageBackend: suite.Depth0Server.StorageBackend,
		UseStatefiles:  true,
		GenIndex:       true,
	})
	suite.Equal("exited 0", suite.LastCrashMessage, "no error with --gen-index")
	suite.Equal(0, suite.LastExitCode, "--gen-index flag exits 0")
	suite.Contains(suite.LastPrinted, "apiVersion:", "--gen-index prints yaml")
}

func (suite *MultiTenantServerTestSuite) TestDisabledServer() {
	// Test that all /api routes disabled if EnableAPI=false
	res := suite.doRequest("disabled", "GET", "/api/charts", nil, "")
	suite.Equal(404, res.Status(), "404 GET /api/charts")

	res = suite.doRequest("disabled", "GET", "/api/charts/mychart", nil, "")
	suite.Equal(404, res.Status(), "404 GET /api/charts")

	res = suite.doRequest("disabled", "GET", "/api/charts/mychart/0.1.0", nil, "")
	suite.Equal(404, res.Status(), "404 GET /api/charts")

	body := bytes.NewBuffer([]byte{})
	res = suite.doRequest("disabled", "POST", "/api/charts", body, "")
	suite.Equal(404, res.Status(), "404 POST /api/charts")

	body = bytes.NewBuffer([]byte{})
	res = suite.doRequest("disabled", "POST", "/api/prov", body, "")
	suite.Equal(404, res.Status(), "404 POST /api/prov")

	res = suite.doRequest("disabled", "DELETE", "/api/charts/mychart/0.1.0", nil, "")
	suite.Equal(404, res.Status(), "404 DELETE /api/charts/mychart/0.1.0")
}

func (suite *MultiTenantServerTestSuite) TestOverwriteServer() {
	// Check if files can be overwritten
	content, err := ioutil.ReadFile(testTarballPath)
	suite.Nil(err, "no error opening test tarball")
	body := bytes.NewBuffer(content)
	res := suite.doRequest("overwrite", "POST", "/api/charts", body, "")
	suite.Equal(201, res.Status(), "201 POST /api/charts")
	body = bytes.NewBuffer(content)
	res = suite.doRequest("overwrite", "POST", "/api/charts", body, "")
	suite.Equal(201, res.Status(), "201 POST /api/charts")

	content, err = ioutil.ReadFile(testProvfilePath)
	suite.Nil(err, "no error opening test provenance file")
	body = bytes.NewBuffer(content)
	res = suite.doRequest("overwrite", "POST", "/api/prov", body, "")
	suite.Equal(201, res.Status(), "201 POST /api/prov")
	body = bytes.NewBuffer(content)
	res = suite.doRequest("overwrite", "POST", "/api/prov", body, "")
	suite.Equal(201, res.Status(), "201 POST /api/prov")

	buf, w := suite.getBodyWithMultipartFormFiles([]string{"chart", "prov"}, []string{testTarballPath, testProvfilePath})
	res = suite.doRequest("overwrite", "POST", "/api/charts", buf, w.FormDataContentType())
	suite.Equal(201, res.Status(), "201 POST /api/charts")
	buf, w = suite.getBodyWithMultipartFormFiles([]string{"chart", "prov"}, []string{testTarballPath, testProvfilePath})
	res = suite.doRequest("overwrite", "POST", "/api/charts", buf, w.FormDataContentType())
	suite.Equal(201, res.Status(), "201 POST /api/charts")
}

func (suite *MultiTenantServerTestSuite) TestCustomChartURLServer() {
	res := suite.doRequest("charturl", "GET", "/index.yaml", nil, "")
	suite.Equal(200, res.Status(), "200 GET /index.yaml")
}

func (suite *MultiTenantServerTestSuite) TestMaxObjectsServer() {
	// Overwrites should still be allowed if limit is reached
	content, err := ioutil.ReadFile(testTarballPath)
	suite.Nil(err, "no error opening test tarball")
	body := bytes.NewBuffer(content)
	res := suite.doRequest("maxobjects", "POST", "/api/charts", body, "")
	suite.Equal(201, res.Status(), "201 POST /api/charts")

	content, err = ioutil.ReadFile(testProvfilePath)
	suite.Nil(err, "no error opening test provenance file")
	body = bytes.NewBuffer(content)
	res = suite.doRequest("maxobjects", "POST", "/api/prov", body, "")
	suite.Equal(201, res.Status(), "201 POST /api/prov")

	// trigger error, reached max
	content, err = ioutil.ReadFile(otherTestTarballPath)
	suite.Nil(err, "no error opening other test tarball")
	body = bytes.NewBuffer(content)
	res = suite.doRequest("maxobjects", "POST", "/api/charts", body, "")
	suite.Equal(507, res.Status(), "507 POST /api/charts")

	content, err = ioutil.ReadFile(otherTestProvfilePath)
	suite.Nil(err, "no error opening other test provenance file")
	body = bytes.NewBuffer(content)
	res = suite.doRequest("maxobjects", "POST", "/api/prov", body, "")
	suite.Equal(507, res.Status(), "507 POST /api/prov")
}

func (suite *MultiTenantServerTestSuite) TestMaxUploadSizeServer() {
	// trigger 413s, "request too large"
	content, err := ioutil.ReadFile(testTarballPath)
	suite.Nil(err, "no error opening test tarball")
	body := bytes.NewBuffer(content)
	res := suite.doRequest("maxuploadsize", "POST", "/api/charts", body, "")
	suite.Equal(413, res.Status(), "413 POST /api/charts")

	content, err = ioutil.ReadFile(testProvfilePath)
	suite.Nil(err, "no error opening test provenance file")
	body = bytes.NewBuffer(content)
	res = suite.doRequest("maxuploadsize", "POST", "/api/prov", body, "")
	suite.Equal(413, res.Status(), "201 POST /api/prov")

	buf, w := suite.getBodyWithMultipartFormFiles([]string{"chart", "prov"}, []string{testTarballPath, testProvfilePath})
	res = suite.doRequest("maxuploadsize", "POST", "/api/charts", buf, w.FormDataContentType())
	suite.Equal(413, res.Status(), "413 POST /api/charts")
}

func (suite *MultiTenantServerTestSuite) TestMetrics() {

	apiPrefix := pathutil.Join("/api", "a")

	content, err := ioutil.ReadFile(testTarballPath)
	suite.Nil(err, "error opening test tarball")

	body := bytes.NewBuffer(content)
	res := suite.doRequest("depth1", "POST", fmt.Sprintf("%s/charts", apiPrefix), body, "")
	suite.Equal(201, res.Status(), fmt.Sprintf("201 post %s/charts", apiPrefix))

	otherChart, err := ioutil.ReadFile(testTarballPathV2)
	suite.Nil(err, "error opening test tarball")

	body = bytes.NewBuffer(otherChart)
	res = suite.doRequest("depth1", "POST", fmt.Sprintf("%s/charts", apiPrefix), body, "")
	suite.Equal(201, res.Status(), fmt.Sprintf("201 POST %s/charts", apiPrefix))

	// GET /a/index.yaml to regenerate index (and metrics)
	res = suite.doRequest("depth1", "GET", "/a/index.yaml", nil, "")
	suite.Equal(200, res.Status(), "200 GET /a/index.yaml")

	// GET /b/index.yaml to regenerate b index (and metrics)
	res = suite.doRequest("depth1", "GET", "/b/index.yaml", nil, "")
	suite.Equal(200, res.Status(), "200 GET /b/index.yaml")

	// Get metrics
	buffer := bytes.NewBufferString("")
	res = suite.doRequest("depth1", "GET", "/metrics", nil, "", buffer)
	suite.Equal(200, res.Status(), "200 GET /metrics")

	metrics := buffer.String()
	//fmt.Print(metrics) // observe the metric output

	// Ensure that we have the Gauges as documented
	suite.True(strings.Contains(metrics, "# TYPE chartmuseum_chart_versions_served_total gauge"))
	suite.True(strings.Contains(metrics, "# TYPE chartmuseum_charts_served_total gauge"))

	suite.True(strings.Contains(metrics, "chartmuseum_charts_served_total{repo=\"a\"} 1"))
	suite.True(strings.Contains(metrics, "chartmuseum_chart_versions_served_total{repo=\"a\"} 2"))

	// Ensure that the b repo has no charts
	suite.True(strings.Contains(metrics, "chartmuseum_chart_versions_served_total{repo=\"b\"} 0"))
}

func (suite *MultiTenantServerTestSuite) TestRoutes() {
	suite.testAllRoutes("", 0)
	for org, teams := range suite.StorageDirectory {
		suite.testAllRoutes(org, 1)
		for team, repos := range teams {
			suite.testAllRoutes(pathutil.Join(org, team), 2)
			for _, repo := range repos {
				suite.testAllRoutes(pathutil.Join(org, team, repo), 3)
			}
		}
	}
}

func (suite *MultiTenantServerTestSuite) testAllRoutes(repo string, depth int) {
	var res gin.ResponseWriter

	stype := fmt.Sprintf("depth%d", depth)

	// GET /
	res = suite.doRequest(stype, "GET", "/", nil, "")
	suite.Equal(200, res.Status(), "200 GET /")

	// GET /health
	res = suite.doRequest(stype, "GET", "/health", nil, "")
	suite.Equal(200, res.Status(), "200 GET /health")

	var repoPrefix string
	if repo != "" {
		repoPrefix = pathutil.Join("/", repo)
	} else {
		repoPrefix = ""
	}

	// GET /:repo/index.yaml
	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/index.yaml", repoPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 GET %s/index.yaml", repoPrefix))

	// Issue #21
	suite.NotEqual("", res.Header().Get("X-Request-Id"), "X-Request-Id header is present")
	suite.Equal("", res.Header().Get("X-Blah-Blah-Blah"), "X-Blah-Blah-Blah header is not present")

	// GET /:repo/charts/:filename
	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/mychart-0.1.0.tgz", repoPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 GET %s/charts/mychart-0.1.0.tgz", repoPrefix))

	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/mychart-0.1.0.tgz.prov", repoPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 GET %s/charts/mychart-0.1.0.tgz.prov", repoPrefix))

	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/fakechart-0.1.0.tgz", repoPrefix), nil, "")
	suite.Equal(404, res.Status(), fmt.Sprintf("404 GET %s/charts/fakechart-0.1.0.tgz", repoPrefix))

	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/fakechart-0.1.0.tgz.prov", repoPrefix), nil, "")
	suite.Equal(404, res.Status(), fmt.Sprintf("404 GET %s/charts/fakechart-0.1.0.tgz.prov", repoPrefix))

	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/fakechart-0.1.0.bad", repoPrefix), nil, "")
	suite.Equal(500, res.Status(), fmt.Sprintf("500 GET %s/charts/fakechart-0.1.0.bad", repoPrefix))

	apiPrefix := pathutil.Join("/api", repo)

	// GET /api/:repo/charts
	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts", apiPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 GET %s/charts", apiPrefix))

	// GET /api/:repo/charts/:name
	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/mychart", apiPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 GET %s/charts/mychart", apiPrefix))

	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/fakechart", apiPrefix), nil, "")
	suite.Equal(404, res.Status(), fmt.Sprintf("404 GET %s/charts/fakechart", apiPrefix))

	// GET /api/:repo/charts/:name/:version
	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/mychart/0.1.0", apiPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 GET %s/charts/mychart/0.1.0", apiPrefix))

	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/mychart/latest", apiPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 GET %s/charts/mychart/latest", apiPrefix))

	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/mychart/0.1.1", apiPrefix), nil, "")
	suite.Equal(404, res.Status(), fmt.Sprintf("200 GET %s/charts/mychart/0.1.1", apiPrefix))

	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/charts/fakechart/0.1.0", apiPrefix), nil, "")
	suite.Equal(404, res.Status(), fmt.Sprintf("200 GET %s/charts/fakechart/0.1.0", apiPrefix))

	// DELETE /api/:repo/charts/:name/:version
	res = suite.doRequest(stype, "DELETE", fmt.Sprintf("%s/charts/mychart/0.1.0", apiPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 DELETE %s/charts/mychart/0.1.0", apiPrefix))

	res = suite.doRequest(stype, "DELETE", fmt.Sprintf("%s/charts/mychart/0.1.0", apiPrefix), nil, "")
	suite.Equal(404, res.Status(), fmt.Sprintf("200 DELETE %s/charts/mychart/0.1.0", apiPrefix))

	// GET /:repo/index.yaml (after delete)
	res = suite.doRequest(stype, "GET", fmt.Sprintf("%s/index.yaml", repoPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 GET %s/index.yaml", repoPrefix))

	// POST /api/:repo/charts
	body := bytes.NewBuffer([]byte{})
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), body, "")
	suite.Equal(500, res.Status(), fmt.Sprintf("500 POST %s/charts", apiPrefix))

	// POST /api/:repo/prov
	body = bytes.NewBuffer([]byte{})
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/prov", apiPrefix), body, "")
	suite.Equal(500, res.Status(), fmt.Sprintf("500 POST %s/prov", apiPrefix))

	// POST /api/:repo/charts
	content, err := ioutil.ReadFile(testTarballPath)
	suite.Nil(err, "no error opening test tarball")

	body = bytes.NewBuffer(content)
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), body, "")
	suite.Equal(201, res.Status(), fmt.Sprintf("201 POST %s/charts", apiPrefix))

	body = bytes.NewBuffer(content)
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), body, "")
	suite.Equal(409, res.Status(), fmt.Sprintf("409 POST %s/charts", apiPrefix))

	// POST /api/:repo/prov
	content, err = ioutil.ReadFile(testProvfilePath)
	suite.Nil(err, "no error opening test provenance file")

	body = bytes.NewBuffer(content)
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/prov", apiPrefix), body, "")
	suite.Equal(201, res.Status(), fmt.Sprintf("201 POST %s/prov", apiPrefix))

	body = bytes.NewBuffer(content)
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/prov", apiPrefix), body, "")
	suite.Equal(409, res.Status(), fmt.Sprintf("409 POST %s/prov", apiPrefix))

	// Clear test repo to allow uploading again
	res = suite.doRequest(stype, "DELETE", fmt.Sprintf("%s/charts/mychart/0.1.0", apiPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 DELETE %s/charts/mychart/0.1.0", apiPrefix))

	// Create form file with chart=@mychart-0.1.0.tgz
	buf, w := suite.getBodyWithMultipartFormFiles([]string{"chart"}, []string{testTarballPath})
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), buf, w.FormDataContentType())
	suite.Equal(201, res.Status(), fmt.Sprintf("201 POST %s/charts", apiPrefix))

	// Create form file with prov=@mychart-0.1.0.tgz.prov
	buf, w = suite.getBodyWithMultipartFormFiles([]string{"prov"}, []string{testProvfilePath})
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), buf, w.FormDataContentType())
	suite.Equal(201, res.Status(), fmt.Sprintf("201 POST %s/charts", apiPrefix))

	// Clear test repo to allow uploading again
	res = suite.doRequest(stype, "DELETE", fmt.Sprintf("%s/charts/mychart/0.1.0", apiPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 DELETE %s/charts/mychart/0.1.0", apiPrefix))

	// Create form file with chart=@mychart-0.1.0.tgz and prov=@mychart-0.1.0.tgz.prov
	buf, w = suite.getBodyWithMultipartFormFiles([]string{"chart", "prov"}, []string{testTarballPath, testProvfilePath})
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), buf, w.FormDataContentType())
	suite.Equal(201, res.Status(), fmt.Sprintf("201 POST %s/charts", apiPrefix))

	// Clear test repo to allow uploading again
	res = suite.doRequest(stype, "DELETE", fmt.Sprintf("%s/charts/mychart/0.1.0", apiPrefix), nil, "")
	suite.Equal(200, res.Status(), fmt.Sprintf("200 DELETE %s/charts/mychart/0.1.0", apiPrefix))

	// Create form file with unknown=@mychart-0.1.0.tgz, which should fail because the server doesn't know about the unknown field
	buf, w = suite.getBodyWithMultipartFormFiles([]string{"unknown"}, []string{testTarballPath})
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), buf, w.FormDataContentType())
	suite.Equal(400, res.Status(), fmt.Sprintf("400 POST %s/charts", apiPrefix))

	// Create form file with chart=@mychart-0.1.0.tgz
	buf, w = suite.getBodyWithMultipartFormFiles([]string{"chart"}, []string{testTarballPath})
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), buf, w.FormDataContentType())
	suite.Equal(201, res.Status(), fmt.Sprintf("201 POST %s/charts", apiPrefix))

	// Create form file with chart=@mychart-0.1.0.tgz, which should fail because it is already there
	buf, w = suite.getBodyWithMultipartFormFiles([]string{"chart"}, []string{testTarballPath})
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), buf, w.FormDataContentType())
	suite.Equal(409, res.Status(), fmt.Sprintf("409 POST %s/charts", apiPrefix))

	// Create form file with chart=@mychart-0.1.0.tgz.prov, which should fail because it is not a valid chart package
	buf, w = suite.getBodyWithMultipartFormFiles([]string{"chart"}, []string{testProvfilePath})
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), buf, w.FormDataContentType())
	suite.Equal(400, res.Status(), fmt.Sprintf("400 POST %s/charts", apiPrefix))

	// Create form file with prov=@mychart-0.1.0.tgz, which should fail because it is not a valid provenance file
	buf, w = suite.getBodyWithMultipartFormFiles([]string{"prov"}, []string{testTarballPath})
	res = suite.doRequest(stype, "POST", fmt.Sprintf("%s/charts", apiPrefix), buf, w.FormDataContentType())
	suite.Equal(400, res.Status(), fmt.Sprintf("400 POST %s/charts", apiPrefix))

}

func (suite *MultiTenantServerTestSuite) getBodyWithMultipartFormFiles(fields []string, filenames []string) (io.Reader, *multipart.Writer) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	for i := range fields {
		fw, err := w.CreateFormFile(fields[i], filenames[i])
		suite.Nil(err, "no error creating form file")
		fd, err := os.Open(filenames[i])
		suite.Nil(err, "no error opening test file")
		defer fd.Close()
		_, err = io.Copy(fw, fd)
		suite.Nil(err, "no error copying test file to form file")
	}
	w.Close()
	return buf, w
}

func TestMultiTenantServerTestSuite(t *testing.T) {
	suite.Run(t, new(MultiTenantServerTestSuite))
}
